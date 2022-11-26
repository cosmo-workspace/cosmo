package template

import (
	"encoding/json"
	"errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

type Builder interface {
	Build() ([]unstructured.Unstructured, error)
}

func BuildObjects(tmplSpec cosmov1alpha1.TemplateSpec, inst cosmov1alpha1.InstanceObject) (objects []unstructured.Unstructured, err error) {
	if tmplSpec.RawYaml != "" {
		objects, err = NewRawYAMLBuilder(tmplSpec.RawYaml, inst).
			ReplaceDefaultVars().
			ReplaceCustomVars().
			Build()
		if err != nil {
			return nil, err
		}

	} else {
		return nil, errors.New("invalid template spec: no template")
	}
	return objects, nil
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
