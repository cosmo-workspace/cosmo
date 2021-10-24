package transformer

import (
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
)

// ToUnstructured makes object to a property of Unstructured
// object must be bool, int64, float64, string, []interface{}, map[string]interface{}, json.Number or nil
func ToUnstructured(obj interface{}) (map[string]interface{}, error) {
	return runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
}

// ToObject makes Unstructured to object
// obj must to be the pointer of object struct
func ToObject(u map[string]interface{}, obj interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(u, obj)
}

func NestedMap(objMap map[string]interface{}, path string) (map[string]interface{}, bool) {
	nested := objMap
	ps := strings.Split(path, ".")
	for _, p := range ps {
		found, ok := nested[p]
		if ok {
			nested, ok = found.(map[string]interface{})
			if !ok {
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	return nested, true
}

func NestedSlice(objMap map[string]interface{}, path string) ([]interface{}, bool) {
	nested := objMap
	ps := strings.Split(path, ".")
	for i, p := range ps {
		if i == len(ps)-1 {
			if found, ok := nested[p]; ok {
				if nestedSlice, ok := found.([]interface{}); ok {
					return nestedSlice, true
				}
			} else {
				return nil, false
			}
		}

		if found, ok := nested[p]; ok {
			nested, ok = found.(map[string]interface{})
			if !ok {
				return nil, false
			}
		} else {
			return nil, false
		}
	}
	return nil, false
}

func NestedMapDelete(objMap map[string]interface{}, path string) bool {
	nested := objMap
	ps := strings.Split(path, ".")
	for i, p := range ps {
		if found, ok := nested[p]; ok {
			if i == len(ps)-1 {
				delete(nested, p)
				return true
			}

			nested, ok = found.(map[string]interface{})
			if !ok {
				return false
			}
		} else {
			return false
		}
	}
	return false
}

func Name(trans Transformer) string {
	return strings.Split(reflect.TypeOf(trans).String(), ".")[1]
}
