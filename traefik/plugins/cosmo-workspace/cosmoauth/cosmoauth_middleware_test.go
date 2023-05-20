package cosmoauth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
	"github.com/cosmo-workspace/cosmoauth"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/gorilla/sessions"
)

func TestCreateConfig(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "✅ OK",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cosmoauth.CreateConfig()
			snaps.MatchSnapshot(t, got)
		})
	}
}

func TestNew(t *testing.T) {

	tests := []struct {
		name   string
		config *cosmoauth.Config
	}{
		{
			name: "✅ OK",
			config: &cosmoauth.Config{
				LogLevel:          "DEBUG",
				CookieSessionName: "sessionName",
				CookieDomain:      "domain.com",
				CookieHashKey:     "1234567890",
				CookieBlockKey:    "abcdefghij",
				SignInUrl:         "https://xxxx.domain.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cosmoauth.New(
				context.Background(),
				http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
				tt.config,
				"auth",
			)
			snaps.MatchSnapshot(t, got)
			snaps.MatchSnapshot(t, err)
		})
	}
}

func TestCosmoAuth_ServeHTTP(t *testing.T) {

	tests := []struct {
		name             string
		url              string
		hasSession       bool
		hasInvaidSession bool
		header           *map[string]string
	}{
		{
			name: "✅ manifest.json",
			url:  "http://localhost/manifest.json",
		},
		{
			name: "❌ cookie empty",
			url:  "http://localhost",
		},
		{
			name:             "❌ invalid cookie",
			url:              "http://localhost",
			hasInvaidSession: true,
		},
		{
			name:       "✅ valid session",
			url:        "http://localhost",
			hasSession: true,
		},
		{
			name:       "✅ valid user",
			url:        "http://localhost",
			hasSession: true,
			header: &map[string]string{
				"X-Cosmo-UserName": "user1",
			},
		},
		{
			name:       "❌ invalid user",
			url:        "http://localhost",
			hasSession: true,
			header: &map[string]string{
				"X-Cosmo-UserName": "userxxx",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
			cfg := &cosmoauth.Config{
				LogLevel:          "DEBUG",
				CookieSessionName: "sessionName",
				CookieDomain:      "domain.com",
				CookieHashKey:     "12345678901234567890123456789012",
				CookieBlockKey:    "abcdefghijklmnopqrstuABCDEFGHIJK",
				SignInUrl:         "https://xxxx.domain.com",
			}
			handler, err := cosmoauth.New(context.Background(), next, cfg, "cosmo-auth-middleware")
			if err != nil {
				snaps.MatchSnapshot(t, err)
				return
			}
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, tt.url, nil)
			res := httptest.NewRecorder()
			if tt.hasSession {
				store := sessions.NewCookieStore([]byte(cfg.CookieHashKey), []byte(cfg.CookieBlockKey))
				sesInfo := session.Info{
					UserName: "user1",
					Deadline: time.Date(2999, 4, 1, 0, 0, 0, 0, time.Local).Unix(),
				}
				tempRes := httptest.NewRecorder()
				ses, _ := store.New(req, cfg.CookieSessionName)
				ses = session.Set(ses, sesInfo)
				ses.Save(req, tempRes)
				req.Header.Set("Cookie", tempRes.Header().Get("Set-Cookie"))
			} else if tt.hasInvaidSession {
				req.Header.Set("Cookie", cfg.CookieSessionName+"=xxxxxxxxxx")
			}
			if tt.header != nil {
				for key, value := range *tt.header {
					req.Header.Add(key, value)
				}
			}
			// test target
			handler.ServeHTTP(res, req)

			snaps.MatchSnapshot(t, res)
			snaps.MatchSnapshot(t, res.Body.String())
		})
	}
}
