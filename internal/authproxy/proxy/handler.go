package proxy

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
	authv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/auth-proxy/v1alpha1"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

type ctxKeySession struct{}

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
		ctx := context.WithValue(r.Context(), ctxKeySession{}, saveSessionInfo)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (p *ProxyServer) Login(ctx context.Context, req *connect.Request[authv1alpha1.LoginRequest]) (*connect.Response[emptypb.Empty], error) {
	log := p.Log.WithCaller()
	log.Info("Login", "userName", req.Msg.UserName, "passExist", req.Msg.Password != "")

	res := connect.NewResponse(&emptypb.Empty{})
	res.Header().Set("AuthProxy-Version", "v1alpha1")

	// check args
	if req.Msg.UserName == "" || req.Msg.Password == "" {
		log.Info("invalid request", "username", req.Msg.UserName, "passExist", req.Msg.Password != "")
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid request"))
	}

	// check request user name is instance owner name
	if p.User != req.Msg.UserName {
		log.Info("forbidden request: user name is not owner name", "userName", req.Msg.UserName)
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("user name is not owner name"))
	}

	// authorize at upstream
	authorized, err := p.authorizer.Authorize(ctx, req.Msg)
	if !authorized || err != nil {
		log.Info("upstream authorization failed", "error", err, "userName", req.Msg.UserName)
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("upstream authorization failed"))
	}

	saveSessionInfo := ctx.Value(ctxKeySession{}).(func(sesInfo *session.Info))
	saveSessionInfo(&session.Info{
		UserName: req.Msg.UserName,
		Deadline: time.Now().Add(time.Duration(p.MaxAgeSeconds) * time.Second).Unix(),
	})

	log.Info("successfully logined", "username", req.Msg.UserName, "passExist", req.Msg.Password != "")
	return res, nil
}
