package proxy

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"
	authv1alpha1 "github.com/cosmo-workspace/cosmo/api/auth-proxy/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
)

type sessionCtxKey int

const sesCtxKey sessionCtxKey = iota

func (p *ProxyServer) serveLoginPage() http.Handler {
	return http.StripPrefix(p.RedirectPath, http.FileServer(http.Dir(p.StaticFileDir)))
}

func (p *ProxyServer) loginCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := p.Log.WithCaller()

		saveSessionInfo := func(sesInfo *session.Info) {
			ses, _ := p.sessionStore.New(r, p.SessionName)
			ses = session.Set(ses, *sesInfo)
			err := p.sessionStore.Save(r, w, ses)
			if err != nil {
				log.Error(err, "failed to save session")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		ctx := context.WithValue(r.Context(), sesCtxKey, saveSessionInfo)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (p *ProxyServer) Login(ctx context.Context, req *connect.Request[authv1alpha1.LoginRequest]) (*connect.Response[authv1alpha1.Empty], error) {
	log := p.Log.WithCaller()
	log.Info("Login", "id", req.Msg.Id, "passExist", req.Msg.Password != "")

	res := connect.NewResponse(&authv1alpha1.Empty{})
	res.Header().Set("AuthProxy-Version", "v1alpha1")

	// check args
	if req.Msg.Id == "" || req.Msg.Password == "" {
		log.Info("invalid request", "id", req.Msg.Id, "passExist", req.Msg.Password != "")
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid request"))
	}

	// check request user ID is instance owner ID
	if p.User != req.Msg.Id {
		log.Info("forbidden request: user ID is not owner ID", "id", req.Msg.Id)
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user ID is not owner ID"))
	}

	// authorize at upstream
	authorized, err := p.authorizer.Authorize(ctx, req.Msg)
	if !authorized || err != nil {
		log.Info("upstream authorization failed", "error", err, "id", req.Msg.Id)
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("upstream authorization failed"))
	}

	saveSessionInfo := ctx.Value(sesCtxKey).(func(sesInfo *session.Info))
	saveSessionInfo(&session.Info{
		UserID:   req.Msg.Id,
		Deadline: time.Now().Add(time.Duration(p.MaxAgeSeconds) * time.Second).Unix(),
	})

	log.Info("successfully logined", "id", req.Msg.Id, "passExist", req.Msg.Password != "")
	return res, nil
}
