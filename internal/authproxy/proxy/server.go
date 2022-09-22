package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/cosmo-workspace/cosmo/pkg/auth"
	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

const (
	pathLoginAPI = "/api/login"
)

// ProxyServer is a http(s) server that serve login UI and reverse-proxy to the backend with authentication
type ProxyServer struct {
	Log           *clog.Logger
	User          string
	StaticFileDir string
	MaxAgeSeconds int
	SessionName   string
	RedirectPath  string

	Insecure          bool
	TLSCertPath       string
	TLSPrivateKeyPath string

	http         *http.Server
	listener     net.Listener
	sessionStore sessions.Store
	authorizer   auth.Authorizer
}

func (p *ProxyServer) sessionCookieKey() *http.Cookie {
	cookie := &http.Cookie{
		Name:     p.SessionName,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	if p.MaxAgeSeconds > 0 {
		cookie.MaxAge = p.MaxAgeSeconds
	}

	return cookie
}

func (p *ProxyServer) SetupSessionStore(hashKey, blockKey []byte) {
	store := session.NewStore(hashKey, blockKey, p.sessionCookieKey())
	p.sessionStore = store
}

func (p *ProxyServer) SetupReverseProxy(addr string, targetURL *url.URL) *ProxyServer {
	router := mux.NewRouter()
	router.Use(p.log)
	router.NotFoundHandler = p.log(p.auth(httputil.NewSingleHostReverseProxy(targetURL)))

	router.Path(p.RedirectPath + pathLoginAPI).Methods(http.MethodPost).HandlerFunc(p.handleLogin)
	router.PathPrefix(p.RedirectPath).Methods(http.MethodGet).Handler(p.serveLoginPage())

	p.http = &http.Server{
		Addr:    addr,
		Handler: router,
	}
	return p
}

func (p *ProxyServer) SetupAuthorizer(a auth.Authorizer) {
	p.authorizer = a
}

func (p *ProxyServer) GetListenerPort() int {
	if p.listener == nil {
		return 0
	}
	tcpAddr, ok := p.listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0
	}
	return tcpAddr.Port
}

func (p *ProxyServer) Start(ctx context.Context, gracefulShutdownDur time.Duration) error {
	if err := p.validate(); err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	if p.Insecure {
		ln, err := net.Listen("tcp", p.http.Addr)
		if err != nil {
			return err
		}
		p.listener = ln
		p.Log.Info("WARNING: starting insecure server")

	} else {
		cer, err := tls.LoadX509KeyPair(p.TLSCertPath, p.TLSPrivateKeyPath)
		if err != nil {
			p.Log.Error(err, "failed to load keypair")
			return err
		}
		cfg := tls.Config{Certificates: []tls.Certificate{cer}}
		ln, err := tls.Listen("tcp", p.http.Addr, &cfg)
		if err != nil {
			return err
		}
		p.listener = ln
	}

	go func() {
		<-ctx.Done()
		p.Log.Info("shutdown server")
		p.shutdown(gracefulShutdownDur)
	}()

	p.Log.Info("start server")
	return p.http.Serve(p.listener)
}

func (p *ProxyServer) shutdown(gracefulShutdownDur time.Duration) error {
	gracefulShutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownDur)
	defer cancel()
	return p.http.Shutdown(gracefulShutdownCtx)
}

func (p *ProxyServer) validate() error {
	if p.http == nil {
		return errors.New("reverse proxy is not initialized")
	}
	if p.sessionStore == nil {
		return errors.New("session store is not initialized")
	}
	return nil
}

func (p *ProxyServer) log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.Log.WithName("access").Info(fmt.Sprintf("%s %s", r.Method, r.URL),
			"method", r.Method, "path", r.URL, "host", r.Host, "X-Forwarded-For", r.Header.Get("X-Forwarded-For"),
			"upgrade", r.Header.Get("upgrade"), "user-agent", r.UserAgent())
		next.ServeHTTP(w, r)
	})
}

func (p *ProxyServer) auth(next http.Handler) http.Handler {
	log := p.Log.WithName("auth")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Bypass manifest.json not to check session. By default, manifest.json is requested without cookie.
		// https://developer.mozilla.org/en-US/docs/Web/Manifest
		if strings.Contains(strings.ToLower(r.URL.Path), "/manifest.json") {
			next.ServeHTTP(w, r)
			return
		}

		ses, err := p.sessionStore.Get(r, p.SessionName)
		if ses == nil || err != nil {
			log.Error(err, "failed to get session from store")
			p.redirectToLoginPage(w, r)
			return
		}
		if ses.IsNew {
			log.Info("not authorized")
			p.redirectToLoginPage(w, r)
			return
		}

		sesInfo := session.Get(ses)
		log.Debug().Info("get session", "userID", sesInfo.UserID, "deadline", sesInfo.Deadline)

		// check user ID is owner's
		if sesInfo.UserID != p.User {
			log.Info("invalid authorization", "storedUserID", sesInfo.UserID, "ownerID", p.User)
			p.redirectToLoginPage(w, r)
			return
		}

		// set deadline on request if enabled
		ctx := r.Context()
		if p.MaxAgeSeconds > 0 {
			deadline := time.Unix(sesInfo.Deadline, 0)
			log.DebugAll().Info("set deadline", "at", deadline)

			var cancel context.CancelFunc
			ctx, cancel = context.WithDeadline(ctx, deadline)
			defer cancel()
		}

		log.DebugAll().Info("authorized", "path", r.URL.Path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (p *ProxyServer) redirectToLoginPage(w http.ResponseWriter, r *http.Request) {
	sourcePath := r.URL.RequestURI()

	redirectURL := url.URL{
		Path: p.RedirectPath,
	}
	q := make(url.Values)
	q.Add("redirect_to", sourcePath)
	redirectURL.RawQuery = q.Encode()
	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}
