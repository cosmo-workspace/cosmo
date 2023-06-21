package v1alpha1

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=ws
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Template",type=string,JSONPath=`.spec.template.name`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// Workspace is the Schema for the workspaces API
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceSpec   `json:"spec,omitempty"`
	Status WorkspaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// WorkspaceList contains a list of Workspace
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

// WorkspaceSpec defines the desired state of Workspace
type WorkspaceSpec struct {
	// +kubebuilder:validation:Required
	Template TemplateRef       `json:"template"`
	Replicas *int64            `json:"replicas,omitempty"`
	Vars     map[string]string `json:"vars,omitempty"`
	Network  []NetworkRule     `json:"network,omitempty"`
}

// WorkspaceStatus has status of Workspace
type WorkspaceStatus struct {
	Instance ObjectRef         `json:"instance,omitempty"`
	Phase    string            `json:"phase,omitempty"`
	URLs     map[string]string `json:"urls,omitempty"`
	Config   Config            `json:"config,omitempty"`
}

// Config defines workspace-dependent configuration
type Config struct {
	DeploymentName      string `json:"deploymentName,omitempty"`
	ServiceName         string `json:"serviceName,omitempty"`
	ServiceMainPortName string `json:"mainServicePortName,omitempty"`
}

const (
	// WorkspaceTemplateAnnKeys are annotation keys for WorkspaceConfig
	WorkspaceTemplateAnnKeyDeploymentName  = "workspace.cosmo-workspace.github.io/deployment"
	WorkspaceTemplateAnnKeyServiceName     = "workspace.cosmo-workspace.github.io/service"
	WorkspaceTemplateAnnKeyServiceMainPort = "workspace.cosmo-workspace.github.io/service-main-port"
)

const (
	// TemplateVars are Template variables to set WorkspaceConfig info on resources in the Template
	WorkspaceTemplateVarDeploymentName      = "{{WORKSPACE_DEPLOYMENT_NAME}}"
	WorkspaceTemplateVarServiceName         = "{{WORKSPACE_SERVICE_NAME}}"
	WorkspaceTemplateVarServiceMainPortName = "{{WORKSPACE_SERVICE_MAIN_PORT_NAME}}"
)

// NetworkRule is an abstract network configuration rule for workspace
type NetworkRule struct {
	Name             string  `json:"name"`
	PortNumber       int32   `json:"portNumber"`
	HTTPPath         string  `json:"httpPath"`
	TargetPortNumber *int32  `json:"targetPortNumber,omitempty"`
	Host             *string `json:"host,omitempty"`
	Group            *string `json:"group,omitempty"`
	Public           bool    `json:"public"`
}

func (r *NetworkRule) Default() {
	if r.HTTPPath == "" {
		r.HTTPPath = "/"
	}
	if r.Group == nil || *r.Group == "" {
		r.Group = &r.Name
	}
}

func (r *NetworkRule) portName() string {
	return fmt.Sprintf("port%d", r.PortNumber)
}

func (r *NetworkRule) ServicePort() corev1.ServicePort {
	targetPort := r.PortNumber
	if r.TargetPortNumber != nil && *r.TargetPortNumber != 0 {
		targetPort = *r.TargetPortNumber
	}

	return corev1.ServicePort{
		Name:       r.portName(),
		Port:       r.PortNumber,
		Protocol:   corev1.ProtocolTCP,
		TargetPort: intstr.FromInt(int(targetPort)),
	}
}

func NetworkRulesByService(svc corev1.Service) []NetworkRule {
	netRules := make([]NetworkRule, 0, len(svc.Spec.Ports))
	for _, p := range svc.Spec.Ports {
		var netRule NetworkRule
		netRule.Name = p.Name
		netRule.PortNumber = p.Port

		if p.TargetPort.IntValue() != 0 {
			netRule.TargetPortNumber = pointer.Int32(int32(p.TargetPort.IntValue()))
		}
		netRule.Default()
		netRules = append(netRules, netRule)
	}
	return netRules
}
