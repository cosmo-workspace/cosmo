package transformer

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
)

type MetadataTransformer struct {
	inst   *cosmov1alpha1.Instance
	tmpl   *cosmov1alpha1.Template
	scheme *runtime.Scheme
}

func NewMetadataTransformer(inst *cosmov1alpha1.Instance, tmpl *cosmov1alpha1.Template, scheme *runtime.Scheme) *MetadataTransformer {
	return &MetadataTransformer{inst: inst, tmpl: tmpl, scheme: scheme}
}

func (t *MetadataTransformer) Transform(src *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	obj := src.DeepCopy()

	// Set name prefix
	obj.SetName(cosmov1alpha1.InstanceResourceName(t.inst.Name, obj.GetName()))

	// Set namespace
	obj.SetNamespace(t.inst.Namespace)

	// Set labels
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[cosmov1alpha1.LabelKeyInstance] = t.inst.Name
	labels[cosmov1alpha1.LabelKeyTemplate] = t.tmpl.Name
	obj.SetLabels(labels)

	// Set owner reference
	err := ctrl.SetControllerReference(t.inst, obj, t.scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to set owner reference on %s: %w", obj.GetObjectKind().GroupVersionKind(), err)
	}

	return obj, nil
}
