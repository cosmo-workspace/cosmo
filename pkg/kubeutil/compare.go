package kubeutil

import (
	"io"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Comparable interface {
	GetManagedFields() []metav1.ManagedFieldsEntry
	SetManagedFields(managedFields []metav1.ManagedFieldsEntry)
	SetResourceVersion(resourceVersion string)
	DeepCopyObject() runtime.Object
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

func WithPrintDiff(w io.Writer) DeepEqualOption {
	return printDiff{out: w}
}

// LooseDeepEqual deep equal objects without dynamic values
func LooseDeepEqual(xObj, yObj Comparable, opts ...DeepEqualOption) bool {
	x := xObj.DeepCopyObject().(Comparable)
	y := yObj.DeepCopyObject().(Comparable)

	resetManagedFieldTime(x)
	resetManagedFieldTime(y)

	resetResourceVersion(x)
	resetResourceVersion(y)

	for _, o := range opts {
		o.Apply(x, y)
	}

	return equality.Semantic.DeepEqual(x, y)
}
