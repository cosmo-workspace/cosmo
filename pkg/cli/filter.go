package cli

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Filter struct {
	Key      string
	Value    string
	Operator string
}

const (
	OperatorEqual    = "=="
	OperatorNotEqual = "!="
)

func opBool(op string) bool {
	if op == OperatorEqual {
		return true
	} else if op == OperatorNotEqual {
		return false
	}
	panic("unknown operator: " + op)
}

func ParseFilters(filterExpressions []string) ([]Filter, error) {
	filters := make([]Filter, 0, len(filterExpressions))
	for _, e := range filterExpressions {
		var f *Filter
		if strings.Contains(e, OperatorNotEqual) {
			f = parseFilterExpression(e, OperatorNotEqual)
		} else if strings.Contains(e, OperatorEqual) {
			f = parseFilterExpression(e, OperatorEqual)
		} else {
			return nil, fmt.Errorf("invalid filter expression: %s", e)
		}
		if f != nil {
			filters = append(filters, *f)
		}
	}
	return filters, nil
}

func parseFilterExpression(exp, op string) *Filter {
	f := Filter{}
	s := strings.Split(exp, op)
	if len(s) != 2 {
		return nil
	}
	f.Key = s[0]
	f.Value = s[1]
	f.Operator = op
	return &f
}

func DoFilter[T any](objects []T, objectFilterKeyFunc func(T) []string, f Filter) []T {
	filtered := make([]T, 0, len(objects))
	for _, o := range objects {
		values := objectFilterKeyFunc(o)

		matched := false

	KeysLoop:
		for _, v := range values {
			found, err := filepath.Match(f.Value, v)
			if err != nil {
				continue KeysLoop
			}
			switch f.Operator {
			case OperatorEqual:
				if found {
					matched = true
					break KeysLoop
				}
			case OperatorNotEqual:
				if found {
					matched = false
					break KeysLoop
				} else {
					matched = true
				}
			}
		}

		if matched {
			filtered = append(filtered, o)
		}
	}
	return filtered
}
