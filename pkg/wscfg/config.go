package wscfg

import (
	"errors"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
)

var (
	ErrNotTypeWorkspace = errors.New("not type workspace")
)

func SetConfigOnTemplateAnnotations(tmpl *cosmov1alpha1.Template, cfg wsv1alpha1.Config) {
	ann := tmpl.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string)
	}
	ann[wsv1alpha1.TemplateAnnKeyWorkspaceDeployment] = cfg.DeploymentName
	ann[wsv1alpha1.TemplateAnnKeyWorkspaceService] = cfg.ServiceName
	ann[wsv1alpha1.TemplateAnnKeyWorkspaceIngress] = cfg.IngressName
	ann[wsv1alpha1.TemplateAnnKeyWorkspaceServiceMainPort] = cfg.ServiceMainPortName
	ann[wsv1alpha1.TemplateAnnKeyURLBase] = cfg.URLBase
	tmpl.SetAnnotations(ann)
}

func ConfigFromTemplateAnnotations(tmpl *cosmov1alpha1.Template) (cfg wsv1alpha1.Config, err error) {
	// check TemplateType label is for Workspace
	tmplType, ok := tmpl.Labels[cosmov1alpha1.LabelKeyTemplateType]
	if !ok || tmplType != wsv1alpha1.TemplateTypeWorkspace {
		return cfg, ErrNotTypeWorkspace
	}

	ann := tmpl.GetAnnotations()
	if ann == nil {
		return cfg, nil
	}

	cfg = wsv1alpha1.Config{
		DeploymentName:      ann[wsv1alpha1.TemplateAnnKeyWorkspaceDeployment],
		ServiceName:         ann[wsv1alpha1.TemplateAnnKeyWorkspaceService],
		IngressName:         ann[wsv1alpha1.TemplateAnnKeyWorkspaceIngress],
		ServiceMainPortName: ann[wsv1alpha1.TemplateAnnKeyWorkspaceServiceMainPort],
		URLBase:             ann[wsv1alpha1.TemplateAnnKeyURLBase],
	}
	return cfg, nil
}
