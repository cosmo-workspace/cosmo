package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
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
	TLSPrivateKeyPath   string
	TLSCertPath         string
	Insecure            bool

	CookieDomain      string
	CookieHashKey     string
	CookieBlockKey    string
	CookieSessionName string

	Authorizers map[cosmov1alpha1.UserAuthType]auth.Authorizer

	http         *http.Server
	sessionStore sessions.Store
}

func (s *Server) setupRouter() {

	mux := http.NewServeMux()

	// setup proto api
	s.AuthServiceHandler(mux)
	s.UserServiceHandler(mux)
	s.TemplateServiceHandler(mux)
	s.WorkspaceServiceHandler(mux)

	// setup serving static files
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(s.StaticFileDir))))

	// setup middlewares for all routers to use HTTPRequestLogger and TimeoutHandler.
	// deadline of the Timeout handler takes precedence over any subsequent deadlines
	reqLogr := NewHTTPRequestLogger(s.Log)
	s.http.Handler = reqLogr.Middleware(s.timeoutHandler(mux))
}

func (s *Server) setupSessionStore() {
	store := session.NewStore([]byte(s.CookieHashKey), []byte(s.CookieBlockKey), s.sessionCookieKey())
	s.sessionStore = store
}

func (s *Server) sessionCookieKey() *http.Cookie {
	return &http.Cookie{
		Name:     s.CookieSessionName,
		Domain:   s.CookieDomain,
		MaxAge:   s.MaxAgeSeconds,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
}

func (s *Server) timeoutHandler(next http.Handler) http.Handler {
	//return http.TimeoutHandler(next, s.ResponseTimeout, `{"message": "Request timeout"}`+"\n")
	return http.TimeoutHandler(next, s.ResponseTimeout, "")
}

// Start run server
func (s *Server) Start(ctx context.Context) error {
	s.http = &http.Server{
		Addr: fmt.Sprintf(":%d", s.Port),
	}

	s.setupRouter()

	s.setupSessionStore()

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
