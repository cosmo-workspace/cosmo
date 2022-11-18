package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
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
	Template cosmov1alpha1.TemplateRef `json:"template"`
	Replicas *int64                    `json:"replicas,omitempty"`
	Vars     map[string]string         `json:"vars,omitempty"`
	Network  []NetworkRule             `json:"network,omitempty"`
}

// WorkspaceStatus has status of Workspace
type WorkspaceStatus struct {
	Instance cosmov1alpha1.ObjectRef `json:"instance,omitempty"`
	Phase    string                  `json:"phase,omitempty"`
	URLs     map[string]string       `json:"urls,omitempty"`
	Config   Config                  `json:"config,omitempty"`
}

// Config defines workspace-dependent configuration
type Config struct {
	DeploymentName      string `json:"deploymentName,omitempty"`
	ServiceName         string `json:"serviceName,omitempty"`
	IngressName         string `json:"ingressName,omitempty"`
	ServiceMainPortName string `json:"mainServicePortName,omitempty"`
	URLBase             string `json:"urlbase,omitempty"`
}

// NetworkRule is an abstract network configuration rule for workspace
type NetworkRule struct {
	Name             string  `json:"name"`
	PortNumber       int     `json:"portNumber"`
	HTTPPath         string  `json:"httpPath"`
	TargetPortNumber *int32  `json:"targetPortNumber,omitempty"`
	Host             *string `json:"host,omitempty"`
	Group            *string `json:"group,omitempty"`
	Public           bool    `json:"public"`
}

func (r *NetworkRule) Default() {
	if r.TargetPortNumber == nil || *r.TargetPortNumber == 0 || r.Public {
		r.TargetPortNumber = pointer.Int32(int32(r.PortNumber))
	}
	if r.HTTPPath == "" {
		r.HTTPPath = "/"
	}
	if r.Group == nil || *r.Group == "" {
		r.Group = &r.Name
	}
}

func (r *NetworkRule) ServicePort() corev1.ServicePort {
	return corev1.ServicePort{
		Name:       r.Name,
		Port:       int32(r.PortNumber),
		Protocol:   corev1.ProtocolTCP,
		TargetPort: intstr.FromInt(int(*r.TargetPortNumber)),
	}
}

func (r *NetworkRule) IngressRule(backendSvcName string) netv1.IngressRule {
	pathTypePrefix := netv1.PathTypePrefix
	var host string
	if r.Host != nil {
		host = *r.Host
	}
	return netv1.IngressRule{
		Host: host,
		IngressRuleValue: netv1.IngressRuleValue{
			HTTP: &netv1.HTTPIngressRuleValue{
				Paths: []netv1.HTTPIngressPath{
					{
						Path:     r.HTTPPath,
						PathType: &pathTypePrefix,
						Backend: netv1.IngressBackend{
							Service: &netv1.IngressServiceBackend{
								Name: backendSvcName,
								Port: netv1.ServiceBackendPort{
									Name: r.Name,
								},
							},
						},
					},
				},
			},
		},
	}
}

func NetworkRulesByServiceAndIngress(svc corev1.Service, ing netv1.Ingress) []NetworkRule {
	netRules := make([]NetworkRule, 0, len(svc.Spec.Ports))
	for _, p := range svc.Spec.Ports {
		var netRule NetworkRule
		netRule.Name = p.Name
		netRule.PortNumber = int(p.Port)

		if p.TargetPort.IntValue() != 0 {
			netRule.TargetPortNumber = pointer.Int32(int32(p.TargetPort.IntValue()))
		}
		for _, rule := range ing.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				if path.Backend.Service != nil {
					if path.Backend.Service.Name == svc.Name {
						if path.Backend.Service.Port.Name == p.Name || path.Backend.Service.Port.Number == p.Port {
							netRule.HTTPPath = path.Path
							netRule.Host = pointer.String(rule.Host)
						}
					}
				}
			}
		}
		netRule.Default()
		netRules = append(netRules, netRule)
	}
	return netRules
}
