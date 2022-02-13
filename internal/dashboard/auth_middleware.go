package dashboard

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	apierrs "k8s.io/apimachinery/pkg/api/errors"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/gorilla/mux"
)

type ctxKeyCaller struct{}

func newContextWithCaller(ctx context.Context, caller *wsv1alpha1.User) context.Context {
	return context.WithValue(ctx, ctxKeyCaller{}, caller)
}

func callerFromContext(ctx context.Context) *wsv1alpha1.User {
	caller, ok := ctx.Value(ctxKeyCaller{}).(*wsv1alpha1.User)
	if ok && caller != nil {
		return caller.DeepCopy()
	}
	return nil
}

type ctxKeyDeadline struct{}

func newContextWithDeadline(ctx context.Context, deadline time.Time) context.Context {
	return context.WithValue(ctx, ctxKeyDeadline{}, deadline)
}

func deadlineFromContext(ctx context.Context) time.Time {
	deadline, ok := ctx.Value(ctxKeyDeadline{}).(time.Time)
	if ok {
		return deadline
	}
	return time.Time{}
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

var (
	ErrNotAuthorized = errors.New("not authroized")
)

func (s *Server) authorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := clog.FromContext(ctx).WithName("authorization")

		userID, deadline, err := s.authorizeWithSession(r)
		if err != nil {
			log.Error(err, "session authorization err")

			if errors.Is(err, ErrNotAuthorized) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		caller, err := s.Klient.GetUser(ctx, userID)
		if err != nil {
			if apierrs.IsNotFound(err) {
				log.Error(err, "caller not found", "callerID", userID)
				w.WriteHeader(http.StatusUnauthorized)
				return
			} else {
				log.Error(err, "failed to get caller", "callerID", userID)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		ctx = newContextWithCaller(ctx, caller)
		ctx = newContextWithDeadline(ctx, deadline)

		ctx, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) authorizeWithSession(r *http.Request) (userID string, deadline time.Time, err error) {
	ses, err := s.sessionStore.Get(r, s.SessionName)
	if ses == nil || err != nil {
		return userID, deadline, fmt.Errorf("%w: failed to get session from store: %v", ErrNotAuthorized, err)
	}
	if ses.IsNew {
		return userID, deadline, fmt.Errorf("%w: %v", ErrNotAuthorized, err)
	}

	sesInfo := session.Get(ses)

	userID = sesInfo.UserID
	if userID == "" {
		return userID, deadline, fmt.Errorf("userID is empty")
	}

	deadline = time.Unix(sesInfo.Deadline, 0)
	if deadline.Before(time.Now()) {
		return userID, deadline, fmt.Errorf("deadline is before the current time: deadline %v", deadline)
	}

	return userID, deadline, nil
}

func (s *Server) userAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := clog.FromContext(ctx).WithName("userAuthentication")

		caller := callerFromContext(ctx)
		if caller == nil {
			log.Info("invalid user authentication: NOT authorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Get UserID from path
		vars := mux.Vars(r)
		userID := vars["userid"]

		if caller.Name != userID {
			if wsv1alpha1.UserRole(caller.Spec.Role).IsAdmin() {
				// Admin user have access to all resources
				log.WithName("audit").Info(fmt.Sprintf("admin request %s %s %s", caller.Name, r.Method, r.URL),
					"userid", caller.Name, "method", r.Method, "path", r.URL, "host", r.Host, "X-Forwarded-For", r.Header.Get("X-Forwarded-For"), "user-agent", r.UserAgent())

			} else {
				// General User have access only to the own resources
				log.Info("invalid user authentication: general user trying to access other's resource", "userid", caller.Name, "target", userID)
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) adminAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := clog.FromContext(ctx).WithName("adminAuthentication")

		caller := callerFromContext(ctx)
		if caller == nil {
			log.Info("invalid user authentication: NOT authorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Check if the user role is Admin
		if !wsv1alpha1.UserRole(caller.Spec.Role).IsAdmin() {
			log.Info("invalid admin authentication: NOT cosmo-admin", "userid", caller.Name)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Get UserID from path
		vars := mux.Vars(r)
		userID := vars["userid"]

		if caller.Name == userID {
			var match mux.RouteMatch
			if s.http.Handler.(*mux.Router).Match(r, &match) {
				if match.Route.GetName() == "DeleteUser" {
					// Prevent deletion of own user
					log.Info("invalid user authentication: trying to delete own user", "userid", caller.Name, "target", userID)
					w.WriteHeader(http.StatusForbidden)
					return
				}
			}
		}

		log.WithName("audit").Info(fmt.Sprintf("admin request %s %s %s", caller.Name, r.Method, r.URL),
			"userid", caller.Name, "method", r.Method, "path", r.URL, "host", r.Host, "X-Forwarded-For", r.Header.Get("X-Forwarded-For"), "user-agent", r.UserAgent())
		next.ServeHTTP(w, r)
	})
}
