package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	connect_go "github.com/bufbuild/connect-go"
	webauthnproto "github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/api/errors"

	cosmowebauthn "github.com/cosmo-workspace/cosmo/pkg/auth/webauthn"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

func (s *Server) WebAuthnServiceHandler(mux *http.ServeMux) {
	path, handler := dashboardv1alpha1connect.NewWebAuthnServiceHandler(s,
		connect_go.WithInterceptors(s.validatorInterceptor()))
	mux.Handle(path, s.contextMiddleware(handler))
}

func (s *Server) storeWebAuthnSession(sess *webauthn.SessionData) {
	s.webauthnSessionMap.Store(string(sess.UserID), sess)
}

func (s *Server) getWebAuthnSession(id []byte) (webauthn.SessionData, error) {
	sess, ok := s.webauthnSessionMap.LoadAndDelete(string(id))
	if !ok {
		return webauthn.SessionData{}, fmt.Errorf("session is not found")
	}
	session, ok := sess.(*webauthn.SessionData)
	if !ok {
		panic("session is not *webauthn.SessionData")
	}
	return *session, nil
}

func convertCredentialsToDashCredentials(creds []cosmowebauthn.Credential) []*dashv1alpha1.Credential {
	ret := make([]*dashv1alpha1.Credential, len(creds))
	for i, cred := range creds {
		ret[i] = &dashv1alpha1.Credential{
			Id:          cred.Base64URLEncodedId,
			DisplayName: cred.DisplayName,
			Timestamp:   timestamppb.New(time.Unix(cred.Timestamp, 0)),
		}
	}
	return ret
}

func webauthnErr(log *clog.Logger, err error) {
	if e, ok := err.(*webauthnproto.Error); ok && e != nil {
		log.Error(err, e.DevInfo)
	}
}

// BeginRegistration is proto interface for webauthn.BeginRegistration
func (s *Server) BeginRegistration(ctx context.Context, req *connect_go.Request[dashv1alpha1.BeginRegistrationRequest]) (*connect_go.Response[dashv1alpha1.BeginRegistrationResponse], error) {

	log := clog.FromContext(ctx).WithCaller()

	user, err := cosmowebauthn.GetUser(ctx, s.Klient, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	credCreateOpt, session, err := s.webauthn.BeginRegistration(user)
	if err != nil {
		webauthnErr(log, err)
		return nil, ErrResponse(log, fmt.Errorf("failed at webauthn begin registration: %w", err))
	}

	s.storeWebAuthnSession(session)

	o, err := json.Marshal(credCreateOpt)
	if err != nil {
		return nil, ErrResponse(log, fmt.Errorf("failed to serialize credentil creation options: %w", err))
	}

	return connect_go.NewResponse(&dashv1alpha1.BeginRegistrationResponse{
		CredentialCreationOptions: string(o),
	}), nil
}

// FinishRegistration is proto interface for webauthn.FinishRegistration
func (s *Server) FinishRegistration(ctx context.Context, req *connect_go.Request[dashv1alpha1.FinishRegistrationRequest]) (*connect_go.Response[dashv1alpha1.FinishRegistrationResponse], error) {

	log := clog.FromContext(ctx).WithCaller()

	log.Debug().Info("fetching webauthn user", "user", req.Msg.UserName)
	user, err := cosmowebauthn.GetUser(ctx, s.Klient, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	log.Debug().Info("webauthn parse credential creation response", "user", req.Msg.UserName, "response", req.Msg.CredentialCreationResponse)
	credCreateRes, err := webauthnproto.ParseCredentialCreationResponseBody(strings.NewReader(req.Msg.CredentialCreationResponse))
	if err != nil {
		webauthnErr(log, err)
		return nil, ErrResponse(log, fmt.Errorf("failed to parse credential creation response: %w", err))
	}

	log.Debug().Info("get begin session from store", "user", req.Msg.UserName, "webAuthnID", user.WebAuthnID())
	session, err := s.getWebAuthnSession(user.WebAuthnID())
	if err != nil {
		return nil, ErrResponse(log, fmt.Errorf("failed to get session: %w", err))
	}

	log.Debug().Info("webauthn create credential", "user", req.Msg.UserName)
	cred, err := s.webauthn.CreateCredential(user, session, credCreateRes)
	if err != nil {
		webauthnErr(log, err)
		return nil, ErrResponse(log, fmt.Errorf("failed at webauthn create credential: %w", err))
	}
	log.Info("successfully created and verified credential. saving...", "user", req.Msg.UserName)

	r := requestFromContext(ctx)

	c := cosmowebauthn.Credential{
		DisplayName: r.UserAgent(),
		Timestamp:   time.Now().Unix(),
		Cred:        *cred,
	}
	if err := user.RegisterCredential(ctx, &c); err != nil {
		return nil, ErrResponse(log, fmt.Errorf("failed to save credential: %w", err))
	}
	log.Info("successfully saved credential", "user", req.Msg.UserName)

	return connect_go.NewResponse(&dashv1alpha1.FinishRegistrationResponse{
		Message: "Successfully registered new credential",
	}), nil
}

// BeginLogin is proto interface for webauthn.BeginLogin
func (s *Server) BeginLogin(ctx context.Context, req *connect_go.Request[dashv1alpha1.BeginLoginRequest]) (*connect_go.Response[dashv1alpha1.BeginLoginResponse], error) {

	log := clog.FromContext(ctx).WithCaller()

	log.Debug().Info("fetching webauthn user", "user", req.Msg.UserName)
	user, err := cosmowebauthn.GetUser(ctx, s.Klient, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	log.Debug().Info("webauthn begin login", "user", req.Msg.UserName)
	credAssert, session, err := s.webauthn.BeginLogin(user)
	if err != nil {
		webauthnErr(log, err)
		return nil, ErrResponse(log, errors.NewBadRequest(err.Error()))
	}
	s.storeWebAuthnSession(session)

	a, err := json.Marshal(credAssert)
	if err != nil {
		return nil, ErrResponse(log, fmt.Errorf("failed to serialize credentila creation options: %w", err))
	}

	return connect_go.NewResponse(&dashv1alpha1.BeginLoginResponse{
		CredentialRequestOptions: string(a),
	}), nil
}

// FinishLogin is proto interface for webauthn.FinishLogin
func (s *Server) FinishLogin(ctx context.Context, req *connect_go.Request[dashv1alpha1.FinishLoginRequest]) (*connect_go.Response[dashv1alpha1.FinishLoginResponse], error) {

	log := clog.FromContext(ctx).WithCaller()

	w := responseWriterFromContext(ctx)
	r := requestFromContext(ctx)

	log.Debug().Info("fetching webauthn user", "user", req.Msg.UserName)
	user, err := cosmowebauthn.GetUser(ctx, s.Klient, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	log.Debug().Info("webauthn parse credential request response", "user", req.Msg.UserName, "response", req.Msg.CredentialRequestResult)
	credReqRes, err := webauthnproto.ParseCredentialRequestResponseBody(strings.NewReader(req.Msg.CredentialRequestResult))
	if err != nil {
		webauthnErr(log, err)
		return nil, ErrResponse(log, fmt.Errorf("failed to parse credential creation response: %w", err))
	}

	log.Debug().Info("get begin session from store", "user", req.Msg.UserName, "webAuthnID", user.WebAuthnID())
	sess, err := s.getWebAuthnSession(user.WebAuthnID())
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	log.Debug().Info("webauthn validate login", "user", req.Msg.UserName)
	_, err = s.webauthn.ValidateLogin(user, sess, credReqRes)
	if err != nil {
		webauthnErr(log, err)
		return nil, ErrResponse(log, err)
	}

	// Create session
	sesInfo, expireAt := s.SessionInfo(req.Msg.UserName)
	if err = s.CreateSession(w, r, sesInfo); err != nil {
		log.Error(err, "failed to save session")
		return nil, ErrResponse(log, err)
	}

	return connect_go.NewResponse(&dashv1alpha1.FinishLoginResponse{
		Message:  "Login Success",
		ExpireAt: timestamppb.New(expireAt),
	}), nil
}

// ListCredentials returns all the credentials of a user
func (s *Server) ListCredentials(ctx context.Context, req *connect_go.Request[dashv1alpha1.ListCredentialsRequest]) (*connect_go.Response[dashv1alpha1.ListCredentialsResponse], error) {

	log := clog.FromContext(ctx).WithCaller()

	user, err := cosmowebauthn.GetUser(ctx, s.Klient, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	return connect_go.NewResponse(&dashv1alpha1.ListCredentialsResponse{
		Credentials: convertCredentialsToDashCredentials(user.CredentialList.Creds),
	}), nil
}

// UpdateCredential updates credentia
func (s *Server) UpdateCredential(ctx context.Context, req *connect_go.Request[dashv1alpha1.UpdateCredentialRequest]) (*connect_go.Response[dashv1alpha1.UpdateCredentialResponse], error) {

	log := clog.FromContext(ctx).WithCaller()

	user, err := cosmowebauthn.GetUser(ctx, s.Klient, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	if err := user.UpdateCredential(ctx, req.Msg.CredId, &req.Msg.CredDisplayName); err != nil {
		return nil, ErrResponse(log, err)
	}

	return connect_go.NewResponse(&dashv1alpha1.UpdateCredentialResponse{
		Message: "Successfully updated credential",
	}), nil
}

// DeleteCredential remove credential of a user by given id
func (s *Server) DeleteCredential(ctx context.Context, req *connect_go.Request[dashv1alpha1.DeleteCredentialRequest]) (*connect_go.Response[dashv1alpha1.DeleteCredentialResponse], error) {

	log := clog.FromContext(ctx).WithCaller()

	user, err := cosmowebauthn.GetUser(ctx, s.Klient, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	if err := user.RemoveCredential(ctx, req.Msg.CredId); err != nil {
		return nil, ErrResponse(log, err)
	}

	return connect_go.NewResponse(&dashv1alpha1.DeleteCredentialResponse{
		Message: "Successfully removed credential",
	}), nil
}
