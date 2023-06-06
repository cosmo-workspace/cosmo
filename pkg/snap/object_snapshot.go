package snap

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

func UserSnapshot(in *cosmov1alpha1.User) *cosmov1alpha1.User {
	obj := in.DeepCopy()
	RemoveDynamicFields(obj)

	obj.Status.Namespace.CreationTimestamp = nil
	obj.Status.Namespace.UID = ""
	obj.Status.Namespace.ResourceVersion = ""

	for i, v := range obj.Status.Addons {
		v.CreationTimestamp = nil
		v.UID = ""
		v.ResourceVersion = ""
		obj.Status.Addons[i] = v
	}
	sort.Slice(obj.Status.Addons, func(i, j int) bool {
		return obj.Status.Addons[i].Kind < obj.Status.Addons[j].Kind
	})
	sort.Slice(obj.Status.Addons, func(i, j int) bool {
		return obj.Status.Addons[i].Name < obj.Status.Addons[j].Name
	})

	return obj
}

func InstanceSnapshot(in cosmov1alpha1.InstanceObject) cosmov1alpha1.InstanceObject {
	o := in.DeepCopyObject()
	obj := o.(cosmov1alpha1.InstanceObject)
	RemoveDynamicFields(obj)

	for i, v := range obj.GetStatus().LastApplied {
		v.CreationTimestamp = nil
		v.UID = ""
		v.ResourceVersion = ""
		obj.GetStatus().LastApplied[i] = v
	}
	sort.Slice(obj.GetStatus().LastApplied, func(i, j int) bool {
		return obj.GetStatus().LastApplied[i].Kind < obj.GetStatus().LastApplied[j].Kind
	})
	sort.Slice(obj.GetStatus().LastApplied, func(i, j int) bool {
		return obj.GetStatus().LastApplied[i].Name < obj.GetStatus().LastApplied[j].Name
	})
	obj.GetStatus().TemplateResourceVersion = ""

	return obj
}

func ServiceSnapshot(in *corev1.Service) *corev1.Service {
	obj := in.DeepCopy()
	RemoveDynamicFields(obj)

	obj.Spec.ClusterIP = ""
	obj.Spec.ClusterIPs = nil

	for i, p := range obj.Spec.Ports {
		if p.NodePort >= 30000 {
			obj.Spec.Ports[i].NodePort = 30000
		}
	}

	return obj
}

func ObjectSnapshot(obj client.Object) client.Object {
	t := obj.DeepCopyObject()
	o := t.(client.Object)
	RemoveDynamicFields(o)
	return o
}

func RemoveDynamicFields(o client.Object) {
	o.SetCreationTimestamp(metav1.Time{})
	o.SetResourceVersion("")
	o.SetGeneration(0)
	o.SetUID(types.UID(""))
	o.SetManagedFields(nil)

	ownerRefs := make([]metav1.OwnerReference, len(o.GetOwnerReferences()))
	for i, v := range o.GetOwnerReferences() {
		v.UID = ""
		ownerRefs[i] = v
	}
	o.SetOwnerReferences(ownerRefs)
}
