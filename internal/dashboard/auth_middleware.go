package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

type ctxKeyCaller struct{}

func newContextWithCaller(ctx context.Context, caller *cosmov1alpha1.User) context.Context {
	return context.WithValue(ctx, ctxKeyCaller{}, caller)
}

func callerFromContext(ctx context.Context) *cosmov1alpha1.User {
	caller, ok := ctx.Value(ctxKeyCaller{}).(*cosmov1alpha1.User)
	if ok && caller != nil {
		return caller.DeepCopy()
	}
	return nil
}

type ctxKeyResponseWriter struct{}

func responseWriterFromContext(ctx context.Context) http.ResponseWriter {
	w := ctx.Value(ctxKeyResponseWriter{}).(http.ResponseWriter)
	return w
}

type ctxKeyRequest struct{}

func requestFromContext(ctx context.Context) *http.Request {
	r := ctx.Value(ctxKeyRequest{}).(*http.Request)
	return r
}

func (s *Server) contextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxKeyResponseWriter{}, w)
		ctx = context.WithValue(ctx, ctxKeyRequest{}, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) authorizationInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {

			log := clog.FromContext(ctx).WithName("authorization")

			loginUser, deadline, err := s.verifyAndGetLoginUser(ctx)
			if err != nil {
				return nil, ErrResponse(log, err)
			}

			ctx = newContextWithCaller(ctx, loginUser)
			ctx, cancel := context.WithDeadline(ctx, deadline)
			defer cancel()

			return next(ctx, req)
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}

func (s *Server) verifyAndGetLoginUser(ctx context.Context) (loginUser *cosmov1alpha1.User, deadline time.Time, err error) {
	r := requestFromContext(ctx)
	if r.Header.Get("Cookie") == "" {
		return nil, deadline, kosmo.NewUnauthorizedError("session is not found", err)
	}
	ses, err := s.sessionStore.Get(r, s.SessionName)
	if ses == nil || err != nil {
		return nil, deadline, kosmo.NewUnauthorizedError("failed to get session from store", err)
	}
	if ses.IsNew {
		return nil, deadline, kosmo.NewUnauthorizedError("session is invarild", err)
	}

	sesInfo := session.Get(ses)

	userName := sesInfo.UserName
	if userName == "" {
		return nil, deadline, kosmo.NewInternalServerError("userName is empty", nil)
	}

	deadline = time.Unix(sesInfo.Deadline, 0)
	if deadline.Before(time.Now()) {
		return nil, deadline,
			kosmo.NewUnauthorizedError(fmt.Sprintf("deadline is before the current time: deadline %v", deadline), nil)
	}

	loginUser, err = s.Klient.GetUser(ctx, userName)
	if err != nil {
		return nil, deadline, err
	}

	return loginUser, deadline, nil
}

func (s *Server) userAuthentication(ctx context.Context, userName string) error {
	log := clog.FromContext(ctx).WithCaller()

	caller := callerFromContext(ctx)
	if caller == nil {
		return kosmo.NewInternalServerError("invalid user authentication: NOT authorized", nil)
	}

	if caller.Name != userName {
		if cosmov1alpha1.HasAdminRole(caller.Spec.Roles) {
			// Admin user have access to all resources
			log.WithName("audit").Info(fmt.Sprintf("admin request %s", caller.Name), "username", caller.Name)
		} else {
			// General User have access only to the own resources
			log.Info("invalid user authentication: general user trying to access other's resource", "username", caller.Name, "target", userName)
			return kosmo.NewForbiddenError("", nil)
		}
	}
	return nil
}

func (s *Server) adminAuthentication(ctx context.Context) error {
	log := clog.FromContext(ctx).WithCaller()

	caller := callerFromContext(ctx)
	if caller == nil {
		return kosmo.NewInternalServerError("invalid user authentication: NOT authorized", nil)
	}

	// Check if the user role is Admin
	if !cosmov1alpha1.HasAdminRole(caller.Spec.Roles) {
		log.Info("invalid admin authentication: NOT cosmo-admin", "username", caller.Name)
		return kosmo.NewForbiddenError("", nil)
	}

	log.WithName("audit").Info(fmt.Sprintf("admin request %s", caller.Name), "username", caller.Name)
	return nil
}
