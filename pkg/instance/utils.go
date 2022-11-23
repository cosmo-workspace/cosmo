package instance

import (
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
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

type GVKNameGetter interface {
	GroupVersionKind() schema.GroupVersionKind
	GetName() string
}

func ExistInLastApplyed(inst cosmov1alpha1.InstanceObject, targetObj GVKNameGetter) bool {
	lastApplied := inst.GetStatus().LastApplied
	if len(lastApplied) == 0 {
		return false
	}

	for _, ref := range lastApplied {
		if IsTarget(ref, inst.GetName(), targetObj) {
			return true
		}
	}
	return false
}

func IsTarget(ref cosmov1alpha1.ObjectRef, instanceName string, obj GVKNameGetter) bool {
	return reflect.DeepEqual(ref.GroupVersionKind(), obj.GroupVersionKind()) && EqualInstanceResourceName(instanceName, ref.Name, obj.GetName())
}
