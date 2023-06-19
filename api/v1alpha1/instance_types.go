package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// LabelKeyInstanceName is a instance name label on the each child resources associated with the instance
	LabelKeyInstanceName = "cosmo-workspace.github.io/instance"
	// LabelKeyTemplateName is a template name label on the resources created by instance
	LabelKeyTemplateName = "cosmo-workspace.github.io/template"
)

func init() {
	SchemeBuilder.Register(&Instance{}, &InstanceList{})
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=inst
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Template",type=string,JSONPath=`.spec.template.name`
// +kubebuilder:printcolumn:name="AppliedResources",type=string,JSONPath=`.status.lastAppliedObjectsCount`
// Instance is the Schema for the instances API
type Instance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InstanceSpec   `json:"spec,omitempty"`
	Status InstanceStatus `json:"status,omitempty"`
}

func (i *Instance) GetSpec() *InstanceSpec {
	return &i.Spec
}

func (i *Instance) GetStatus() *InstanceStatus {
	return &i.Status
}

func (i *Instance) GetScope() meta.RESTScope {
	return meta.RESTScopeNamespace
}

// +kubebuilder:object:root=true
// InstanceList contains a list of Instance
type InstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Instance `json:"items"`
}

func (l *InstanceList) InstanceObjects() []InstanceObject {
	i := make([]InstanceObject, 0, len(l.Items))
	for _, v := range l.Items {
		i = append(i, v.DeepCopy())
	}
	return i
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
	PatchesJson6902 []Json6902 `json:"patchesJson6902,omitempty"`
}

// Json6902 defines JSONPatch specs.
type Json6902 struct {
	Target ObjectRef `json:"target"`
	Patch  string    `json:"patch,omitempty"`
}

// InstanceStatus has status of Instance
type InstanceStatus struct {
	TemplateName            string      `json:"templateName,omitempty"`
	TemplateResourceVersion string      `json:"templateResourceVersion,omitempty"`
	LastApplied             []ObjectRef `json:"lastApplied,omitempty"`
	LastAppliedObjectsCount int         `json:"lastAppliedObjectsCount,omitempty"`
	TemplateObjectsCount    int         `json:"templateObjectsCount,omitempty"`
}

// ObjectRef is a reference of resource which is created by the Instance
type ObjectRef struct {
	corev1.ObjectReference `json:",inline"`
	CreationTimestamp      *metav1.Time `json:"creationTimestamp,omitempty"`
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
