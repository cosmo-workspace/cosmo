package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	SchemeBuilder.Register(&ClusterTemplate{}, &ClusterTemplateList{})
}

// +kubebuilder:object:generate=false
type TemplateObject interface {
	metav1.Object
	runtime.Object
	SetGroupVersionKind(gvk schema.GroupVersionKind)
	GetSpec() *TemplateSpec
	GetScope() meta.RESTScope
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster",shortName=ctmpl
// +kubebuilder:storageversion
// ClusterTemplate is the Schema for the Templates API
type ClusterTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TemplateSpec `json:"spec,omitempty"`
}

func (t *ClusterTemplate) GetSpec() *TemplateSpec {
	return &t.Spec
}

func (t *ClusterTemplate) GetScope() meta.RESTScope {
	return meta.RESTScopeRoot
}

// +kubebuilder:object:root=true
// ClusterTemplateList contains a list of ClusterTemplate
type ClusterTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterTemplate `json:"items"`
}
