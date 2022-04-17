package gomega

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func BeEqualityDeepEqual(expected interface{}) types.GomegaMatcher {
	return &EqualityDeepEqualMatcher{
		Expected: expected,
	}
}

type Comparable interface {
	GetName() string
	GetNamespace() string
	GroupVersionKind() schema.GroupVersionKind
}

type EqualityDeepEqualMatcher struct {
	Expected interface{}
}

func (matcher *EqualityDeepEqualMatcher) Match(actual interface{}) (success bool, err error) {
	if actual == nil && matcher.Expected == nil {
		return false, fmt.Errorf("Refusing to compare <nil> to <nil>.\nBe explicit and use BeNil() instead.  This is to avoid mistakes where both sides of an assertion are erroneously uninitialized.")
	}
	return equality.Semantic.DeepEqual(actual, matcher.Expected), nil
}

func (matcher *EqualityDeepEqualMatcher) FailureMessage(actual interface{}) (message string) {
	diff := cmp.Diff(actual, matcher.Expected)

	return fmt.Sprintf("Actual\n%s\nshouled be equal to\n%s\ndiff: %s",
		format.Object(actual, 1), format.Object(matcher.Expected, 1), format.Object(diff, 1))
}

func (matcher *EqualityDeepEqualMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to equal", matcher.Expected)
}
