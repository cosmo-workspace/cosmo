package dashboard

import (
	"context"
	"fmt"
	"reflect"

	connect_go "github.com/bufbuild/connect-go"
	"k8s.io/apimachinery/pkg/types"

	apierrs "k8s.io/apimachinery/pkg/api/errors"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/useraddon"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

func (s *Server) UpdateUserAddons(ctx context.Context, req *connect_go.Request[dashv1alpha1.UpdateUserAddonsRequest]) (*connect_go.Response[dashv1alpha1.UpdateUserAddonsResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	currentUser, err := s.Klient.GetUser(ctx, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	// caller can attach or detach only:
	//   - User who have group-role which caller is admin for
	//   - Addons which is allowed for caller to manage
	err = adminAuthentication(ctx,
		validateCallerHasAdminForAtLeastOneRole(currentUser.Spec.Roles))
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	caller := callerFromContext(ctx)
	if caller == nil {
		return nil, apierrs.NewInternalError(fmt.Errorf("unable get caller"))
	}
	for _, addon := range diff(currentUser.Spec.Addons, apiconv.D2C_UserAddons(req.Msg.Addons)) {
		tmpl := useraddon.EmptyTemplateObject(addon)
		err := s.Klient.Get(ctx, types.NamespacedName{Name: tmpl.GetName()}, tmpl)
		if err != nil {
			return nil, apierrs.NewInternalError(fmt.Errorf("failed to fetch addon '%s'", tmpl.GetName()))
		}
		if ok := kosmo.IsAllowedToUseTemplate(ctx, caller, tmpl); !ok {
			roles := kubeutil.GetAnnotation(tmpl, cosmov1alpha1.TemplateAnnKeyUserRoles)
			return nil, NewForbidden(fmt.Errorf("roles '%s' is required for addon '%s'", roles, tmpl.GetName()))
		}
	}

	addons := apiconv.D2C_UserAddons(req.Msg.Addons)
	user, err := s.Klient.UpdateUser(ctx, req.Msg.UserName, kosmo.UpdateUserOpts{
		UserAddons: addons,
	})
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.UpdateUserAddonsResponse{
		Message: "Successfully updated",
		User:    apiconv.C2D_User(*user),
	}
	log.Info(res.Message, "username", req.Msg.UserName)
	return connect_go.NewResponse(res), nil
}

func (s *Server) UpdateUserDisplayName(ctx context.Context, req *connect_go.Request[dashv1alpha1.UpdateUserDisplayNameRequest]) (*connect_go.Response[dashv1alpha1.UpdateUserDisplayNameResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	if err := userAuthentication(ctx, req.Msg.UserName); err != nil {
		return nil, ErrResponse(log, err)
	}

	user, err := s.Klient.UpdateUser(ctx, req.Msg.UserName, kosmo.UpdateUserOpts{DisplayName: &req.Msg.DisplayName})
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.UpdateUserDisplayNameResponse{
		Message: "Successfully updated",
		User:    apiconv.C2D_User(*user),
	}
	log.Info(res.Message, "username", req.Msg.UserName)
	return connect_go.NewResponse(res), nil
}

func diff[T any](slice1 []T, slice2 []T) []T {
	var diff []T
	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if reflect.DeepEqual(s1, s2) {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices
		slice1, slice2 = slice2, slice1
	}
	return diff
}

func (s *Server) UpdateUserRole(ctx context.Context, req *connect_go.Request[dashv1alpha1.UpdateUserRoleRequest]) (*connect_go.Response[dashv1alpha1.UpdateUserRoleResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	currentUser, err := s.Klient.GetUser(ctx, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	// group-admin can attach or detach only group-roles
	changingRoles := diff(apiconv.C2S_UserRole(currentUser.Spec.Roles), req.Msg.Roles)
	if err := adminAuthentication(ctx, validateCallerHasAdminForAllRoles(changingRoles)); err != nil {
		return nil, ErrResponse(log, err)
	}

	roles := apiconv.S2C_UserRoles(req.Msg.Roles)
	user, err := s.Klient.UpdateUser(ctx, req.Msg.UserName, kosmo.UpdateUserOpts{UserRoles: roles})
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.UpdateUserRoleResponse{
		Message: "Successfully updated",
		User:    apiconv.C2D_User(*user),
	}
	log.Info(res.Message, "username", req.Msg.UserName)
	return connect_go.NewResponse(res), nil
}

func (s *Server) UpdateUserPassword(ctx context.Context, req *connect_go.Request[dashv1alpha1.UpdateUserPasswordRequest]) (*connect_go.Response[dashv1alpha1.UpdateUserPasswordResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "username", req.Msg.UserName)

	if err := userAuthentication(ctx, req.Msg.UserName); err != nil {
		return nil, ErrResponse(log, err)
	}

	// check current password is valid
	verified, _, err := s.Klient.VerifyPassword(ctx, req.Msg.UserName, []byte(req.Msg.CurrentPassword))
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	if !verified {
		return nil, ErrResponse(log, NewForbidden(fmt.Errorf("incorrect user or password")))
	}

	// Upsert password
	if err := s.Klient.RegisterPassword(ctx, req.Msg.UserName, []byte(req.Msg.NewPassword)); err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.UpdateUserPasswordResponse{
		Message: "Successfully updated",
	}
	log.Info(res.Message, "username", req.Msg.UserName)
	return connect_go.NewResponse(res), nil
}
