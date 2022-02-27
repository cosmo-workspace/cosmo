package dashboard

import (
	"context"
	"net/http"
	"sort"

	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/pointer"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/gorilla/mux"
)

func (s *Server) useWorkspaceMiddleWare(router *mux.Router, routes dashv1alpha1.Routes) {
	for _, rtName := range []string{
		"GetWorkspace", "PatchWorkspace", "DeleteWorkspace",
		"PutNetworkRule", "DeleteNetworkRule"} {
		router.Get(rtName).Handler(s.preFetchWorkspaceMiddleware(router.Get(rtName).GetHandler()))
	}
	for _, rt := range routes {
		router.Get(rt.Name).Handler(s.userAuthenticationMiddleware(router.Get(rt.Name).GetHandler()))
		router.Get(rt.Name).Handler(s.preFetchUserMiddleware(router.Get(rt.Name).GetHandler()))
		router.Get(rt.Name).Handler(s.authorizationMiddleware(router.Get(rt.Name).GetHandler()))
	}
}

func (s *Server) PostWorkspace(ctx context.Context, userId string, req dashv1alpha1.CreateWorkspaceRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req, "userId", userId)

	cfg, err := s.Klient.GetWorkspaceConfig(ctx, req.Template)
	if err != nil {
		log.Error(err, "failed to get workspace config from template", "template", req.Template)
		return ErrorResponse(http.StatusBadRequest, "failed to get workspace config in template")
	}

	ws := &wsv1alpha1.Workspace{}
	ws.SetName(req.Name)
	ws.SetNamespace(wsv1alpha1.UserNamespace(userId))
	ws.Spec = wsv1alpha1.WorkspaceSpec{
		Template: cosmov1alpha1.TemplateRef{
			Name: req.Template,
		},
		Vars: req.Vars,
	}

	if err := s.Klient.Create(ctx, ws); err != nil {
		if apierrs.IsAlreadyExists(err) {
			return ErrorResponse(http.StatusTooManyRequests, "Workspace already exists")

		} else {
			message := "failed to create workspace"
			log.Error(err, message, "userid", userId, "workspace", req.Name, "template", req.Template)
			return ErrorResponse(http.StatusInternalServerError, message)
		}
	}

	ws.Status.Phase = "Pending"
	ws.Status.Config = cfg

	res := &dashv1alpha1.CreateWorkspaceResponse{}
	res.Workspace = convertWorkspaceTodashv1alpha1Workspace(*ws)
	res.Message = "Successfully created"
	log.Info(res.Message, "userid", userId, "workspace", req.Name, "template", req.Template)
	return NormalResponse(http.StatusCreated, res)
}

func (s *Server) GetWorkspaces(ctx context.Context, userId string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId)

	wss, err := s.Klient.ListWorkspacesByUserID(ctx, userId)
	if err != nil {
		message := "failed to list workspaces"
		log.Error(err, message, "userid", userId)
		return ErrorResponse(http.StatusInternalServerError, message)
	}

	apiwss := make([]dashv1alpha1.Workspace, len(wss))
	for i, v := range wss {
		apiwss[i] = *convertWorkspaceTodashv1alpha1Workspace(v)
	}

	res := &dashv1alpha1.ListWorkspaceResponse{}
	res.Items = apiwss
	sort.Slice(res.Items, func(i, j int) bool { return res.Items[i].Name < res.Items[j].Name })

	if len(res.Items) == 0 {
		res.Message = "No items found"
	}
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) GetWorkspace(ctx context.Context, userId string, workspaceName string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "workspaceName", workspaceName)

	ws := workspaceFromContext(ctx)
	if ws == nil {
		log.Info("workspace not found in context")
		return ErrorResponse(http.StatusInternalServerError, "")
	}

	res := &dashv1alpha1.GetWorkspaceResponse{}
	res.Workspace = convertWorkspaceTodashv1alpha1Workspace(*ws)

	return NormalResponse(http.StatusOK, res)
}

func (s *Server) DeleteWorkspace(ctx context.Context, userId string, workspaceName string) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "workspaceName", workspaceName)

	ws := workspaceFromContext(ctx)
	if ws == nil {
		log.Info("workspace not found in context")
		return ErrorResponse(http.StatusInternalServerError, "")
	}

	res := &dashv1alpha1.DeleteWorkspaceResponse{}

	err := s.Klient.Delete(ctx, ws)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return ErrorResponse(http.StatusNotFound, err.Error())
		} else {
			res.Message = "failed to delete workspace"
			log.Error(err, res.Message, "userid", userId, "workspace", ws.Name)
			return ErrorResponse(http.StatusInternalServerError, "")
		}
	}

	res.Workspace = convertWorkspaceTodashv1alpha1Workspace(*ws)

	res.Message = "Successfully deleted"
	log.Info(res.Message, "userid", userId, "workspace", ws.Name)
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) PatchWorkspace(ctx context.Context, userId string, workspaceName string, req dashv1alpha1.PatchWorkspaceRequest) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "userId", userId, "workspaceName", workspaceName, "req", req)

	ws := workspaceFromContext(ctx)
	if ws == nil {
		log.Info("workspace not found in context")
		return ErrorResponse(http.StatusInternalServerError, "")
	}

	res := &dashv1alpha1.PatchWorkspaceResponse{}

	before := ws.DeepCopy()

	if req.Replicas != nil {
		ws.Spec.Replicas = req.Replicas
	}

	if !equality.Semantic.DeepEqual(before, ws) {
		err := s.Klient.Update(ctx, ws)
		if err != nil {
			if apierrs.IsNotFound(err) {
				message := err.Error()
				log.Error(err, message, "userid", userId, "workspace", ws.Name)
				return ErrorResponse(http.StatusInternalServerError, message)
			} else {
				message := "failed to update workspace"
				log.Error(err, message, "userid", userId, "workspace", ws.Name)
				return ErrorResponse(http.StatusInternalServerError, message)
			}
		}
		res.Message = "Successfully updated"
	} else {
		res.Message = "No change"
	}

	res.Workspace = convertWorkspaceTodashv1alpha1Workspace(*ws)

	log.Info(res.Message, "userid", userId, "workspace", ws.Name)
	return NormalResponse(http.StatusOK, res)
}

func convertWorkspaceTodashv1alpha1Workspace(ws wsv1alpha1.Workspace) *dashv1alpha1.Workspace {
	replicas := ws.Spec.Replicas
	if replicas == nil {
		replicas = pointer.Int64(0)
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
