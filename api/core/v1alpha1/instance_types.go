package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// LabelKeyInstance is a instance name label on the each child resources associated with the instance
	LabelKeyInstance = "cosmo/instance"
	// AnnKeyTemplateUpdated is a annotation on instance to notify template updates to reconcile
	AnnKeyTemplateUpdated = "cosmo/template-updated"
)

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=inst
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Template",type=string,JSONPath=`.status.templateName`
// +kubebuilder:printcolumn:name="Applied-Resources",type=string,JSONPath=`.status.lastApplied[*].kind`
// Instance is the Schema for the instances API
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// InstanceList contains a list of Instance
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

// InstanceSpec defines the desired state of Instance
type InstanceSpec struct {
	// +kubebuilder:validation:Required
	Template TemplateRef       `json:"template"`
	Vars     map[string]string `json:"vars,omitempty"`
	Override OverrideSpec      `json:"override,omitempty"`
}

// TemplateRef defines template to use in Instance creation
type TemplateRef struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

// OverrideSpec defines overrides to transform built objects
type OverrideSpec struct {
	Scale           []ScalingOverrideSpec `json:"scale,omitempty"`
	Network         *NetworkOverrideSpec  `json:"network,omitempty"`
	PatchesJson6902 []Json6902            `json:"patchesJson6902,omitempty"`
}

// NetworkOverrideSpec defines overrides to transform network resources
type NetworkOverrideSpec struct {
	Ingress []IngressOverrideSpec `json:"ingress,omitempty"`
	Service []ServiceOverrideSpec `json:"service,omitempty"`
}

// IngressOverrideSpec defines overrides to transform Ingress resources
type IngressOverrideSpec struct {
	TargetName  string              `json:"targetName,omitempty"`
	Annotations map[string]string   `json:"annotations,omitempty"`
	Rules       []netv1.IngressRule `json:"rules,omitempty"`
}

// ServiceOverrideSpec defines overrides to transform Service resources
type ServiceOverrideSpec struct {
	TargetName string               `json:"targetName,omitempty"`
	Ports      []corev1.ServicePort `json:"ports,omitempty"`
}

// ScalingOverrideSpec defines workload scales.
type ScalingOverrideSpec struct {
	Target   ObjectRef `json:"target"`
	Replicas int64     `json:"replicas"`
}

// Json6902 defines JSONPatch specs.
type Json6902 struct {
	Target ObjectRef `json:"target"`
	Patch  string    `json:"patch,omitempty"`
}

// InstanceStatus has status of Instance
type InstanceStatus struct {
	TemplateName string      `json:"templateName,omitempty"`
	LastApplied  []ObjectRef `json:"lastApplied,omitempty"`
}

// ObjectRef is a reference of resource which is created by the Instance
type ObjectRef struct {
	APIVersion        string       `json:"apiVersion"`
	Kind              string       `json:"kind"`
	Name              string       `json:"name,omitempty"`
	Namespace         string       `json:"namespace,omitempty"`
	CreationTimestamp *metav1.Time `json:"creationTimestamp,omitempty"`
	UpdateTimestamp   *metav1.Time `json:"updateTimestamp,omitempty"`
}

type gvkObject interface {
	GroupVersionKind() schema.GroupVersionKind
	GetName() string
}

func (r ObjectRef) IsTarget(instanceName string, obj gvkObject) bool {
	return IsGVKEqual(r.GroupVersionKind(), obj.GroupVersionKind()) && EqualInstanceResourceName(instanceName, r.Name, obj.GetName())
}

func (r *ObjectRef) SetName(name string) {
	r.Name = name
}

func (r *ObjectRef) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	r.APIVersion = gvk.GroupVersion().String()
	r.Kind = gvk.Kind
}

func (r ObjectRef) GroupVersionKind() schema.GroupVersionKind {
	gv, err := schema.ParseGroupVersion(r.APIVersion)
	if err != nil {
		return schema.GroupVersionKind{}
	}
	gvk := gv.WithKind(r.Kind)
	return gvk
}

func (r ObjectRef) GetName() string {
	return r.Name
}
