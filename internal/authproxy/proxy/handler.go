package proxy

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
)

func (p *ProxyServer) serveLoginPage() http.Handler {
	return http.StripPrefix(p.RedirectPath, http.FileServer(http.Dir(p.StaticFileDir)))
}

func (p *ProxyServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := p.Log.WithCaller()

	// get data from request body
	req := &dashv1alpha1.LoginRequest{}

	var body io.Reader = r.Body
	defer r.Body.Close()

	err := json.NewDecoder(body).Decode(req)
	if err != nil || req.Id == "" || req.Password == "" {
		log.Info("invalid request", "error", err, "id", req.Id, "passExist", req.Password != "")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// check request user ID is instance owner ID
	if p.User != req.Id {
		log.Info("forbidden request: user ID is not owner ID", "id", req.Id)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// authorize at upstream
	authorized, err := p.authorizer.Authorize(ctx, *req)
	if !authorized || err != nil {
		log.Info("upstream authorization failed", "error", err, "id", req.Id)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ses, _ := p.sessionStore.New(r, p.SessionName)

	sesInfo := session.Info{
		UserID:   req.Id,
		Deadline: time.Now().Add(time.Duration(p.MaxAgeSeconds) * time.Second).Unix(),
	}
	log.Debug().Info("save session", "userID", sesInfo.UserID, "deadline", sesInfo.Deadline)
	ses = session.Set(ses, sesInfo)

	err = p.sessionStore.Save(r, w, ses)
	if err != nil {
		log.Error(err, "failed to save session")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info("successfully logined", "id", req.Id)
	w.WriteHeader(http.StatusOK)
}
