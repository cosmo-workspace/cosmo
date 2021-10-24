package dashboard

import (
	"context"
	"net/http"
	"time"

	apierrs "k8s.io/apimachinery/pkg/api/errors"

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
	for _, rtName := range []string{"GetUser", "DeleteUser", "PutUserPassword"} {
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
		return dashv1alpha1.Response(http.StatusBadRequest, nil), nil
	}

	authtype := wsv1alpha1.UserAuthType(req.AuthType)
	if authtype == "" {
		authtype = wsv1alpha1.UserAuthTypeKosmoSecert
	}
	if !authtype.IsValid() {
		log.Info("invalid request", "id", req.Id, "authtype", authtype)
		return dashv1alpha1.Response(http.StatusBadRequest, nil), nil
	}

	res := &dashv1alpha1.CreateUserResponse{}

	user := &wsv1alpha1.User{
		ID:          req.Id,
		DisplayName: req.DisplayName,
		Role:        userrole,
		AuthType:    authtype,
	}
	log.Debug().Info("creating user object", "user", user)
	var err error
	user, err = s.Klient.CreateUser(ctx, user)
	if err != nil {
		if apierrs.IsAlreadyExists(err) {
			res.Message = "User already exists"
			return dashv1alpha1.Response(http.StatusTooManyRequests, res), nil
		} else {
			res.Message = "failed to create user"
			log.Error(err, res.Message, "userid", req.Id)
			return dashv1alpha1.Response(http.StatusServiceUnavailable, res), nil
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
			res.Message = "Reached to timeout in user creation"
			log.Error(err, res.Message, "userid", user.ID)
			return dashv1alpha1.Response(http.StatusInternalServerError, res), nil
		default:
			<-tk.C
		}
	}

	res.User = convertUserToDashv1alpha1User(*user)
	res.User.DefaultPassword = *defaultPassword

	res.Message = "Successfully created"
	log.Info(res.Message, "userid", user.ID)
	return dashv1alpha1.Response(http.StatusCreated, res), nil
}

func (s *Server) GetUsers(ctx context.Context) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()

	res := &dashv1alpha1.ListUsersResponse{}

	users, err := s.Klient.ListUsers(ctx)
	if err != nil {
		res.Message = "failed to list users"
		log.Error(err, res.Message)
		return dashv1alpha1.Response(http.StatusInternalServerError, res), nil
	}
	res.Items = make([]dashv1alpha1.User, len(users))
	for i := range users {
		res.Items[i] = *convertUserToDashv1alpha1User(users[i])
	}

	if len(res.Items) == 0 {
		res.Message = "No items found"
	}
	return dashv1alpha1.Response(http.StatusOK, res), nil
}

// func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
func (s *Server) GetUser(ctx context.Context, userId string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId)

	user := userFromContext(ctx)
	if user == nil {
		log.Info("user not found in context")
		return dashv1alpha1.Response(http.StatusInternalServerError, nil), nil
	}

	res := &dashv1alpha1.GetUserResponse{}
	res.User = convertUserToDashv1alpha1User(*user)
	return dashv1alpha1.Response(http.StatusOK, res), nil
}

func (s *Server) DeleteUser(ctx context.Context, userId string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId)

	user := userFromContext(ctx)
	if user == nil {
		log.Info("user not found in context")
		return dashv1alpha1.Response(http.StatusInternalServerError, nil), nil
	}

	res := &dashv1alpha1.DeleteUserResponse{}

	deleted, err := s.Klient.DeleteUser(ctx, user.ID)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return dashv1alpha1.Response(http.StatusNotFound, res), nil
		} else {
			res.Message = "failed to delete user"
			log.Error(err, res.Message, "userid", user.ID)
			return dashv1alpha1.Response(http.StatusInternalServerError, res), nil
		}
	}

	res.User = convertUserToDashv1alpha1User(*deleted)
	res.Message = "Successfully deleted"
	log.Info(res.Message, "userid", user.ID)
	return dashv1alpha1.Response(http.StatusOK, res), nil
}

func convertUserToDashv1alpha1User(user wsv1alpha1.User) *dashv1alpha1.User {
	return &dashv1alpha1.User{
		Id:          user.ID,
		DisplayName: user.DisplayName,
		Role:        user.Role.String(),
		AuthType:    user.AuthType.String(),
		Status:      string(user.Status),
	}
}
