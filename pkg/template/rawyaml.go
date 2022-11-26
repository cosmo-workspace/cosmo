package template

import (
	"errors"
	"regexp"
	"strings"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	DefaultVarsInstance  = "{{INSTANCE}}"
	DefaultVarsNamespace = "{{NAMESPACE}}"
	DefaultVarsTemplate  = "{{TEMPLATE}}"
)

var (
	ErrInvalidVars = errors.New("invalid Vars string")
)

type RawYAMLBuilder struct {
	rawYaml string
	inst    cosmov1alpha1.InstanceObject
}

func NewRawYAMLBuilder(rawYaml string, inst cosmov1alpha1.InstanceObject) *RawYAMLBuilder {
	return &RawYAMLBuilder{
		rawYaml: rawYaml,
		inst:    inst,
	}
}

func (t *RawYAMLBuilder) Build() ([]unstructured.Unstructured, error) {
	splitString := regexp.MustCompile(`(?m)^---$`).Split(t.rawYaml, -1)
	resources := make([]unstructured.Unstructured, 0, len(splitString))
	for _, v := range splitString {
		if strings.TrimSpace(v) == "" {
			continue
		}
		_, obj, err := StringToUnstructured(v)
		if err != nil {
			return nil, err
		}
		resources = append(resources, *obj)
	}
	return resources, nil
}

func (t *RawYAMLBuilder) ReplaceDefaultVars() *RawYAMLBuilder {
	t.rawYaml = strings.ReplaceAll(t.rawYaml, DefaultVarsInstance, t.inst.GetName())
	t.rawYaml = strings.ReplaceAll(t.rawYaml, DefaultVarsTemplate, t.inst.GetSpec().Template.Name)

	if t.inst.GetScope() == meta.RESTScopeNamespace {
		t.rawYaml = strings.ReplaceAll(t.rawYaml, DefaultVarsNamespace, t.inst.GetNamespace())
	}
	return t
}

func (t *RawYAMLBuilder) ReplaceCustomVars() *RawYAMLBuilder {
	if t.inst.GetSpec().Vars != nil {
		for key, val := range t.inst.GetSpec().Vars {
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

func ValidCustomVars(varString string) error {
	if strings.HasPrefix(varString, "{{") && strings.HasSuffix(varString, "}}") {
		return nil
	}
	return ErrInvalidVars
}
