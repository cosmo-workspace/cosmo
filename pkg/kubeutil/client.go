package kubeutil

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Apply(ctx context.Context, c client.Client, obj *unstructured.Unstructured, fieldManager string, dryrun, force bool) (patched *unstructured.Unstructured, err error) {
	patched = obj.DeepCopy()

	options := &client.PatchOptions{
		FieldManager: fieldManager,
		Force:        &force,
	}
	if dryrun {
		options.DryRun = []string{metav1.DryRunAll}
	}

	if err := c.Patch(ctx, patched, client.Apply, options); err != nil {
		return nil, err
	}
	return patched, nil
}

func Update(ctx context.Context, c client.Client, obj *unstructured.Unstructured, fieldManager string, dryrun, force bool) (patched *unstructured.Unstructured, err error) {
	patched = obj.DeepCopy()

	options := &client.UpdateOptions{
		FieldManager: fieldManager,
	}
	if dryrun {
		options.DryRun = []string{metav1.DryRunAll}
	}
	if err := c.Update(ctx, patched, options); err != nil {
		return nil, err
	}
	return patched, nil
}

func ApplyOrUpdate(ctx context.Context, c client.Client, obj *unstructured.Unstructured, fieldManager string, dryrun, force bool) (patched *unstructured.Unstructured, err error) {
	patched, err = Apply(ctx, c, obj, fieldManager, dryrun, force)
	if err != nil {
		patched, err = Update(ctx, c, obj, fieldManager, dryrun, force)
	}
	return patched, err
}

func GetUnstructured(ctx context.Context, c client.Client, gvk schema.GroupVersionKind, name, namespace string) (*unstructured.Unstructured, error) {
	var obj unstructured.Unstructured
	obj.SetGroupVersionKind(gvk)

	key := types.NamespacedName{Namespace: namespace, Name: name}
	if err := c.Get(ctx, key, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}
