package dashboard

import (
	"context"
	"net/http"
	"strconv"

	connect_go "github.com/bufbuild/connect-go"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/utils/pointer"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

func (s *Server) TemplateServiceHandler(mux *http.ServeMux) {
	path, handler := dashboardv1alpha1connect.NewTemplateServiceHandler(s,
		connect_go.WithInterceptors(s.authorizationInterceptor()),
		connect_go.WithInterceptors(s.validatorInterceptor()),
	)
	mux.Handle(path, s.contextMiddleware(handler))
}

func (s *Server) GetWorkspaceTemplates(ctx context.Context, req *connect_go.Request[emptypb.Empty]) (*connect_go.Response[dashv1alpha1.GetWorkspaceTemplatesResponse], error) {
	log := clog.FromContext(ctx).WithCaller()

	user := callerFromContext(ctx)

	tmpls, err := s.Klient.ListWorkspaceTemplates(ctx, user.Spec.Roles)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	addonTmpls := make([]*dashv1alpha1.Template, 0, len(tmpls))
	for _, v := range tmpls {
		addonTmpls = append(addonTmpls, convertTemplateToDashv1alpha1Template(v))
	}

	res := &dashv1alpha1.GetWorkspaceTemplatesResponse{
		Items: addonTmpls,
	}

	if len(res.Items) == 0 {
		res.Message = "No items found"
	}

	return connect_go.NewResponse(res), nil
}

func (s *Server) GetUserAddonTemplates(ctx context.Context, req *connect_go.Request[emptypb.Empty]) (*connect_go.Response[dashv1alpha1.GetUserAddonTemplatesResponse], error) {
	log := clog.FromContext(ctx).WithCaller()

	user := callerFromContext(ctx)

	tmpls, err := s.Klient.ListUserAddonTemplates(ctx, user.Spec.Roles)
	if err != nil {
		return nil, ErrResponse(log, err)
	}

	addonTmpls := make([]*dashv1alpha1.Template, len(tmpls))
	for i, v := range tmpls {
		tmpl := convertTemplateToDashv1alpha1Template(v)

		if ann := v.GetAnnotations(); ann != nil {
			if b, ok := ann[cosmov1alpha1.UserAddonTemplateAnnKeyDefaultUserAddon]; ok {
				if defaultAddon, err := strconv.ParseBool(b); err == nil && defaultAddon {
					tmpl.IsDefaultUserAddon = pointer.Bool(true)
				}
			}
		}

		addonTmpls[i] = tmpl
	}

	res := &dashv1alpha1.GetUserAddonTemplatesResponse{
		Items: addonTmpls,
	}

	if len(res.Items) == 0 {
		res.Message = "No items found"
	}

	return connect_go.NewResponse(res), nil
}

func convertTemplateToDashv1alpha1Template(tmpl cosmov1alpha1.TemplateObject) *dashv1alpha1.Template {
	requiredVars := make([]*dashv1alpha1.TemplateRequiredVars, len(tmpl.GetSpec().RequiredVars))
	for i, v := range tmpl.GetSpec().RequiredVars {
		requiredVars[i] = &dashv1alpha1.TemplateRequiredVars{
			VarName:      v.Var,
			DefaultValue: v.Default,
		}
	}

	return &dashv1alpha1.Template{
		Name:           tmpl.GetName(),
		Description:    tmpl.GetSpec().Description,
		RequiredVars:   requiredVars,
		IsClusterScope: tmpl.GetScope() == meta.RESTScopeRoot,
	}
}
