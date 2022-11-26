package template

import (
	"strconv"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

type LabelHolder interface {
	GetLabels() map[string]string
	SetLabels(map[string]string)
}

func SetTemplateType(l LabelHolder, tmplType string) {
	labels := l.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[cosmov1alpha1.TemplateLabelKeyType] = tmplType
	l.SetLabels(labels)
}

func GetTemplateType(l LabelHolder) (string, bool) {
	labels := l.GetLabels()
	if labels == nil {
		return "", false
	}

	tmplType, ok := labels[cosmov1alpha1.TemplateLabelKeyType]
	return tmplType, ok
}

func IsDisableNamePrefix(tmpl cosmov1alpha1.TemplateObject) bool {
	ann := tmpl.GetAnnotations()
	if ann == nil {
		return false
	}
	val := ann[cosmov1alpha1.TemplateAnnKeyDisableNamePrefix]
	disable, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}
	return disable
}

func IsSkipValidation(tmpl cosmov1alpha1.TemplateObject) bool {
	ann := tmpl.GetAnnotations()
	if ann == nil {
		return false
	}
	val := ann[cosmov1alpha1.TemplateAnnKeySkipValidation]
	skip, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}
	return skip
}
