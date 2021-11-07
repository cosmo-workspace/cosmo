package template

import (
	"encoding/json"
	"strings"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

type TemplateBuilder struct {
	data string
	inst *cosmov1alpha1.Instance
}

func NewTemplateBuilder(data string, inst *cosmov1alpha1.Instance) *TemplateBuilder {
	return &TemplateBuilder{
		data: data,
		inst: inst,
	}
}

func (t *TemplateBuilder) Build() ([]unstructured.Unstructured, error) {
	splitString := strings.Split(t.data, "---")
	resources := make([]unstructured.Unstructured, 0, len(splitString))
	for _, v := range splitString {
		if v == "" {
			continue
		}
		_, obj, err := StringToUnstructured(v)
		if err != nil {
			return nil, err
		}
		resources = append(resources, *obj)
	}
	return resources, nil
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
