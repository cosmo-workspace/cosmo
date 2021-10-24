package template

import (
	"errors"
	"strings"
)

const (
	DefaultVarsInstance  = "{{INSTANCE}}"
	DefaultVarsNamespace = "{{NAMESPACE}}"
	DefaultVarsTemplate  = "{{TEMPLATE}}"
)

func (t *TemplateBuilder) ReplaceDefaultVars() *TemplateBuilder {
	t.data = strings.ReplaceAll(t.data, DefaultVarsInstance, t.inst.Name)
	t.data = strings.ReplaceAll(t.data, DefaultVarsNamespace, t.inst.Namespace)
	t.data = strings.ReplaceAll(t.data, DefaultVarsTemplate, t.inst.Spec.Template.Name)
	return t
}

func (t *TemplateBuilder) ReplaceCustomVars() *TemplateBuilder {
	if t.inst.Spec.Vars != nil {
		for key, val := range t.inst.Spec.Vars {
			key = FixupTemplateVarKey(key)
			t.data = strings.ReplaceAll(t.data, key, val)
		}
	}
	return t
}

func FixupTemplateVarKey(key string) string {
	if !strings.HasPrefix(key, "{{") {
		key = "{{" + key
	}
	if !strings.HasSuffix(key, "}}") {
		key = key + "}}"
	}
	return key
}

var (
	ErrInvalidVars = errors.New("invalid Vars string")
)

func ValidCustomVars(varString string) error {
	if strings.HasPrefix(varString, "{{") && strings.HasSuffix(varString, "}}") {
		return nil
	}
	return ErrInvalidVars
}
