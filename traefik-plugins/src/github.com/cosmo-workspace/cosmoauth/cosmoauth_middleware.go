package cosmoauth

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"strings"
	"time"

	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
	"github.com/gorilla/sessions"
)

// nolint
var (
	LoggerDEBUG = log.New(io.Discard, "cosmo-auth: (DEBUG)", log.Ldate|log.Ltime|log.Lshortfile)
	LoggerINFO  = log.New(io.Discard, "cosmo-auth: (INFO) ", log.Ldate|log.Ltime|log.Lshortfile)
	LoggerERROR = log.New(io.Discard, "cosmo-auth: (ERROR)", log.Ldate|log.Ltime|log.Lshortfile)
)

func accessLog(r *http.Request, statusCode int, ses session.Info, msg string) {
	deadline, _ := time.Parse(time.UnixDate, fmt.Sprint(ses.Deadline))
	userName := r.Header.Get("X-Cosmo-UserName")
	if userName == "" {
		userName = "-"
	}
	LoggerINFO.Printf("%s %s %s %d %s %s: %s", r.RemoteAddr, r.Method, r.URL, statusCode, ses.UserName, deadline.Format(time.RFC3339), msg)
}

// Config the plugin configuration.
type Config struct {
	LogLevel          string `json:"logLevel,omitempty" yaml:"logLevel,omitempty"`
	CookieSessionName string `json:"cookieSessionName,omitempty" yaml:"cookieSessionName,omitempty"`
	CookieDomain      string `json:"cookieDomain,omitempty" yaml:"cookieDomain,omitempty"`
	CookieHashKey     string `json:"cookieHashKey,omitempty" yaml:"cookieHashKey,omitempty"`
	CookieBlockKey    string `json:"cookieBlockKey,omitempty" yaml:"cookieBlockKey,omitempty"`
	SignInUrl         string `json:"signInUrl,omitempty" yaml:"signInUrl,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		LogLevel:          "INFO",
		CookieSessionName: "",
		CookieDomain:      "",
		CookieHashKey:     "----+----X----+----X----+----X----+----X----+----X----+----X----",
		CookieBlockKey:    "----+----X----+----X----+----X--",
		SignInUrl:         "",
	}
}

type CosmoAuth struct {
	config       *Config
	next         http.Handler
	name         string
	RedirectPath string

	SessionStore sessions.Store
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	SetLogger(config.LogLevel)
	LoggerINFO.Printf("Starting %s Middleware...", name)

	conf := &Config{
		LogLevel:          os.ExpandEnv(config.LogLevel),
		CookieSessionName: os.ExpandEnv(config.CookieSessionName),
		CookieDomain:      os.ExpandEnv(config.CookieDomain),
		CookieHashKey:     os.ExpandEnv(config.CookieHashKey),
		CookieBlockKey:    os.ExpandEnv(config.CookieBlockKey),
		SignInUrl:         os.ExpandEnv(config.SignInUrl),
	}

	p := &CosmoAuth{
		config:       conf,
		next:         next,
		name:         name,
		SessionStore: sessions.NewCookieStore([]byte(conf.CookieHashKey), []byte(conf.CookieBlockKey)),
	}

	return p, nil
}

func (p *CosmoAuth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	LoggerDEBUG.Printf("%s %s %s: headers=%v", r.RemoteAddr, r.Method, r.URL, r.Header)

	// Bypass manifest.json not to check session. By default, manifest.json is requested without cookie.
	// https://developer.mozilla.org/en-US/docs/Web/Manifest
	if strings.Contains(strings.ToLower(r.URL.Path), "/manifest.json") {
		p.next.ServeHTTP(w, r)
		return
	}
	ses, err := p.SessionStore.Get(r, p.config.CookieSessionName)
	if err != nil {
		accessLog(r, http.StatusFound, session.Info{}, fmt.Sprintf("failed to get session from store. err: %s", err))
		p.redirectToLoginPage(w, r)
		return
	}
	if ses == nil || ses.IsNew {
		accessLog(r, http.StatusFound, session.Info{}, "no session")
		p.redirectToLoginPage(w, r)
		return
	}

	sesInfo := session.Get(ses)

	// check user name is owner's
	userName := r.Header.Get("X-Cosmo-UserName")
	if userName != "" && sesInfo.UserName != userName {

		// check workspace is shared
		allowed := false
		for header := range r.Header {
			shareUser, found := strings.CutPrefix(strings.ToLower(header), strings.ToLower("X-Cosmo-UserName-"))
			if found && userName != shareUser {
				LoggerDEBUG.Printf("%s %s %s %s: shareUser=%s", r.RemoteAddr, r.Method, r.URL, sesInfo.UserName, shareUser)

				if sesInfo.UserName == shareUser {
					allowed = true
					break
				}
			}
		}
		if !allowed {
			accessLog(r, http.StatusForbidden, sesInfo, "access is denied")
			p.forbidden(w, r)
			return
		}
	}

	// set deadline on request if enabled
	ctx := r.Context()
	if sesInfo.Deadline > 0 {
		deadline := time.Unix(sesInfo.Deadline, 0)

		now := time.Now()
		if deadline.Before(now) {
			accessLog(r, http.StatusFound, sesInfo, "session expired")
			p.redirectToLoginPage(w, r)
			return
		}

		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}

	accessLog(r, http.StatusOK, sesInfo, "access is allowed")
	p.next.ServeHTTP(w, r.WithContext(ctx))
	w.Header().Set("X-Cosmo-UserName", sesInfo.UserName)
}

func (p *CosmoAuth) redirectToLoginPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusFound)
	err := writeRedirectHTML(w, p.config)
	if err != nil {
		LoggerERROR.Printf("failed to write redirect html. err=%s", err)
	}
}

func (p *CosmoAuth) forbidden(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusForbidden)
	writeForbiddenHTML(w)
}

func SetLogger(level string) {
	switch level {
	case "ERROR":
		LoggerERROR.SetOutput(os.Stderr)
	case "INFO":
		LoggerERROR.SetOutput(os.Stderr)
		LoggerINFO.SetOutput(os.Stdout)
	case "DEBUG":
		LoggerERROR.SetOutput(os.Stderr)
		LoggerINFO.SetOutput(os.Stdout)
		LoggerDEBUG.SetOutput(os.Stdout)
	default:
		LoggerERROR.SetOutput(os.Stderr)
		LoggerINFO.SetOutput(os.Stdout)
	}
}
