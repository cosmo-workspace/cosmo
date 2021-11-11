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

func (t *UnstructuredBuilder) ReplaceDefaultVars() *UnstructuredBuilder {
	t.rawYaml = strings.ReplaceAll(t.rawYaml, DefaultVarsInstance, t.inst.Name)
	t.rawYaml = strings.ReplaceAll(t.rawYaml, DefaultVarsNamespace, t.inst.Namespace)
	t.rawYaml = strings.ReplaceAll(t.rawYaml, DefaultVarsTemplate, t.inst.Spec.Template.Name)
	return t
}

func (t *UnstructuredBuilder) ReplaceCustomVars() *UnstructuredBuilder {
	if t.inst.Spec.Vars != nil {
		for key, val := range t.inst.Spec.Vars {
			key = FixupTemplateVarKey(key)
			t.rawYaml = strings.ReplaceAll(t.rawYaml, key, val)
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
