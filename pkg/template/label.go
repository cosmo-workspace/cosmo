package template

import cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"

type LabelHolder interface {
	GetLabels() map[string]string
	SetLabels(map[string]string)
}

func SetTemplateType(l LabelHolder, tmplType string) {
	labels := l.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[cosmov1alpha1.LabelKeyTemplateType] = tmplType
	l.SetLabels(labels)
}

func GetTemplateType(l LabelHolder) (string, bool) {
	labels := l.GetLabels()
	if labels == nil {
		return "", false
	}

	tmplType, ok := labels[cosmov1alpha1.LabelKeyTemplateType]
	return tmplType, ok
}
