package transformer

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Transformer interface {
	Transform(*unstructured.Unstructured) (*unstructured.Unstructured, error)
}
