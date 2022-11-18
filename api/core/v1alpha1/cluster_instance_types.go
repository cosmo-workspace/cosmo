package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func init() {
	SchemeBuilder.Register(&ClusterInstance{}, &ClusterInstanceList{})
}

// +kubebuilder:object:generate=false
type InstanceObject interface {
	metav1.Object
	runtime.Object
	GetSpec() *InstanceSpec
	GetStatus() *InstanceStatus
	GetScope() meta.RESTScope
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster",shortName=cinst
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="ClusterTemplate",type=string,JSONPath=`.spec.template.name`
// +kubebuilder:printcolumn:name="AppliedResources",type=string,JSONPath=`.status.lastAppliedObjectsCount`
// ClusterInstance is the Schema for the instances API
type ClusterInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

func (i *ClusterInstance) GetSpec() *InstanceSpec {
	return &i.Spec
}

func (i *ClusterInstance) GetStatus() *InstanceStatus {
	return &i.Status
}

func (i *ClusterInstance) GetScope() meta.RESTScope {
	return meta.RESTScopeRoot
}

// +kubebuilder:object:root=true
// ClusterInstanceList contains a list of ClusterInstance
type ClusterInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterInstance `json:"items"`
}

func (l *ClusterInstanceList) InstanceObjects() []InstanceObject {
	i := make([]InstanceObject, 0, len(l.Items))
	for _, v := range l.Items {
		i = append(i, v.DeepCopy())
	}
	return i
}
