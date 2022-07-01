package gomega

import (
	"fmt"

	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"k8s.io/client-go/kubernetes/scheme"
)

func BeLooseDeepEqual(expected interface{}) types.GomegaMatcher {
	return &LooseDeepEqualMatcher{
		Expected: expected,
	}
}

type LooseDeepEqualMatcher struct {
	Expected interface{}
}

func (matcher *LooseDeepEqualMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil && matcher.Expected == nil {
		return false, fmt.Errorf("Refusing to compare <nil> to <nil>.\nBe explicit and use BeNil() instead.  This is to avoid mistakes where both sides of an assertion are erroneously uninitialized.")
	}

	actualObj, actualObjOk := actual.(kubeutil.Comparable)
	expectedObj, expectedObjOk := matcher.Expected.(kubeutil.Comparable)
	if !actualObjOk || !expectedObjOk {
		return false, fmt.Errorf("Refusing to compare non kubeutil.Comparable objects.")
	}

	return kubeutil.LooseDeepEqual(actualObj, expectedObj, kubeutil.WithFixGVK(scheme.Scheme)), nil
}

func (matcher *LooseDeepEqualMatcher) FailureMessage(actual interface{}) (message string) {
	a, e, diff := looseOutput(actual, matcher.Expected)
	format.MaxLength = 0
	return fmt.Sprintf("Actual\n%s\nshouled be equal to\n%s\ndiff: %s",
		format.Object(a, 1), format.Object(e, 1), format.Object(diff, 1))
}

func (matcher *LooseDeepEqualMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	a, e, _ := looseOutput(actual, matcher.Expected)
	format.MaxLength = 0
	return format.Message(a, "not to equal", e)
}

func looseOutput(actual interface{}, expected interface{}) (a kubeutil.Comparable, e kubeutil.Comparable, diff string) {
	actualObj := actual.(kubeutil.Comparable)
	expectedObj := expected.(kubeutil.Comparable)

	a = actualObj.DeepCopyObject().(kubeutil.Comparable)
	e = expectedObj.DeepCopyObject().(kubeutil.Comparable)

	kubeutil.RemoveDynamicFields(a)
	kubeutil.RemoveDynamicFields(e)
	return a, e, cmp.Diff(a, e)
}
