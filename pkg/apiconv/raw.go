package apiconv

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

var scheme = runtime.NewScheme()

func init() {
	_ = cosmov1alpha1.AddToScheme(scheme)
}

func ToYAML(obj client.Object) *string {
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return nil
	}

	obj.GetObjectKind().SetGroupVersionKind(gvk)
	raw, err := yaml.Marshal(removeUnnecessaryFields(obj))

	if err != nil || raw == nil {
		return nil
	}
	return ptr.To(string(raw))
}

func DecodeYAML[T client.Object](raw string, obj T) error {
	return yaml.Unmarshal([]byte(raw), obj)
}

func removeUnnecessaryFields(obj client.Object) client.Object {
	newObj := obj.DeepCopyObject().(client.Object)
	newObj.SetManagedFields(nil)
	return newObj
}
