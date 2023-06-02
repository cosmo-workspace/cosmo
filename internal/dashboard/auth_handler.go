package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"time"

	connect_go "github.com/bufbuild/connect-go"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

func (s *Server) AuthServiceHandler(mux *http.ServeMux) {
	path, handler := dashboardv1alpha1connect.NewAuthServiceHandler(s)
	mux.Handle(path, s.contextMiddleware(handler))
}

func (s *Server) Verify(ctx context.Context, req *connect_go.Request[emptypb.Empty]) (*connect_go.Response[dashv1alpha1.VerifyResponse], error) {

	log := clog.FromContext(ctx).WithCaller()

	loginUser, deadline, err := s.verifyAndGetLoginUser(ctx)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	return connect_go.NewResponse(&dashv1alpha1.VerifyResponse{
		UserName:              loginUser.Name,
		ExpireAt:              timestamppb.New(deadline),
		RequirePasswordUpdate: false,
	}), nil
}

func (s *Server) Login(ctx context.Context, req *connect_go.Request[dashv1alpha1.LoginRequest]) (*connect_go.Response[dashv1alpha1.LoginResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "username", req.Msg.UserName)

	w := responseWriterFromContext(ctx)
	r := requestFromContext(ctx)

	// Check name
	user, err := s.Klient.GetUser(ctx, req.Msg.UserName)
	if err != nil {
		log.Info(err.Error(), "username", req.Msg.UserName)
		return nil, ErrResponse(log, kosmo.NewForbiddenError("incorrect user or password", nil))
	}
	// Check password
	authrizer, ok := s.Authorizers[user.Spec.AuthType]
	if !ok {
		log.Info("authrizer not found", "username", req.Msg.UserName, "authType", user.Spec.AuthType)
		return nil, ErrResponse(log, kosmo.NewServiceUnavailableError(
			fmt.Sprintf("auth-type '%s' is not supported", user.Spec.AuthType), nil))
	}
	verified, err := authrizer.Authorize(ctx, req.Msg)
	if err != nil {
		log.Error(err, "authorize failed", "username", req.Msg.UserName)
		return nil, ErrResponse(log, kosmo.NewForbiddenError("incorrect user or password", nil))

	}
	if !verified {
		log.Info("login failed: password invalid", "username", req.Msg.UserName)
		return nil, ErrResponse(log, kosmo.NewForbiddenError("incorrect user or password", nil))
	}
	var isDefault bool
	if cosmov1alpha1.UserAuthType(user.Spec.AuthType) == cosmov1alpha1.UserAuthTypePasswordSecert {
		isDefault, err = s.Klient.IsDefaultPassword(ctx, req.Msg.UserName)
		if err != nil {
			log.Error(err, "failed to check is default password", "username", req.Msg.UserName)
			return nil, ErrResponse(log, kosmo.NewInternalServerError("", nil))
		}
	}

	// Create session
	now := time.Now()
	expireAt := now.Add(time.Duration(s.MaxAgeSeconds) * time.Second)

	ses, _ := s.sessionStore.New(r, s.CookieSessionName)
	sesInfo := session.Info{
		UserName: req.Msg.UserName,
		Deadline: expireAt.Unix(),
	}
	log.DebugAll().Info("save session", "userName", sesInfo.UserName, "deadline", sesInfo.Deadline)
	ses = session.Set(ses, sesInfo)

	err = s.sessionStore.Save(r, w, ses)
	if err != nil {
		log.Error(err, "failed to save session")
		return nil, ErrResponse(log, err)
	}

	return connect_go.NewResponse(&dashv1alpha1.LoginResponse{
		UserName:              req.Msg.UserName,
		ExpireAt:              timestamppb.New(expireAt),
		RequirePasswordUpdate: isDefault,
	}), nil
}

func (s *Server) Logout(ctx context.Context, req *connect_go.Request[emptypb.Empty]) (*connect_go.Response[emptypb.Empty], error) {

	log := clog.FromContext(ctx).WithCaller()

	_, _, err := s.verifyAndGetLoginUser(ctx)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	// clear session
	cookie := s.sessionCookieKey()
	cookie.MaxAge = -1
	w := responseWriterFromContext(ctx)
	http.SetCookie(w, cookie)

	resp := connect_go.NewResponse(&emptypb.Empty{})

	return resp, nil
}
