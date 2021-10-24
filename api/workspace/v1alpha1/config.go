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

func AddWorkspaceTemplateVars(vars map[string]string, cfg Config) map[string]string {
	if vars == nil {
		vars = make(map[string]string)
	}
	vars[TemplateVarDeploymentName] = cfg.DeploymentName
	vars[TemplateVarServiceName] = cfg.ServiceName
	vars[TemplateVarIngressName] = cfg.IngressName
	vars[TemplateVarServiceMainPortName] = cfg.ServiceMainPortName
	return vars
}

func SetConfigOnTemplateAnnotations(tmpl *cosmov1alpha1.Template, cfg Config) {
	ann := tmpl.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string)
	}
	ann[InstanceAnnKeyWorkspaceDeployment] = cfg.DeploymentName
	ann[InstanceAnnKeyWorkspaceService] = cfg.ServiceName
	ann[InstanceAnnKeyWorkspaceIngress] = cfg.IngressName
	ann[InstanceAnnKeyWorkspaceServiceMainPort] = cfg.ServiceMainPortName
	ann[InstanceAnnKeyURLBase] = cfg.URLBase
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
		DeploymentName:      ann[InstanceAnnKeyWorkspaceDeployment],
		ServiceName:         ann[InstanceAnnKeyWorkspaceService],
		IngressName:         ann[InstanceAnnKeyWorkspaceIngress],
		ServiceMainPortName: ann[InstanceAnnKeyWorkspaceServiceMainPort],
		URLBase:             ann[InstanceAnnKeyURLBase],
	}
	cfg.Default()

	if cfg.URLBase == "" {
		return cfg, ErrURLBaseNotFound
	}

	return cfg, nil
}
