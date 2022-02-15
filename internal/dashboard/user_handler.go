package dashboard

import (
	"context"
	"net/http"
	"sort"
	"time"

	apierrs "k8s.io/apimachinery/pkg/api/errors"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/gorilla/mux"
)

func (s *Server) useUserMiddleWare(router *mux.Router, routes dashv1alpha1.Routes) {
	for _, rtName := range []string{"GetUsers", "PostUser"} {
		router.Get(rtName).Handler(s.adminAuthenticationMiddleware(router.Get(rtName).GetHandler()))
		router.Get(rtName).Handler(s.authorizationMiddleware(router.Get(rtName).GetHandler()))
	}
	for _, rtName := range []string{"PutUserRole"} {
		router.Get(rtName).Handler(s.adminAuthenticationMiddleware(router.Get(rtName).GetHandler()))
		router.Get(rtName).Handler(s.preFetchUserMiddleware(router.Get(rtName).GetHandler()))
		router.Get(rtName).Handler(s.authorizationMiddleware(router.Get(rtName).GetHandler()))
	}
	for _, rtName := range []string{"GetUser", "DeleteUser", "PutUserPassword", "PutUserName"} {
		router.Get(rtName).Handler(s.userAuthenticationMiddleware(router.Get(rtName).GetHandler()))
		router.Get(rtName).Handler(s.preFetchUserMiddleware(router.Get(rtName).GetHandler()))
		router.Get(rtName).Handler(s.authorizationMiddleware(router.Get(rtName).GetHandler()))
	}
}

func (s *Server) PostUser(ctx context.Context, req dashv1alpha1.CreateUserRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	log.Info("creating user", "id", req.Id, "displayName", req.DisplayName, "role", req.Role)

	if req.DisplayName == "" {
		req.DisplayName = req.Id
	}

	userrole := wsv1alpha1.UserRole(req.Role)
	if !userrole.IsValid() {
		log.Info("invalid request", "id", req.Id, "role", userrole)
		return ErrorResponse(http.StatusBadRequest, "'userrole' is invalid")
	}

	authtype := wsv1alpha1.UserAuthType(req.AuthType)
	if authtype == "" {
		authtype = wsv1alpha1.UserAuthTypeKosmoSecert
	}
	if !authtype.IsValid() {
		log.Info("invalid request", "id", req.Id, "authtype", authtype)
		return ErrorResponse(http.StatusBadRequest, "'authtype' is invalid")
	}

	res := &dashv1alpha1.CreateUserResponse{}

	user := &wsv1alpha1.User{}
	user.SetName(req.Id)
	user.Spec = wsv1alpha1.UserSpec{
		DisplayName: req.DisplayName,
		Role:        userrole,
		AuthType:    authtype,
		Addons:      convertDashv1alpha1UserToUserAddon(req.Addons),
	}

	log.Debug().Info("creating user object", "user", user)

	err := s.Klient.Create(ctx, user)
	if err != nil {
		if apierrs.IsAlreadyExists(err) {
			return ErrorResponse(http.StatusTooManyRequests, "user already exists")
		} else {
			message := "failed to create user"
			log.Error(err, message, "userid", req.Id)
			return ErrorResponse(http.StatusServiceUnavailable, message)
		}
	}

	// Wait until user created
	tk := time.NewTicker(time.Second)
	defer tk.Stop()
	var defaultPassword *string
	log.Debug().Info("wait for default password creation", "user", user)

UserCreationWaitLoop:
	for {
		p, err := s.Klient.GetDefaultPassword(ctx, req.Id)
		if err == nil {
			log.Debug().Info("got default password")
			tk.Stop()
			defaultPassword = p
			break UserCreationWaitLoop
		}

		select {
		case <-ctx.Done():
			tk.Stop()
			message := "Reached to timeout in user creation"
			log.Error(err, message, "userid", user.Name)
			return ErrorResponse(http.StatusInternalServerError, message)
		default:
			<-tk.C
		}
	}

	res.User = convertUserToDashv1alpha1User(*user)
	res.User.DefaultPassword = *defaultPassword

	res.Message = "Successfully created"
	log.Info(res.Message, "userid", user.Name)
	return NormalResponse(http.StatusCreated, res)
}

func (s *Server) GetUsers(ctx context.Context) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()

	res := &dashv1alpha1.ListUsersResponse{}

	users, err := s.Klient.ListUsers(ctx)
	if err != nil {
		message := "failed to list users"
		log.Error(err, message)
		return ErrorResponse(http.StatusInternalServerError, message)
	}
	res.Items = make([]dashv1alpha1.User, len(users))
	for i := range users {
		res.Items[i] = *convertUserToDashv1alpha1User(users[i])
	}

	sort.Slice(res.Items, func(i, j int) bool { return res.Items[i].Id < res.Items[j].Id })

	if len(res.Items) == 0 {
		res.Message = "No items found"
	}
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) GetUser(ctx context.Context, userId string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId)

	user := userFromContext(ctx)
	if user == nil {
		log.Info("user not found in context")
		return ErrorResponse(http.StatusInternalServerError, "")
	}

	res := &dashv1alpha1.GetUserResponse{}
	res.User = convertUserToDashv1alpha1User(*user)
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) DeleteUser(ctx context.Context, userId string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId)

	user := userFromContext(ctx)
	if user == nil {
		log.Info("user not found in context")
		return ErrorResponse(http.StatusInternalServerError, "")
	}

	res := &dashv1alpha1.DeleteUserResponse{}

	err := s.Klient.Delete(ctx, user)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return dashv1alpha1.Response(http.StatusNotFound, res), nil
		} else {
			message := "failed to delete user"
			log.Error(err, message, "userid", user.Name)
			return ErrorResponse(http.StatusInternalServerError, message)
		}
	}

	res.User = convertUserToDashv1alpha1User(*user)
	res.Message = "Successfully deleted"
	log.Info(res.Message, "userid", user.Name)
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
