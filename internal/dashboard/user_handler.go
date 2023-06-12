package dashboard

import (
	"context"
	"fmt"
	"net/http"

	connect_go "github.com/bufbuild/connect-go"
	"google.golang.org/protobuf/types/known/emptypb"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

func (s *Server) UserServiceHandler(mux *http.ServeMux) {
	path, handler := dashboardv1alpha1connect.NewUserServiceHandler(s,
		connect_go.WithInterceptors(s.authorizationInterceptor()),
		connect_go.WithInterceptors(s.validatorInterceptor()),
	)
	mux.Handle(path, s.contextMiddleware(handler))
}

func (s *Server) CreateUser(ctx context.Context, req *connect_go.Request[dashv1alpha1.CreateUserRequest]) (*connect_go.Response[dashv1alpha1.CreateUserResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	// group-admin user can create users which have only the their groups
	if err := adminAuthentication(ctx, validateCallerHasAdminForAllRoles(req.Msg.Roles)); err != nil {
		return nil, ErrResponse(log, err)
	}

	if req.Msg.AuthType != "" {
		if _, ok := s.Authorizers[cosmov1alpha1.UserAuthType(req.Msg.AuthType)]; !ok {
			log.Info("authrizer not found", "username", req.Msg.UserName, "authType", req.Msg.AuthType)
			return nil, ErrResponse(log, kosmo.NewBadRequestError(
				fmt.Sprintf("auth-type '%s' is not supported", req.Msg.AuthType), nil))
		}
	}

	// create user
	user, err := s.Klient.CreateUser(ctx, req.Msg.UserName, req.Msg.DisplayName,
		req.Msg.Roles, req.Msg.AuthType, convertDashv1alpha1UserAddonToUserAddon(req.Msg.Addons))
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	resUser := convertUserToDashv1alpha1User(*user)

	if user.Spec.AuthType == cosmov1alpha1.UserAuthTypePasswordSecert {
		// Wait until user created
		defaultPassword, err := s.Klient.GetDefaultPasswordAwait(ctx, req.Msg.UserName)
		if err != nil {
			return nil, ErrResponse(log, err)
		}
		resUser.DefaultPassword = *defaultPassword
	}

	res := &dashv1alpha1.CreateUserResponse{
		Message: "Successfully created",
		User:    resUser,
	}
	log.Info(res.Message, "username", user.Name)
	return connect_go.NewResponse(res), nil
}

func (s *Server) GetUsers(ctx context.Context, req *connect_go.Request[emptypb.Empty]) (*connect_go.Response[dashv1alpha1.GetUsersResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	// admin users can get all users
	if err := adminAuthentication(ctx, passAllAdmin); err != nil {
		return nil, ErrResponse(log, err)
	}

	users, err := s.Klient.ListUsers(ctx)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.GetUsersResponse{}
	res.Items = make([]*dashv1alpha1.User, len(users))
	for i := range users {
		res.Items[i] = convertUserToDashv1alpha1User(users[i])
	}
	if len(res.Items) == 0 {
		res.Message = "No items found"
	}
	return connect_go.NewResponse(res), nil
}

func (s *Server) GetUser(ctx context.Context, req *connect_go.Request[dashv1alpha1.GetUserRequest]) (*connect_go.Response[dashv1alpha1.GetUserResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	if err := userAuthentication(ctx, req.Msg.UserName); err != nil {
		return nil, ErrResponse(log, err)
	}

	user, err := s.Klient.GetUser(ctx, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.GetUserResponse{
		User: convertUserToDashv1alpha1User(*user),
	}
	return connect_go.NewResponse(res), nil
}

func (s *Server) DeleteUser(ctx context.Context, req *connect_go.Request[dashv1alpha1.DeleteUserRequest]) (*connect_go.Response[dashv1alpha1.DeleteUserResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	targetUser, err := s.Klient.GetUser(ctx, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	// group-admin user can delete users which have only the their groups
	if err := adminAuthentication(ctx, validateCallerHasAdminForAllRoles(convertUserRolesToStringSlice(targetUser.Spec.Roles))); err != nil {
		return nil, ErrResponse(log, err)
	}

	caller := callerFromContext(ctx)

	if req.Msg.UserName == caller.Name {
		err := kosmo.NewBadRequestError("trying to delete yourself", nil)
		return nil, ErrResponse(log, err)
	}

	user, err := s.Klient.DeleteUser(ctx, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.DeleteUserResponse{
		Message: "Successfully deleted",
		User:    convertUserToDashv1alpha1User(*user),
	}
	log.Info(res.Message, "username", req.Msg.UserName)
	return connect_go.NewResponse(res), nil
}

func convertUserToDashv1alpha1User(user cosmov1alpha1.User) *dashv1alpha1.User {
	addons := make([]*dashv1alpha1.UserAddon, len(user.Spec.Addons))
	for i, v := range user.Spec.Addons {
		addons[i] = &dashv1alpha1.UserAddon{
			Template:      v.Template.Name,
			ClusterScoped: v.Template.ClusterScoped,
			Vars:          v.Vars,
		}
	}

	return &dashv1alpha1.User{
		Name:        user.Name,
		DisplayName: user.Spec.DisplayName,
		Roles:       convertUserRolesToStringSlice(user.Spec.Roles),
		AuthType:    user.Spec.AuthType.String(),
		Addons:      addons,
		Status:      string(user.Status.Phase),
	}
}

func convertUserRolesToStringSlice(apiRoles []cosmov1alpha1.UserRole) []string {
	roles := make([]string, 0, len(apiRoles))
	for _, v := range apiRoles {
		roles = append(roles, v.Name)
	}
	return roles
}

func convertDashv1alpha1UserAddonToUserAddon(addons []*dashv1alpha1.UserAddon) []cosmov1alpha1.UserAddon {
	a := make([]cosmov1alpha1.UserAddon, len(addons))
	for i, v := range addons {
		addon := cosmov1alpha1.UserAddon{
			Template: cosmov1alpha1.UserAddonTemplateRef{
				Name:          v.Template,
				ClusterScoped: v.ClusterScoped,
			},
			Vars: v.Vars,
		}
		a[i] = addon
	}
	return a
}
