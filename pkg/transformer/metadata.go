package transformer

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
)

type MetadataTransformer struct {
	inst            *cosmov1alpha1.Instance
	tmplName        string
	tmplAnnotations map[string]string
	scheme          *runtime.Scheme
}

func NewMetadataTransformer(inst *cosmov1alpha1.Instance, tmpl *cosmov1alpha1.Template, scheme *runtime.Scheme) *MetadataTransformer {
	return &MetadataTransformer{inst: inst, tmplName: tmpl.GetName(), tmplAnnotations: tmpl.GetAnnotations(), scheme: scheme}
}

func (t *MetadataTransformer) Transform(src *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	obj := src.DeepCopy()

	// Set name prefix
	if !t.disableNamePrefix() {
		obj.SetName(instance.InstanceResourceName(t.inst.Name, obj.GetName()))
	}

	// Set namespace
	obj.SetNamespace(t.inst.Namespace)

	// Set labels
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[cosmov1alpha1.LabelKeyInstance] = t.inst.Name
	labels[cosmov1alpha1.LabelKeyTemplate] = t.tmplName
	obj.SetLabels(labels)

	// Set owner reference
	err := ctrl.SetControllerReference(t.inst, obj, t.scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to set owner reference on %s: %w", obj.GetObjectKind().GroupVersionKind(), err)
	}

	return obj, nil
}

func (t *MetadataTransformer) disableNamePrefix() bool {
	ann := t.tmplAnnotations
	if ann == nil {
		return false
	}
	val := ann[cosmov1alpha1.TemplateAnnKeyDisableNamePrefix]
	disable, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}
	return disable
}
