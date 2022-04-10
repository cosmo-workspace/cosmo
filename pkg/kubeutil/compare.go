package kubeutil

import (
	"io"
	"os"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Comparable interface {
	GetManagedFields() []metav1.ManagedFieldsEntry
	SetManagedFields(managedFields []metav1.ManagedFieldsEntry)
	SetResourceVersion(resourceVersion string)
}

func resetManagedFieldTime(obj Comparable) {
	mf := obj.GetManagedFields()
	for i := range mf {
		mf[i].Time = nil
	}
	obj.SetManagedFields(mf)
}

func resetResourceVersion(obj Comparable) {
	obj.SetResourceVersion("")
}

type DeepEqualOption interface {
	Apply(x, y Comparable)
}

type printDiff struct {
	out io.Writer
}

func (o printDiff) Apply(x, y Comparable) {
	clog.PrintObjectDiff(o.out, x, y)
}

func WithPrintDiff() DeepEqualOption {
	return printDiff{out: os.Stderr}
}

// LooseDeepEqual deep equal objects without dynamic values
// This function removes some fields, so you should give deep-copied objects.
func LooseDeepEqual(x, y Comparable, opts ...DeepEqualOption) bool {
	resetManagedFieldTime(x)
	resetManagedFieldTime(y)

	resetResourceVersion(x)
	resetResourceVersion(y)

	for _, o := range opts {
		o.Apply(x, y)
	}

	return equality.Semantic.DeepEqual(x, y)
}
