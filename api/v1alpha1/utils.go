package v1alpha1

// LabelControllerManaged is a label on all resources managed by the controllers
const LabelControllerManaged = "cosmo-workspace.github.io/controller-managed"

// +kubebuilder:object:generate=false
type LabelHolder interface {
	GetLabels() map[string]string
	SetLabels(map[string]string)
}

func SetControllerManaged(obj LabelHolder) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[LabelControllerManaged] = "1"
	obj.SetLabels(labels)
}

// +kubebuilder:object:generate=false
type AnnotationHolder interface {
	GetAnnotations() map[string]string
	SetAnnotations(map[string]string)
}

// KeepResourceDeletePolicy returns true if the resource has annotation delete-policy=keep
func KeepResourceDeletePolicy(obj AnnotationHolder) bool {
	ann := obj.GetAnnotations()
	if ann == nil {
		return false
	}
	v, ok := ann[ResourceAnnKeyDeletePolicy]
	if !ok {
		return false
	}
	return v == ResourceAnnEnumDeletePolicyKeep
}

const (
	EventAnnKeyUserName     = "cosmo-workspace.github.io/user"
	EventAnnKeyInstanceName = LabelKeyInstanceName
)
