package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/auth"
	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

// Server serves dashboard APIs and UI static files
// It implements https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/manager#Runnable
type Server struct {
	Log                 *clog.Logger
	Klient              kosmo.Client
	GracefulShutdownDur time.Duration
	ResponseTimeout     time.Duration
	StaticFileDir       string
	Port                int
	MaxAgeSeconds       int
	SessionName         string
	TLSPrivateKeyPath   string
	TLSCertPath         string
	Insecure            bool

	Authorizers map[wsv1alpha1.UserAuthType]auth.Authorizer

	http         *http.Server
	sessionStore sessions.Store
}

func (s *Server) setupRouter() error {

	authRouter := dashv1alpha1.NewAuthApiController(s)
	workspaceRouter := dashv1alpha1.NewWorkspaceApiController(s)
	templateRouter := dashv1alpha1.NewTemplateApiController(s)
	userRouter := dashv1alpha1.NewUserApiController(s)

	router := dashv1alpha1.NewRouter(authRouter, workspaceRouter, templateRouter, userRouter)

	s.useAuthMiddleWare(router, authRouter.Routes())
	s.useWorkspaceMiddleWare(router, workspaceRouter.Routes())
	s.useTemplateMiddleWare(router, templateRouter.Routes())
	s.useUserMiddleWare(router, userRouter.Routes())

	// setup middlewares for all routers to use HTTPRequestLogger and TimeoutHandler.
	// deadline of the Timeout handler takes precedence over any subsequent deadlines
	reqLogr := NewHTTPRequestLogger(s.Log)
	router.Use(reqLogr.Middleware, s.timeoutHandler)

	// setup serving static files
	router.NotFoundHandler = reqLogr.Middleware(http.StripPrefix("/", http.FileServer(http.Dir(s.StaticFileDir))))

	s.http.Handler = router
	return nil
}

func (s *Server) setupSessionStore() error {
	store := session.NewStore(securecookie.GenerateRandomKey(64), securecookie.GenerateRandomKey(32), s.sessionCookieKey())
	s.sessionStore = store
	return nil
}

func (s *Server) sessionCookieKey() *http.Cookie {
	return &http.Cookie{
		Name:     s.SessionName,
		MaxAge:   s.MaxAgeSeconds,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
}

func (s *Server) timeoutHandler(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, s.ResponseTimeout, "")
}

// Start run server
func (s *Server) Start(ctx context.Context) error {
	s.http = &http.Server{
		Addr: fmt.Sprintf(":%d", s.Port),
	}

	if err := s.setupRouter(); err != nil {
		return err
	}

	if err := s.setupSessionStore(); err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		s.Log.Info("shutdown server")
		s.shutdown()
	}()

	if s.Insecure {
		s.Log.Info("WARNING: start insecure server")
		return s.http.ListenAndServe()

	} else {
		s.Log.Info("start server")
		return s.http.ListenAndServeTLS(s.TLSCertPath, s.TLSPrivateKeyPath)
	}
}

func (s *Server) shutdown() error {
	gracefulShutdownCtx, cancel := context.WithTimeout(context.Background(), s.GracefulShutdownDur)
	defer cancel()
	return s.http.Shutdown(gracefulShutdownCtx)
}
