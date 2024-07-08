package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	connect_go "github.com/bufbuild/connect-go"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
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
		targetUser, err := s.Klient.GetUser(ctx, req.Msg.UserName)
		if err != nil {
			return nil, ErrResponse(log, err)
		}

		// group-admin user can delete users which have only the their groups
		if err := adminAuthentication(ctx, validateCallerHasAdminForAllRoles(targetUser.Spec.Roles)); err != nil {
			return nil, ErrResponse(log, err)
		}
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
		targetUser, err := s.Klient.GetUser(ctx, req.Msg.UserName)
		if err != nil {
			return nil, ErrResponse(log, err)
		}

		// group-admin user can get workspaces of users which have only the their groups
		if err := adminAuthentication(ctx, validateCallerHasAdminForAtLeastOneRole(targetUser.Spec.Roles)); err != nil {
			return nil, ErrResponse(log, err)
		}
	}

	wss, err := s.Klient.ListWorkspacesByUserName(ctx, req.Msg.UserName, func(opt *kosmo.ListWorkspacesOptions) {
		opt.IncludeShared = req.Msg.IncludeShared != nil && *req.Msg.IncludeShared
	})
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.GetWorkspacesResponse{}
	if req.Msg.WithRaw != nil && *req.Msg.WithRaw {
		res.Items = apiconv.C2D_Workspaces(wss, apiconv.WithWorkspaceRaw())
	} else {
		res.Items = apiconv.C2D_Workspaces(wss)
	}
	if len(res.Items) == 0 {
		res.Message = "No items found"
	}
	return connect_go.NewResponse(res), nil
}

func (s *Server) GetWorkspace(ctx context.Context, req *connect_go.Request[dashv1alpha1.GetWorkspaceRequest]) (*connect_go.Response[dashv1alpha1.GetWorkspaceResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	if err := s.sharedWorkspaceAuthorization(ctx, req.Msg.WsName, req.Msg.UserName, false); err != nil {
		targetUser, err := s.Klient.GetUser(ctx, req.Msg.UserName)
		if err != nil {
			return nil, ErrResponse(log, err)
		}

		// group-admin user can get workspaces of users which have only the their groups
		if err := adminAuthentication(ctx, validateCallerHasAdminForAtLeastOneRole(targetUser.Spec.Roles)); err != nil {
			return nil, ErrResponse(log, err)
		}
	}

	ws, err := s.Klient.GetWorkspaceByUserName(ctx, req.Msg.WsName, req.Msg.UserName)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	res := &dashv1alpha1.GetWorkspaceResponse{}
	if req.Msg.WithRaw != nil && *req.Msg.WithRaw {
		var inst cosmov1alpha1.Instance
		s.Klient.Get(ctx, types.NamespacedName{Name: ws.Status.Instance.Name, Namespace: ws.Status.Instance.Namespace}, &inst)
		res.Workspace = apiconv.C2D_Workspace(*ws, apiconv.WithWorkspaceRaw(), apiconv.WithWorkspaceInstanceRaw(&inst))

	} else {
		res.Workspace = apiconv.C2D_Workspace(*ws)
	}

	return connect_go.NewResponse(res), nil
}

func (s *Server) DeleteWorkspace(ctx context.Context, req *connect_go.Request[dashv1alpha1.DeleteWorkspaceRequest]) (*connect_go.Response[dashv1alpha1.DeleteWorkspaceResponse], error) {
	log := clog.FromContext(ctx).WithCaller()
	log.Debug().Info("request", "req", req)

	if err := userAuthentication(ctx, req.Msg.UserName); err != nil {
		targetUser, err := s.Klient.GetUser(ctx, req.Msg.UserName)
		if err != nil {
			return nil, ErrResponse(log, err)
		}

		// group-admin user can delete users which have only the their groups
		if err := adminAuthentication(ctx, validateCallerHasAdminForAllRoles(targetUser.Spec.Roles)); err != nil {
			return nil, ErrResponse(log, err)
		}
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

	if err := s.sharedWorkspaceAuthorization(ctx, req.Msg.WsName, req.Msg.UserName, true); err != nil {
		targetUser, err := s.Klient.GetUser(ctx, req.Msg.UserName)
		if err != nil {
			return nil, ErrResponse(log, err)
		}

		// group-admin user can delete users which have only the their groups
		if err := adminAuthentication(ctx, validateCallerHasAdminForAtLeastOneRole(targetUser.Spec.Roles)); err != nil {
			return nil, ErrResponse(log, err)
		}
	}

	var delPolicy *string
	if req.Msg.DeletePolicy != nil {
		delPolicy = ptr.To(apiconv.D2C_DeletePolicy(req.Msg.DeletePolicy))
	}

	ws, err := s.Klient.UpdateWorkspace(ctx, req.Msg.WsName, req.Msg.UserName, kosmo.UpdateWorkspaceOpts{
		Replicas:     req.Msg.Replicas,
		Vars:         req.Msg.Vars,
		DeletePolicy: delPolicy,
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

func (s *Server) sharedWorkspaceAuthorization(ctx context.Context, wsName, wsOwnerName string, update bool) error {
	log := clog.FromContext(ctx).WithCaller()

	if err := userAuthentication(ctx, wsOwnerName); err == nil {
		// pass if caller is the owner of the workspace
		return nil
	}

	caller := callerFromContext(ctx)
	if !slices.ContainsFunc(caller.Status.SharedWorkspaces, func(sharedRef cosmov1alpha1.ObjectRef) bool {
		return cosmov1alpha1.UserNameByNamespace(sharedRef.Namespace) == wsOwnerName
	}) {
		return NewForbidden(fmt.Errorf("invalid user authentication"))
	}

	// only users who are allowed to access main rule can update workspace
	ws, err := s.Klient.GetWorkspaceByUserName(ctx, wsName, wsOwnerName)
	if err != nil {
		return ErrResponse(log, err)
	}

	if !slices.ContainsFunc(ws.Spec.Network, func(r cosmov1alpha1.NetworkRule) bool {
		return slices.Contains(r.AllowedUsers, caller.Name)
	}) {
		return NewForbidden(fmt.Errorf("invalid user authentication"))
	}

	if update {
		mainRuleIndex := slices.IndexFunc(ws.Spec.Network, func(n cosmov1alpha1.NetworkRule) bool {
			return n.CustomHostPrefix == ws.Status.Config.ServiceMainPortName
		})
		if mainRuleIndex < 0 {
			return ErrResponse(log, ErrResponse(log, apierrs.NewInternalError(fmt.Errorf("main rule not found"))))
		}

		// check caller is allowed to access main rule
		if !slices.Contains(ws.Spec.Network[mainRuleIndex].AllowedUsers, caller.Name) {
			return NewForbidden(fmt.Errorf("invalid user authentication"))
		}
	}

	return nil
}
