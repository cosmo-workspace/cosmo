package v1alpha1

import (
	"fmt"
	"reflect"
	"slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

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

func SetOwnerReferenceIfNotKeepPolicy(owner metav1.Object, obj metav1.Object, scheme *runtime.Scheme) error {
	if !KeepResourceDeletePolicy(owner) && !KeepResourceDeletePolicy(obj) {
		// Set owner reference
		err := ctrl.SetControllerReference(owner, obj, scheme)
		if err != nil {
			return fmt.Errorf("failed to set owner reference on %s: %w", obj.(runtime.Object).GetObjectKind().GroupVersionKind(), err)
		}
		return nil

	} else {
		// Remove owner reference
		if len(obj.GetOwnerReferences()) > 0 {
			gvk, _ := apiutil.GVKForObject(owner.(runtime.Object), scheme)
			refs := slices.DeleteFunc(obj.GetOwnerReferences(), func(v metav1.OwnerReference) bool {
				return reflect.DeepEqual(v, metav1.OwnerReference{
					APIVersion:         gvk.GroupVersion().String(),
					Kind:               gvk.Kind,
					Name:               owner.GetName(),
					UID:                owner.GetUID(),
					Controller:         ptr.To(true),
					BlockOwnerDeletion: ptr.To(true),
				})
			})
			obj.SetOwnerReferences(refs)
		}
		return nil
	}
}

const (
	EventAnnKeyUserName     = "cosmo-workspace.github.io/user"
	EventAnnKeyInstanceName = LabelKeyInstanceName
)
