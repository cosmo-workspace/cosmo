package dashboard

import (
	"context"
	"net/http"
	"strconv"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/gorilla/mux"
	"k8s.io/utils/pointer"
)

func (s *Server) useTemplateMiddleWare(router *mux.Router, routes dashv1alpha1.Routes) {
	for _, rt := range routes {
		router.Get(rt.Name).Handler(s.authorizationMiddleware(router.Get(rt.Name).GetHandler()))
	}
}

func (s *Server) GetWorkspaceTemplates(ctx context.Context) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()

	tmpls, err := s.Klient.ListWorkspaceTemplates(ctx)
	if err != nil {
		return ErrorResponse(log, err)
	}

	addonTmpls := make([]dashv1alpha1.Template, 0, len(tmpls))
	for _, v := range tmpls {
		addonTmpls = append(addonTmpls, convertTemplateToDashv1alpha1Template(v))
	}

	res := dashv1alpha1.ListTemplatesResponse{
		Items: addonTmpls,
	}
	if len(res.Items) == 0 {
		res.Message = "No items found"
	}
	return NormalResponse(http.StatusOK, res)
}

func (s *Server) GetUserAddonTemplates(ctx context.Context) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()

	tmpls, err := s.Klient.ListUserAddonTemplates(ctx)
	if err != nil {
		return ErrorResponse(log, err)
	}

	addonTmpls := make([]dashv1alpha1.Template, len(tmpls))
	for i, v := range tmpls {
		tmpl := convertTemplateToDashv1alpha1Template(v)

		if ann := v.GetAnnotations(); ann != nil {
			if b, ok := ann[wsv1alpha1.TemplateAnnKeyDefaultUserAddon]; ok {
				if defaultAddon, err := strconv.ParseBool(b); err == nil && defaultAddon {
					tmpl.IsDefaultUserAddon = pointer.Bool(true)
				}
			}
		}

		addonTmpls[i] = tmpl
	}

	res := dashv1alpha1.ListTemplatesResponse{
		Items: addonTmpls,
	}
	if len(res.Items) == 0 {
		res.Message = "No items found"
	}
	return NormalResponse(http.StatusOK, res)
}

func convertTemplateToDashv1alpha1Template(tmpl cosmov1alpha1.Template) dashv1alpha1.Template {
	requiredVars := make([]dashv1alpha1.TemplateRequiredVars, len(tmpl.Spec.RequiredVars))
	for i, v := range tmpl.Spec.RequiredVars {
		requiredVars[i] = dashv1alpha1.TemplateRequiredVars{
			VarName:      v.Var,
			DefaultValue: v.Default,
		}
	}

	return dashv1alpha1.Template{
		Name:         tmpl.Name,
		RequiredVars: requiredVars,
		Description:  tmpl.Spec.Description,
	}
}
