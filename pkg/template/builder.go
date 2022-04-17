package template

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

type Builder interface {
	Build() ([]unstructured.Unstructured, error)
}

func StringToUnstructured(str string) (*schema.GroupVersionKind, *unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}
	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	_, gvk, err := dec.Decode([]byte(str), nil, obj)
	if err != nil {
		return nil, nil, err
	}
	return gvk, obj, nil
}

func UnstructuredToJSONBytes(obj *unstructured.Unstructured) ([]byte, error) {
	return json.Marshal(obj)
}
