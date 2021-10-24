package v1alpha1

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// InstanceResourceName is a child resource name created by instance
// All resources associated with instance has "INSTANCE_NAME-" prefix
func InstanceResourceName(instanceName, resourceName string) string {
	if strings.HasPrefix(resourceName, instanceName+"-") {
		return resourceName
	}
	return instanceName + "-" + resourceName
}

// EqualInstanceResourceName compare child resource names
func EqualInstanceResourceName(instanceName, a, b string) bool {
	return InstanceResourceName(instanceName, a) == InstanceResourceName(instanceName, b)
}

// UnstructuredToResourceRef generate ObjectRef by Unstructured object
func UnstructuredToResourceRef(obj unstructured.Unstructured, updateTimestamp metav1.Time) ObjectRef {
	ref := ObjectRef{}
	ref.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
	ref.Name = obj.GetName()
	ref.Namespace = obj.GetNamespace()

	create := obj.GetCreationTimestamp()
	ref.CreationTimestamp = &create
	ref.UpdateTimestamp = &updateTimestamp
	return ref
}

func IsGVKEqual(a, b schema.GroupVersionKind) bool {
	return a.Group == b.Group && a.Kind == b.Kind && a.Version == b.Version
}

func ExistInLastApplyed(inst Instance, gvkObj gvkObject) bool {
	lastApplied := inst.Status.LastApplied
	if len(lastApplied) == 0 {
		return false
	}

	for _, v := range lastApplied {
		if v.IsTarget(inst.Name, gvkObj) {
			return true
		}
	}
	return false
}
