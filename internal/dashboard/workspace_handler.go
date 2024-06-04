package dashboard

import (
	"context"
	"net/http"

	connect_go "github.com/bufbuild/connect-go"

	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

func (s *Server) WorkspaceServiceHandler(mux *http.ServeMux) {
	path, handler := dashboardv1alpha1connect.NewWorkspaceServiceHandler(s,
		connect_go.WithInterceptors(authorizationInterceptorFunc(s.verifyAndGetLoginUser)),
		connect_go.WithInterceptors(s.validatorInterceptor()),
	)
	mux.Handle(path, s.timeoutHandler(s.contextMiddleware(handler)))
}

func (s *Server) CreateWorkspace(ctx context.Context, req *connect_go.Request[dashv1alpha1.CreateWorkspaceRequest]) (*connect_go.Response[dashv1alpha1.CreateWorkspaceResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	if err := userAuthentication(ctx, req.Msg.UserName); err != nil {
		return nil, ErrResponse(log, err)
	}

	m := req.Msg
	ws, err := s.Klient.CreateWorkspace(ctx, m.UserName, m.WsName, m.Template, m.Vars)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.CreateWorkspaceResponse{
		Message:   "Successfully created",
		Workspace: apiconv.C2D_Workspace(*ws),
	}
	log.Info(res.Message, "username", m.UserName, "workspace", m.WsName, "template", m.Template)
	return connect_go.NewResponse(res), nil
}

func (s *Server) GetWorkspaces(ctx context.Context, req *connect_go.Request[dashv1alpha1.GetWorkspacesRequest]) (*connect_go.Response[dashv1alpha1.GetWorkspacesResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	if err := userAuthentication(ctx, req.Msg.UserName); err != nil {
		return nil, ErrResponse(log, err)
	}

	wss, err := s.Klient.ListWorkspacesByUserName(ctx, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.GetWorkspacesResponse{
		Items: apiconv.C2D_Workspaces(wss, apiconv.WithWorkspaceRaw(req.Msg.WithRaw)),
	}
	if len(res.Items) == 0 {
		res.Message = "No items found"
	}
	return connect_go.NewResponse(res), nil
}

func (s *Server) GetWorkspace(ctx context.Context, req *connect_go.Request[dashv1alpha1.GetWorkspaceRequest]) (*connect_go.Response[dashv1alpha1.GetWorkspaceResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	if err := userAuthentication(ctx, req.Msg.UserName); err != nil {
		return nil, ErrResponse(log, err)
	}

	ws, err := s.Klient.GetWorkspaceByUserName(ctx, req.Msg.WsName, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.GetWorkspaceResponse{
		Workspace: apiconv.C2D_Workspace(*ws, apiconv.WithWorkspaceRaw(req.Msg.WithRaw)),
	}

	return connect_go.NewResponse(res), nil
}

func (s *Server) DeleteWorkspace(ctx context.Context, req *connect_go.Request[dashv1alpha1.DeleteWorkspaceRequest]) (*connect_go.Response[dashv1alpha1.DeleteWorkspaceResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	if err := userAuthentication(ctx, req.Msg.UserName); err != nil {
		return nil, ErrResponse(log, err)
	}

	ws, err := s.Klient.DeleteWorkspace(ctx, req.Msg.WsName, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.DeleteWorkspaceResponse{
		Message:   "Successfully deleted",
		Workspace: apiconv.C2D_Workspace(*ws),
	}
	log.Info(res.Message, "username", req.Msg.UserName, "workspaceName", req.Msg.WsName)
	return connect_go.NewResponse(res), nil
}

func (s *Server) UpdateWorkspace(ctx context.Context, req *connect_go.Request[dashv1alpha1.UpdateWorkspaceRequest]) (*connect_go.Response[dashv1alpha1.UpdateWorkspaceResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	if err := userAuthentication(ctx, req.Msg.UserName); err != nil {
		return nil, ErrResponse(log, err)
	}

	ws, err := s.Klient.UpdateWorkspace(ctx, req.Msg.WsName, req.Msg.UserName, kosmo.UpdateWorkspaceOpts{
		Replicas: req.Msg.Replicas,
		Vars:     req.Msg.Vars,
	})
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.UpdateWorkspaceResponse{
		Message:   "Successfully updated",
		Workspace: apiconv.C2D_Workspace(*ws),
	}
	log.Info(res.Message, "username", req.Msg.UserName, "workspaceName", req.Msg.WsName)
	return connect_go.NewResponse(res), nil
}
