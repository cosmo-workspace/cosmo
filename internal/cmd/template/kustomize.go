package template

import (
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

const (
	DefaultPackagedFile = "packaged.yaml"
)

var (
	SecretFileDefaultMode = int32(420)
)

func NewKustomize(disableNamePrefix bool) *types.Kustomization {
	label := make(map[string]string)
	label[cosmov1alpha1.LabelKeyInstanceName] = template.DefaultVarsInstance
	label[cosmov1alpha1.LabelKeyTemplateName] = template.DefaultVarsTemplate

	kust := &types.Kustomization{
		CommonLabels: label,
		Namespace:    template.DefaultVarsNamespace,
		Resources: []string{
			DefaultPackagedFile,
		},
	}
	if !disableNamePrefix {
		kust.NamePrefix = template.DefaultVarsInstance + "-"
	}
	return kust
}

func addPatchesStrategicMerges(kust *types.Kustomization, files ...types.PatchStrategicMerge) {
	if kust.PatchesStrategicMerge == nil {
		kust.PatchesStrategicMerge = files
	} else {
		kust.PatchesStrategicMerge = append(kust.PatchesStrategicMerge, files...)
	}
}

func StructToYaml(obj interface{}) []byte {
	out, _ := yaml.Marshal(obj)
	return out
}
