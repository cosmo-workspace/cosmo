package transformer

import (
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

type JSONPatchTransformer struct {
	instName string
	patch    []cosmov1alpha1.Json6902
}

func NewJSONPatchTransformer(json6902 []cosmov1alpha1.Json6902, instName string) *JSONPatchTransformer {
	return &JSONPatchTransformer{patch: json6902, instName: instName}
}

func (t *JSONPatchTransformer) Transform(src *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	obj := src.DeepCopy()

	for _, v := range t.patch {
		if instance.IsTarget(v.Target, t.instName, obj) {
			bobj, err := template.UnstructuredToJSONBytes(obj)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal unstructured: %w", err)
			}

			// patch JSON6902
			patch, err := jsonpatch.DecodePatch([]byte(v.Patch))
			if err != nil {
				return nil, fmt.Errorf("failed to decode patch: %w: %s", err, v.Patch)
			}

			patched, err := patch.Apply(bobj)
			if err != nil {
				return nil, fmt.Errorf("failed to patch JSON6902: %w", err)
			}

			_, obj, err = template.StringToUnstructured(string(patched))
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal unstructured: %w", err)
			}
		}
	}

	return obj, nil
}
