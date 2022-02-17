package dashboard

import (
	"context"
	"errors"
	"net/http"
	"time"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/gorilla/mux"
)

func (s *Server) useAuthMiddleWare(router *mux.Router, routes dashv1alpha1.Routes) {
	for _, rt := range routes {
		router.Get(rt.Name).Handler(s.contextMiddleware(router.Get(rt.Name).GetHandler()))
	}
	router.Get("Verify").Handler(s.authorizationMiddleware(router.Get("Verify").GetHandler()))
}

func (s *Server) Verify(ctx context.Context) (dashv1alpha1.ImplResponse, error) {

	caller := callerFromContext(ctx)
	if caller == nil {
		return ErrorResponse(http.StatusUnauthorized, "")
	}
	deadline := deadlineFromContext(ctx)
	if deadline.Before(time.Now()) {
		return ErrorResponse(http.StatusUnauthorized, "session is invalid")
	}

	res := &dashv1alpha1.VerifyResponse{
		Id:       caller.Name,
		ExpireAt: deadline,
	}
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) Logout(ctx context.Context) (dashv1alpha1.ImplResponse, error) {

	w := responseWriterFromContext(ctx)
	r := requestFromContext(ctx)

	_, _, err := s.authorizeWithSession(r)
	if err != nil {
		if errors.Is(err, ErrNotAuthorized) {
			return ErrorResponse(http.StatusUnauthorized, "session is invalid")
		} else {
			return ErrorResponse(http.StatusInternalServerError, "")
		}
	}

	cookie := s.sessionCookieKey()
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)

	return NormalResponse(http.StatusOK, nil)
}

func (s *Server) Login(ctx context.Context, req dashv1alpha1.LoginRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	w := responseWriterFromContext(ctx)
	r := requestFromContext(ctx)

	// Check ID
	user, err := s.Klient.GetUser(ctx, req.Id)
	if err != nil {
		log.Info(err.Error(), "userid", req.Id)
		return ErrorResponse(http.StatusForbidden, "incorrect user or password")
	}
	// Check password
	authrizer, ok := s.Authorizers[user.Spec.AuthType]
	if !ok {
		log.Info("authrizer not found", "userid", req.Id, "authType", user.Spec.AuthType)
		return ErrorResponse(http.StatusServiceUnavailable, "")
	}
	verified, err := authrizer.Authorize(ctx, req)
	if err != nil {
		log.Error(err, "authorize failed", "userid", req.Id)
		return ErrorResponse(http.StatusForbidden, "incorrect user or password")
	}
	if !verified {
		log.Info("login failed: password invalid", "userid", req.Id)
		return ErrorResponse(http.StatusForbidden, "incorrect user or password")
	}
	var isDefault bool
	if wsv1alpha1.UserAuthType(user.Spec.AuthType) == wsv1alpha1.UserAuthTypeKosmoSecert {
		isDefault, err = s.Klient.IsDefaultPassword(ctx, req.Id)
		if err != nil {
			log.Error(err, "failed to check is default password", "userid", req.Id)
			return ErrorResponse(http.StatusInternalServerError, "")
		}
	}

	// Create session
	now := time.Now()
	expireAt := now.Add(time.Duration(s.MaxAgeSeconds) * time.Second)

	ses, _ := s.sessionStore.New(r, s.SessionName)
	sesInfo := session.Info{
		UserID:   req.Id,
		Deadline: expireAt.Unix(),
	}
	log.DebugAll().Info("save session", "userID", sesInfo.UserID, "deadline", sesInfo.Deadline)
	ses = session.Set(ses, sesInfo)

	err = s.sessionStore.Save(r, w, ses)
	if err != nil {
		log.Error(err, "failed to save session")
		return ErrorResponse(http.StatusInternalServerError, "")
	}

	res := &dashv1alpha1.LoginResponse{
		Id:                    req.Id,
		ExpireAt:              expireAt,
		RequirePasswordUpdate: isDefault,
	}

	return NormalResponse(http.StatusOK, res)
}
