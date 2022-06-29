package dashboard

import (
	"context"
	"net/http"

	"k8s.io/utils/pointer"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/gorilla/mux"
)

func (s *Server) useWorkspaceMiddleWare(router *mux.Router, routes dashv1alpha1.Routes) {
	for _, rt := range routes {
		router.Get(rt.Name).Handler(
			s.authorizationMiddleware(
				s.userAuthenticationMiddleware(
					router.Get(rt.Name).GetHandler())))
	}
}

func (s *Server) PostWorkspace(ctx context.Context, userId string, req dashv1alpha1.CreateWorkspaceRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req, "userId", userId)

	ws, err := s.Klient.CreateWorkspace(ctx, userId, req.Name, req.Template, req.Vars)
	if err != nil {
		return ErrorResponse(log, err)
	}

	res := &dashv1alpha1.CreateWorkspaceResponse{
		Message:   "Successfully created",
		Workspace: convertWorkspaceTodashv1alpha1Workspace(*ws),
	}
	log.Info(res.Message, "userid", userId, "workspace", req.Name, "template", req.Template)
	return NormalResponse(http.StatusCreated, res)
}

func (s *Server) GetWorkspaces(ctx context.Context, userId string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId)

	wss, err := s.Klient.ListWorkspacesByUserID(ctx, userId)
	if err != nil {
		return ErrorResponse(log, err)
	}

	apiwss := make([]dashv1alpha1.Workspace, len(wss))
	for i, v := range wss {
		apiwss[i] = *convertWorkspaceTodashv1alpha1Workspace(v)
	}
	res := &dashv1alpha1.ListWorkspaceResponse{
		Items: apiwss,
	}
	if len(res.Items) == 0 {
		res.Message = "No items found"
	}
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) GetWorkspace(ctx context.Context, userId string, workspaceName string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "workspaceName", workspaceName)

	ws, err := s.Klient.GetWorkspaceByUserID(ctx, workspaceName, userId)
	if err != nil {
		return ErrorResponse(log, err)
	}

	res := &dashv1alpha1.GetWorkspaceResponse{
		Workspace: convertWorkspaceTodashv1alpha1Workspace(*ws),
	}
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) DeleteWorkspace(ctx context.Context, userId string, workspaceName string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "workspaceName", workspaceName)

	ws, err := s.Klient.DeleteWorkspace(ctx, workspaceName, userId)
	if err != nil {
		return ErrorResponse(log, err)
	}

	res := &dashv1alpha1.DeleteWorkspaceResponse{
		Message:   "Successfully deleted",
		Workspace: convertWorkspaceTodashv1alpha1Workspace(*ws),
	}
	log.Info(res.Message, "userid", userId, "workspace", ws.Name)
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) PatchWorkspace(ctx context.Context, userId, workspaceName string, req dashv1alpha1.PatchWorkspaceRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "workspaceName", workspaceName, "req", req)

	ws, err := s.Klient.UpdateWorkspace(ctx, workspaceName, userId, kosmo.UpdateWorkspaceOpts{Replicas: req.Replicas})
	if err != nil {
		return ErrorResponse(log, err)
	}

	res := &dashv1alpha1.PatchWorkspaceResponse{
		Message:   "Successfully updated",
		Workspace: convertWorkspaceTodashv1alpha1Workspace(*ws),
	}
	log.Info(res.Message, "userid", userId, "workspace", ws.Name)
	return NormalResponse(http.StatusOK, res)
}

func convertWorkspaceTodashv1alpha1Workspace(ws wsv1alpha1.Workspace) *dashv1alpha1.Workspace {
	replicas := ws.Spec.Replicas
	if replicas == nil {
		replicas = pointer.Int64(1)
	}
	return &dashv1alpha1.Workspace{
		Name:    ws.Name,
		OwnerID: wsv1alpha1.UserIDByNamespace(ws.Namespace),
		Spec: dashv1alpha1.WorkspaceSpec{
			Template:          ws.Spec.Template.Name,
			Replicas:          *replicas,
			Vars:              ws.Spec.Vars,
			AdditionalNetwork: convertNetRulesTodashv1alpha1NetRules(ws.Spec.Network, ws.Status.URLs, ws.Status.Config.ServiceMainPortName),
		},
		Status: dashv1alpha1.WorkspaceStatus{
			Phase:   string(ws.Status.Phase),
			MainUrl: ws.Status.URLs[ws.Status.Config.ServiceMainPortName],
			UrlBase: ws.Status.Config.URLBase,
		},
	}
}
