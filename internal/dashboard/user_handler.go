package dashboard

import (
	"context"
	"fmt"
	"net/http"

	connect_go "github.com/bufbuild/connect-go"
	apierrs "k8s.io/apimachinery/pkg/api/errors"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
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
			return nil, ErrResponse(log, apierrs.NewBadRequest(
				fmt.Sprintf("auth-type '%s' is not supported", req.Msg.AuthType)))
		}
	}

	// create user
	user, err := s.Klient.CreateUser(ctx, req.Msg.UserName, req.Msg.DisplayName,
		req.Msg.Roles, req.Msg.AuthType, apiconv.D2C_UserAddons(req.Msg.Addons))
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	resUser := apiconv.C2D_User(*user)

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

func (s *Server) GetUsers(ctx context.Context, req *connect_go.Request[dashv1alpha1.GetUsersRequest]) (*connect_go.Response[dashv1alpha1.GetUsersResponse], error) {
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

	res := &dashv1alpha1.GetUsersResponse{
		Items: apiconv.C2D_Users(users, apiconv.WithUserRaw(req.Msg.WithRaw)),
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
	events, err := s.Klient.ListEvents(ctx, cosmov1alpha1.UserNamespace(user.Name))
	if err != nil {
		log.Error(err, "failed to list events", "user", user.Name)
	}

	res := &dashv1alpha1.GetUserResponse{
		User: apiconv.C2D_User(*user, apiconv.WithUserRaw(req.Msg.WithRaw)),
	}
	res.User.Events = apiconv.K2D_Events(events)

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
	if err := adminAuthentication(ctx, validateCallerHasAdminForAllRoles(apiconv.C2S_UserRole(targetUser.Spec.Roles))); err != nil {
		return nil, ErrResponse(log, err)
	}

	caller := callerFromContext(ctx)

	if req.Msg.UserName == caller.Name {
		err := apierrs.NewBadRequest("trying to delete yourself")
		return nil, ErrResponse(log, err)
	}

	user, err := s.Klient.DeleteUser(ctx, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.DeleteUserResponse{
		Message: "Successfully deleted",
		User:    apiconv.C2D_User(*user),
	}
	log.Info(res.Message, "username", req.Msg.UserName)
	return connect_go.NewResponse(res), nil
}
