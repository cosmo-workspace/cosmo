package transformer

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
)

type ScalingTransformer struct {
	instName   string
	scaleSpecs []cosmov1alpha1.ScalingOverrideSpec
}

func NewScalingTransformer(ScalingOverrideSpecs []cosmov1alpha1.ScalingOverrideSpec, instName string) *ScalingTransformer {
	return &ScalingTransformer{scaleSpecs: ScalingOverrideSpecs, instName: instName}
}

func (t *ScalingTransformer) Transform(src *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	obj := src.DeepCopy()

	for _, scaleSpec := range t.scaleSpecs {
		if instance.IsTarget(scaleSpec.Target, t.instName, obj) {
			overrideReplicas(obj, scaleSpec.Replicas)
		}
	}
	return obj, nil
}

func overrideReplicas(obj *unstructured.Unstructured, replicas int64) {
	specObj, ok := obj.Object["spec"]
	if ok {
		spec, ok := specObj.(map[string]interface{})
		if ok {
			var i interface{} = replicas
			spec["replicas"] = i
		}
	}
}
