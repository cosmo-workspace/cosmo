package kubeutil

import (
	"io"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type Comparable interface {
	runtime.Object
	SetGroupVersionKind(gvk schema.GroupVersionKind)
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
	o.out.Write([]byte(cmp.Diff(x, y)))
}

func WithPrintDiff(w io.Writer) DeepEqualOption {
	return printDiff{out: w}
}

type fixGVK struct {
	scheme *runtime.Scheme
}

func (f fixGVK) Apply(x, y Comparable) {
	xgvk, _ := apiutil.GVKForObject(x, f.scheme)
	x.SetGroupVersionKind(xgvk)

	ygvk, _ := apiutil.GVKForObject(y, f.scheme)
	y.SetGroupVersionKind(ygvk)
}

func WithFixGVK(scheme *runtime.Scheme) DeepEqualOption {
	return fixGVK{scheme: scheme}
}

// LooseDeepEqual deep equal objects without dynamic values
func LooseDeepEqual(xObj, yObj Comparable, opts ...DeepEqualOption) bool {
	if xObj == nil && yObj == nil {
		return true
	}
	if xObj == nil || yObj == nil {
		return false
	}

	xCopy := xObj.DeepCopyObject()
	yCopy := yObj.DeepCopyObject()

	if xCopy == nil && yCopy == nil {
		return true
	}
	if xCopy == nil || yCopy == nil {
		return false
	}

	x := xCopy.(Comparable)
	y := yCopy.(Comparable)

	RemoveDynamicFields(x)
	RemoveDynamicFields(y)

	for _, o := range opts {
		o.Apply(x, y)
	}

	return equality.Semantic.DeepEqual(x, y)
}

func RemoveDynamicFields(obj Comparable) {
	resetManagedFieldTime(obj)
	resetResourceVersion(obj)
}

func IsGVKEqual(a, b schema.GroupVersionKind) bool {
	return reflect.DeepEqual(a, b)
}
