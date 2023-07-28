package v1alpha1

import "strconv"

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

// AnnotationPruneDisabled is a bool annotation for each child resources not to be deleted in GC
const AnnotationPruneDisabled = "cosmo-workspace.github.io/prune-disabled"

func IsPruneDisabled(obj AnnotationHolder) bool {
	ann := obj.GetAnnotations()
	if ann == nil {
		return false
	}
	v, ok := ann[AnnotationPruneDisabled]
	if !ok {
		return false
	}
	isDisabled, err := strconv.ParseBool(v)
	if err != nil {
		// invalid bool value might be accidentally set while trying to be true
		return true
	}
	return isDisabled
}
