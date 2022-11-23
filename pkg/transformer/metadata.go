package transformer

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
)

type MetadataTransformer struct {
	inst              cosmov1alpha1.InstanceObject
	tmplName          string
	scheme            *runtime.Scheme
	disableNamePrefix bool
}

func NewMetadataTransformer(inst cosmov1alpha1.InstanceObject, scheme *runtime.Scheme, disableNamePrefix bool) *MetadataTransformer {
	return &MetadataTransformer{inst: inst, tmplName: inst.GetSpec().Template.Name, scheme: scheme, disableNamePrefix: disableNamePrefix}
}

func (t *MetadataTransformer) Transform(src *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	obj := src.DeepCopy()

	// Set name prefix
	if !t.disableNamePrefix {
		obj.SetName(instance.InstanceResourceName(t.inst.GetName(), obj.GetName()))
	}

	if t.inst.GetScope() == meta.RESTScopeNamespace {
		// Set namespace
		obj.SetNamespace(t.inst.GetNamespace())
	}

	// Set labels
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[cosmov1alpha1.LabelKeyInstanceName] = t.inst.GetName()
	labels[cosmov1alpha1.LabelKeyTemplateName] = t.tmplName
	obj.SetLabels(labels)

	// Set owner reference
	err := ctrl.SetControllerReference(t.inst, obj, t.scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to set owner reference on %s: %w", obj.GetObjectKind().GroupVersionKind(), err)
	}

	return obj, nil
}
