package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"time"

	connect_go "github.com/bufbuild/connect-go"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	apierrs "k8s.io/apimachinery/pkg/api/errors"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/auth/session"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

func (s *Server) AuthServiceHandler(mux *http.ServeMux) {
	path, handler := dashboardv1alpha1connect.NewAuthServiceHandler(s)
	mux.Handle(path, s.contextMiddleware(handler))
}

func (s *Server) CreateSession(w http.ResponseWriter, r *http.Request, sesInfo session.Info) error {
	// Create session
	ses, _ := s.sessionStore.New(r, s.CookieSessionName)
	ses = session.Set(ses, sesInfo)

	err := s.sessionStore.Save(r, w, ses)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}

func (s *Server) SessionInfo(userName string) (session.Info, time.Time) {
	now := time.Now()
	expireAt := now.Add(time.Duration(s.MaxAgeSeconds) * time.Second)
	return session.Info{
		UserName: userName,
		Deadline: expireAt.Unix(),
	}, expireAt
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
		return nil, ErrResponse(log, NewForbidden(fmt.Errorf("incorrect user or password")))
	}
	// Check password
	authrizer, ok := s.Authorizers[user.Spec.AuthType]
	if !ok {
		log.Info("authrizer not found", "username", req.Msg.UserName, "authType", user.Spec.AuthType)
		return nil, ErrResponse(log, apierrs.NewServiceUnavailable(
			fmt.Sprintf("auth-type '%s' is not supported", user.Spec.AuthType)))
	}
	verified, err := authrizer.Authorize(ctx, req.Msg)
	if err != nil {
		log.Error(err, "authorize failed", "username", req.Msg.UserName)
		return nil, ErrResponse(log, NewForbidden(fmt.Errorf("incorrect user or password")))

	}
	if !verified {
		log.Info("login failed: password invalid", "username", req.Msg.UserName)
		return nil, ErrResponse(log, NewForbidden(fmt.Errorf("incorrect user or password")))
	}
	var isDefault bool
	if cosmov1alpha1.UserAuthType(user.Spec.AuthType) == cosmov1alpha1.UserAuthTypePasswordSecert {
		isDefault, err = s.Klient.IsDefaultPassword(ctx, req.Msg.UserName)
		if err != nil {
			log.Error(err, "failed to check is default password", "username", req.Msg.UserName)
			return nil, ErrResponse(log, apierrs.NewInternalError(fmt.Errorf("failed to check is default password: %w", err)))
		}
	}

	// Create session
	sesInfo, expireAt := s.SessionInfo(req.Msg.UserName)
	if err = s.CreateSession(w, r, sesInfo); err != nil {
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
