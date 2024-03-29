package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// TemplateLabelKeyType is a label key of additional type infomartion on Template
	TemplateLabelKeyType           = "cosmo-workspace.github.io/type"
	TemplateLabelEnumTypeWorkspace = "workspace"
	TemplateLabelEnumTypeUserAddon = "useraddon"

	// TemplateAnnKeyDisableNamePrefix is an annotation key on Template to notify controller not to add name prefix
	TemplateAnnKeyDisableNamePrefix = "cosmo-workspace.github.io/disable-nameprefix"
	// TemplateAnnKeyUserRoles is an annotation key on Template for specific UserRoles
	TemplateAnnKeyUserRoles = "cosmo-workspace.github.io/userroles"
	// TemplateAnnKeyRequiredAddons is a annotation key for Template which requires useraddons
	TemplateAnnKeyRequiredAddons = "cosmo-workspace.github.io/required-useraddons"
)

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster",shortName=tmpl
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.metadata.labels.cosmo-workspace\.github\.io/type`
// Template is the Schema for the Templates API
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TemplateSpec `json:"spec,omitempty"`
}

func (t *Template) GetSpec() *TemplateSpec {
	return &t.Spec
}

// +kubebuilder:object:root=true
// TemplateList contains a list of Template
type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}

// TemplateSpec defines the desired state of Template
type TemplateSpec struct {
	Description  string            `json:"description,omitempty"`
	RequiredVars []RequiredVarSpec `json:"requiredVars,omitempty"`
	RawYaml      string            `json:"rawYaml,omitempty"`
}

// RequiredVarSpec defines a required var spec for template
type RequiredVarSpec struct {
	Var     string `json:"var"`
	Default string `json:"default,omitempty"`
}

func (t *Template) GetScope() meta.RESTScope {
	return meta.RESTScopeNamespace
}
