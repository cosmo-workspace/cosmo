package dashboard

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	connect_go "github.com/bufbuild/connect-go"
	apierrs "k8s.io/apimachinery/pkg/api/errors"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
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

type authorizationInterceptorFunc func(ctx context.Context) (loginUser *cosmov1alpha1.User, deadline time.Time, err error)

// WrapUnary implements [Interceptor] with an authorization.
func (f authorizationInterceptorFunc) WrapUnary(next connect_go.UnaryFunc) connect_go.UnaryFunc {
	return connect_go.UnaryFunc(func(ctx context.Context, req connect_go.AnyRequest) (connect_go.AnyResponse, error) {
		log := clog.FromContext(ctx).WithName("unary authorization")

		loginUser, deadline, err := f(ctx)
		if err != nil {
			return nil, ErrResponse(log, err)
		}

		ctx = newContextWithCaller(ctx, loginUser)
		ctx, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()

		return next(ctx, req)
	})
}

// WrapStreamingClient implements [Interceptor] with a no-op.
func (f authorizationInterceptorFunc) WrapStreamingClient(next connect_go.StreamingClientFunc) connect_go.StreamingClientFunc {
	return next
}

// WrapStreamingHandler implements [Interceptor] with an authorization.
func (f authorizationInterceptorFunc) WrapStreamingHandler(next connect_go.StreamingHandlerFunc) connect_go.StreamingHandlerFunc {
	return connect_go.StreamingHandlerFunc(func(ctx context.Context, conn connect_go.StreamingHandlerConn) error {
		log := clog.FromContext(ctx).WithName("streaming handler authorization")

		loginUser, deadline, err := f(ctx)
		if err != nil {
			return ErrResponse(log, err)
		}

		ctx = newContextWithCaller(ctx, loginUser)
		ctx, cancel := context.WithDeadline(ctx, deadline)
		defer cancel()

		return next(ctx, conn)
	})
}

func (s *Server) verifyAndGetLoginUser(ctx context.Context) (loginUser *cosmov1alpha1.User, deadline time.Time, err error) {
	r := requestFromContext(ctx)
	if r.Header.Get("Cookie") == "" {
		return nil, deadline, apierrs.NewUnauthorized("session is not found")
	}
	ses, err := s.sessionStore.Get(r, s.CookieSessionName)
	if ses == nil || err != nil {
		return nil, deadline, apierrs.NewUnauthorized(fmt.Sprintf("failed to get session from store: %v", err))
	}
	if ses.IsNew {
		return nil, deadline, apierrs.NewUnauthorized("session is invarild")
	}

	sesInfo := session.Get(ses)

	userName := sesInfo.UserName
	if userName == "" {
		return nil, deadline, apierrs.NewInternalError(fmt.Errorf("userName is empty"))
	}

	deadline = time.Unix(sesInfo.Deadline, 0)
	if deadline.Before(time.Now()) {
		return nil, deadline,
			apierrs.NewUnauthorized(fmt.Sprintf("deadline is before the current time: deadline %v", deadline))
	}

	loginUser, err = s.Klient.GetUser(ctx, userName)
	if err != nil {
		return nil, deadline, err
	}

	return loginUser, deadline, nil
}

func userAuthentication(ctx context.Context, userName string) error {
	log := clog.FromContext(ctx).WithCaller()

	caller := callerFromContext(ctx)
	if caller == nil {
		return apierrs.NewInternalError(fmt.Errorf("invalid user authentication: NOT authorized"))
	}

	if caller.Name != userName {
		if cosmov1alpha1.HasPrivilegedRole(caller.Spec.Roles) {
			// Admin user have access to all resources
			log.WithName("audit").Info(fmt.Sprintf("admin request %s", caller.Name), "username", caller.Name)
		} else {
			// General User have access only to the own resources
			log.Info("invalid user authentication: general user trying to access other's resource", "username", caller.Name, "target", userName)
			return NewForbidden(fmt.Errorf("invalid user authentication"))
		}
	}
	return nil
}

func adminAuthentication(ctx context.Context, customAuthenFuncs ...func(callerGroupRoleMap map[string]string) error) error {
	log := clog.FromContext(ctx).WithCaller().WithName("audit")

	caller := callerFromContext(ctx)
	if caller == nil {
		return apierrs.NewInternalError(fmt.Errorf("invalid user authentication: NOT authorized"))
	}
	auditlog := log.WithValues("caller", caller.Name, "role", caller.Spec.Roles)

	auditlog.Info("try admin request")

	// pass if the user role is privileged
	if cosmov1alpha1.HasPrivilegedRole(caller.Spec.Roles) {
		auditlog.Info("privileged request")
		return nil
	}

	// deny if the user does not have admin role
	callerGroupRoleMap := caller.GetGroupRoleMap()
	err := validateCallerHasAdmin(callerGroupRoleMap)
	if err != nil {
		auditlog.Info(err.Error())
		return NewForbidden(err)
	}

	// pass if all custom authens are passed
	if len(customAuthenFuncs) > 0 {
		errs := make([]error, 0, len(customAuthenFuncs))
		for _, f := range customAuthenFuncs {
			if err := f(callerGroupRoleMap); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			auditlog.Info("custom admin authentication failed", "errs", errs)
			return NewForbidden(errs[0])
		}
		auditlog.Info("admin request is allowed")
		return nil
	}

	auditlog.Info("admin authentication failed")
	return NewForbidden(fmt.Errorf("admin authentication failed"))
}

func validateCallerHasAdmin(callerGroupRoleMap map[string]string) error {
	for _, role := range callerGroupRoleMap {
		if role == cosmov1alpha1.AdminRoleName {
			// Allow if caller has at least one administrative privilege.
			return nil
		}
	}
	return errors.New("not admin")
}

func validateCallerHasAdminForAllRoles(tryRoleNames []string) func(map[string]string) error {
	return func(callerGroupRoleMap map[string]string) error {
		for _, r := range tryRoleNames {
			tryAccessGroup, _ := (cosmov1alpha1.UserRole{Name: r}).GetGroupAndRole()
			callerRoleForTriedGroup := callerGroupRoleMap[tryAccessGroup]

			// Deny if caller does not have administrative privilege for tried group.
			if callerRoleForTriedGroup != cosmov1alpha1.AdminRoleName {
				return fmt.Errorf("denied to access '%s' group", tryAccessGroup)
			}
		}
		return nil
	}
}

func validateCallerHasAdminForAtLeastOneRole(tryRoleNames []cosmov1alpha1.UserRole) func(map[string]string) error {
	return func(callerGroupRoleMap map[string]string) error {
		for _, r := range tryRoleNames {
			tryAccessGroup, _ := r.GetGroupAndRole()
			callerRoleForTriedGroup := callerGroupRoleMap[tryAccessGroup]

			// Allow if caller has at least one administrative privilege for tried group.
			if callerRoleForTriedGroup == cosmov1alpha1.AdminRoleName {
				return nil
			}
		}
		return fmt.Errorf("denied to access")
	}
}

var passAllAdmin = func(map[string]string) error {
	return nil
}
