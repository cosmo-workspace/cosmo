package v1alpha1

import (
	"errors"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
)

var (
	ErrNoAnnotations    = errors.New("no annotations")
	ErrNotTypeWorkspace = errors.New("not type workspace")
	ErrURLBaseNotFound  = errors.New("urlbase not found")
)

// Config defines template-dependent or workspace-dependent configuration metadata for workspace
type Config struct {
	DeploymentName      string `json:"deploymentName,omitempty"`
	ServiceName         string `json:"serviceName,omitempty"`
	IngressName         string `json:"ingressName,omitempty"`
	ServiceMainPortName string `json:"mainServicePortName,omitempty"`
	URLBase             string `json:"urlbase,omitempty"`
}

const (
	DefaultWorkspaceResourceName        string = "workspace"
	DefaultWorkspaceServiceMainPortName string = "default"
)

func (c *Config) Default() {
	if c.DeploymentName == "" {
		c.DeploymentName = DefaultWorkspaceResourceName
	}
	if c.ServiceName == "" {
		c.ServiceName = c.DeploymentName
	}
	if c.IngressName == "" {
		c.IngressName = c.ServiceName
	}
	if c.ServiceMainPortName == "" {
		c.ServiceMainPortName = DefaultWorkspaceServiceMainPortName
	}
}

func SetConfigOnTemplateAnnotations(tmpl *cosmov1alpha1.Template, cfg Config) {
	ann := tmpl.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string)
	}
	ann[TemplateAnnKeyWorkspaceDeployment] = cfg.DeploymentName
	ann[TemplateAnnKeyWorkspaceService] = cfg.ServiceName
	ann[TemplateAnnKeyWorkspaceIngress] = cfg.IngressName
	ann[TemplateAnnKeyWorkspaceServiceMainPort] = cfg.ServiceMainPortName
	ann[TemplateAnnKeyURLBase] = cfg.URLBase
	tmpl.SetAnnotations(ann)
}

func ConfigFromTemplateAnnotations(tmpl *cosmov1alpha1.Template) (cfg Config, err error) {
	ann := tmpl.GetAnnotations()
	if ann == nil {
		cfg.Default()
		return cfg, ErrNoAnnotations
	}

	// check TemplateType label is for Workspace
	tmplType, ok := tmpl.Labels[cosmov1alpha1.LabelKeyTemplateType]
	if !ok || tmplType != TemplateTypeWorkspace {
		cfg.Default()
		return cfg, ErrNotTypeWorkspace
	}

	cfg = Config{
		DeploymentName:      ann[TemplateAnnKeyWorkspaceDeployment],
		ServiceName:         ann[TemplateAnnKeyWorkspaceService],
		IngressName:         ann[TemplateAnnKeyWorkspaceIngress],
		ServiceMainPortName: ann[TemplateAnnKeyWorkspaceServiceMainPort],
		URLBase:             ann[TemplateAnnKeyURLBase],
	}
	cfg.Default()

	if cfg.URLBase == "" {
		return cfg, ErrURLBaseNotFound
	}

	return cfg, nil
}
