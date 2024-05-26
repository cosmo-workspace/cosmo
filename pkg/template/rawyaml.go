package template

import (
	"errors"
	"fmt"
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
}

func NewRawYAMLBuilder(rawYaml string) *RawYAMLBuilder {
	return &RawYAMLBuilder{
		rawYaml: rawYaml,
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
			return nil, fmt.Errorf("failed to parse yaml: %w: \n--- YAML ---\n%v", err, v)
		}
		resources = append(resources, *obj)
	}
	return resources, nil
}

func (t *RawYAMLBuilder) ReplaceDefaultVars(inst cosmov1alpha1.InstanceObject) *RawYAMLBuilder {
	t.rawYaml = strings.ReplaceAll(t.rawYaml, DefaultVarsInstance, inst.GetName())
	t.rawYaml = strings.ReplaceAll(t.rawYaml, DefaultVarsTemplate, inst.GetSpec().Template.Name)

	if inst.GetScope() == meta.RESTScopeNamespace {
		t.rawYaml = strings.ReplaceAll(t.rawYaml, DefaultVarsNamespace, inst.GetNamespace())
	}
	return t
}

func (t *RawYAMLBuilder) ReplaceCustomVars(inst cosmov1alpha1.InstanceObject) *RawYAMLBuilder {
	if inst.GetSpec().Vars != nil {
		for key, val := range inst.GetSpec().Vars {
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
