package dashboard

import (
	"context"
	"net/http"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/gorilla/mux"
)

func (s *Server) useUserMiddleWare(router *mux.Router, routes dashv1alpha1.Routes) {
	for _, rtName := range []string{"GetUsers", "PostUser", "DeleteUser", "PutUserRole"} {
		router.Get(rtName).Handler(
			s.authorizationMiddleware(
				s.adminAuthenticationMiddleware(
					router.Get(rtName).GetHandler())))
	}
	for _, rtName := range []string{"GetUser", "PutUserPassword", "PutUserName"} {
		router.Get(rtName).Handler(
			s.authorizationMiddleware(
				s.userAuthenticationMiddleware(
					router.Get(rtName).GetHandler())))
	}
}

func (s *Server) PostUser(ctx context.Context, req dashv1alpha1.CreateUserRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	// create user
	user, err := s.Klient.CreateUser(ctx, req.Id, req.DisplayName,
		req.Role, req.AuthType, convertDashv1alpha1UserToUserAddon(req.Addons))
	if err != nil {
		return ErrorResponse(log, err)
	}

	// Wait until user created
	defaultPassword, err := s.Klient.GetDefaultPasswordAwait(ctx, req.Id)
	if err != nil {
		return ErrorResponse(log, err)
	}

	resUser := convertUserToDashv1alpha1User(*user)
	resUser.DefaultPassword = *defaultPassword
	res := &dashv1alpha1.CreateUserResponse{
		Message: "Successfully created",
		User:    resUser,
	}
	log.Info(res.Message, "userid", user.Name)
	return NormalResponse(http.StatusCreated, res)
}

func (s *Server) GetUsers(ctx context.Context) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()

	users, err := s.Klient.ListUsers(ctx)
	if err != nil {
		return ErrorResponse(log, err)
	}

	res := &dashv1alpha1.ListUsersResponse{}
	res.Items = make([]dashv1alpha1.User, len(users))
	for i := range users {
		res.Items[i] = *convertUserToDashv1alpha1User(users[i])
	}
	if len(res.Items) == 0 {
		res.Message = "No items found"
	}
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) GetUser(ctx context.Context, userId string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId)

	user, err := s.Klient.GetUser(ctx, userId)
	if err != nil {
		return ErrorResponse(log, err)
	}

	res := &dashv1alpha1.GetUserResponse{
		User: convertUserToDashv1alpha1User(*user),
	}
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) DeleteUser(ctx context.Context, userId string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId)

	caller := callerFromContext(ctx)
	if caller == nil {
		err := kosmo.NewInternalServerError("user is not found in context", nil)
		return ErrorResponse(log, err)
	}

	if userId == caller.Name {
		err := kosmo.NewForbiddenError("trying to delete own user", nil)
		return ErrorResponse(log, err)
	}

	user, err := s.Klient.DeleteUser(ctx, userId)
	if err != nil {
		return ErrorResponse(log, err)
	}

	res := &dashv1alpha1.DeleteUserResponse{
		Message: "Successfully deleted",
		User:    convertUserToDashv1alpha1User(*user),
	}
	log.Info(res.Message, "userid", userId)
	return NormalResponse(http.StatusOK, res)
}

func convertUserToDashv1alpha1User(user wsv1alpha1.User) *dashv1alpha1.User {
	addons := make([]dashv1alpha1.ApiV1alpha1UserAddons, len(user.Spec.Addons))
	for i, v := range user.Spec.Addons {
		addons[i] = dashv1alpha1.ApiV1alpha1UserAddons{
			Template: v.Template.Name,
			Vars:     v.Vars,
		}
	}

	return &dashv1alpha1.User{
		Id:          user.Name,
		DisplayName: user.Spec.DisplayName,
		Role:        user.Spec.Role.String(),
		AuthType:    user.Spec.AuthType.String(),
		Addons:      addons,
		Status:      string(user.Status.Phase),
	}
}

func convertDashv1alpha1UserToUserAddon(addons []dashv1alpha1.ApiV1alpha1UserAddons) []wsv1alpha1.UserAddon {
	a := make([]wsv1alpha1.UserAddon, len(addons))
	for i, v := range addons {
		addon := wsv1alpha1.UserAddon{
			Template: cosmov1alpha1.TemplateRef{
				Name: v.Template,
			},
			Vars: v.Vars,
		}
		a[i] = addon
	}
	return a
}
