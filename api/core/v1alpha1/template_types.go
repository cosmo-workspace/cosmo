package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// LabelKeyTemplate is a template name label on the resources created by instance
	LabelKeyTemplate = "cosmo/template"
	// TemplateLabelKeyType is a additional type infomartion on template
	TemplateLabelKeyType = "cosmo/type"
	// TemplateAnnKeyDisableNamePrefix is a annotation on template to notify controller not to add name prefix
	TemplateAnnKeyDisableNamePrefix = "cosmo/disable-nameprefix"
	// TemplateAnnKeySkipValidation is a annotation on template to notify webhook not to validate
	TemplateAnnKeySkipValidation = "cosmo/skip-validation"
)

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster",shortName=tmpl
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Required-Vars",type=string,JSONPath=`.spec.requiredVars`
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
