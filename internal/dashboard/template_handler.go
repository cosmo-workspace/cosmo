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

func (s *Server) TemplateServiceHandler(mux *http.ServeMux) {
	path, handler := dashboardv1alpha1connect.NewTemplateServiceHandler(s,
		connect_go.WithInterceptors(authorizationInterceptorFunc(s.verifyAndGetLoginUser)),
		connect_go.WithInterceptors(s.validatorInterceptor()),
	)
	mux.Handle(path, s.timeoutHandler(s.contextMiddleware(handler)))
}

func (s *Server) GetWorkspaceTemplates(ctx context.Context, req *connect_go.Request[dashv1alpha1.GetWorkspaceTemplatesRequest]) (*connect_go.Response[dashv1alpha1.GetWorkspaceTemplatesResponse], error) {
	log := clog.FromContext(ctx).WithCaller()

	user := callerFromContext(ctx)

	tmpls, err := s.Klient.ListWorkspaceTemplates(ctx)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	if req.Msg.UseRoleFilter != nil && *req.Msg.UseRoleFilter {
		tmpls = kosmo.FilterTemplates(ctx, tmpls, user)
	}

	res := &dashv1alpha1.GetWorkspaceTemplatesResponse{
		Items: apiconv.C2D_Templates(tmpls, apiconv.WithTemplateRaw(req.Msg.WithRaw)),
	}

	if len(res.Items) == 0 {
		res.Message = "No items found"
	}

	return connect_go.NewResponse(res), nil
}

func (s *Server) GetUserAddonTemplates(ctx context.Context, req *connect_go.Request[dashv1alpha1.GetUserAddonTemplatesRequest]) (*connect_go.Response[dashv1alpha1.GetUserAddonTemplatesResponse], error) {
	log := clog.FromContext(ctx).WithCaller()

	user := callerFromContext(ctx)

	tmpls, err := s.Klient.ListUserAddonTemplates(ctx)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	if req.Msg.UseRoleFilter != nil && *req.Msg.UseRoleFilter {
		tmpls = kosmo.FilterTemplates(ctx, tmpls, user)
	}

	res := &dashv1alpha1.GetUserAddonTemplatesResponse{
		Items: apiconv.C2D_Templates(tmpls, apiconv.WithTemplateRaw(req.Msg.WithRaw)),
	}

	if len(res.Items) == 0 {
		res.Message = "No items found"
	}

	return connect_go.NewResponse(res), nil
}
