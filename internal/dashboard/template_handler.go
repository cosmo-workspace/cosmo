package dashboard

import (
	"context"
	"net/http"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/gorilla/mux"
)

func (s *Server) useTemplateMiddleWare(router *mux.Router, routes dashv1alpha1.Routes) {
	for _, rt := range routes {
		router.Get(rt.Name).Handler(s.authorizationMiddleware(router.Get(rt.Name).GetHandler()))
	}
}

func (s *Server) GetWorkspaceTemplates(ctx context.Context) (dashv1alpha1.ImplResponse, error) {
	log := clog.FromContext(ctx).WithCaller()

	res := dashv1alpha1.ListTemplatesResponse{}

	tmpls, err := s.Klient.ListTemplatesByType(ctx, []string{wsv1alpha1.TemplateTypeWorkspace})
	if err != nil {
		res.Message = "Failed to list WorkspaceTemplates"
		log.Error(err, res.Message)
		return dashv1alpha1.Response(http.StatusInternalServerError, nil), nil
	}

	wstmpls := make([]dashv1alpha1.Template, 0, len(tmpls))
	for _, v := range tmpls {
		cfg, err := wsv1alpha1.ConfigFromTemplateAnnotations(&v)
		if err != nil {
			log.Info("workspace template is invalid", "error", err, "template", v.Name, "logLevel", "warn")
			continue
		}

		requiredVars := make([]string, 0, len(v.Spec.RequiredVars))
		for _, v := range v.Spec.RequiredVars {
			requiredVars = append(requiredVars, v.Var)
		}

		wstmpl := dashv1alpha1.Template{
			Name:         v.Name,
			RequiredVars: requiredVars,
			UrlBase:      cfg.URLBase,
		}
		wstmpls = append(wstmpls, wstmpl)
	}

	res.Items = wstmpls

	if len(res.Items) == 0 {
		res.Message = "No items found"
	}
	return dashv1alpha1.Response(http.StatusOK, res), nil
}
