package kubeutil

import (
	"context"
	"sort"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

func ListTemplateObjects(ctx context.Context, c client.Client) ([]cosmov1alpha1.TemplateObject, error) {
	tmpls := cosmov1alpha1.TemplateList{}
	if err := c.List(ctx, &tmpls); err != nil {
		return nil, err
	}
	ctmpls := cosmov1alpha1.ClusterTemplateList{}
	if err := c.List(ctx, &ctmpls); err != nil {
		return nil, err
	}

	t := make([]cosmov1alpha1.TemplateObject, 0, len(tmpls.Items)+len(ctmpls.Items))
	for _, v := range tmpls.Items {
		t = append(t, v.DeepCopy())
	}
	for _, v := range ctmpls.Items {
		t = append(t, v.DeepCopy())
	}
	return t, nil
}

func ListTemplateObjectsByType(ctx context.Context, c client.Client, tmplTypes []string) ([]cosmov1alpha1.TemplateObject, error) {

	tmpls, err := ListTemplatesByType(ctx, c, tmplTypes)
	if err != nil {
		return nil, err
	}

	ctmpls, err := ListClusterTemplatesByType(ctx, c, tmplTypes)
	if err != nil {
		return nil, err
	}

	t := make([]cosmov1alpha1.TemplateObject, 0, len(tmpls)+len(ctmpls))
	for _, v := range tmpls {
		t = append(t, v.DeepCopy())
	}
	for _, v := range ctmpls {
		t = append(t, v.DeepCopy())
	}
	return t, nil
}

func ListTemplatesByType(ctx context.Context, c client.Client, tmplTypes []string) ([]cosmov1alpha1.Template, error) {
	tmplList := cosmov1alpha1.TemplateList{}

	req, _ := labels.NewRequirement(cosmov1alpha1.TemplateLabelKeyType, selection.In, tmplTypes)
	opts := &client.ListOptions{
		LabelSelector: labels.NewSelector().Add(*req),
	}

	if err := c.List(ctx, &tmplList, opts); err != nil {
		return nil, err
	}
	sort.Slice(tmplList.Items, func(i, j int) bool { return tmplList.Items[i].Name < tmplList.Items[j].Name })
	return tmplList.Items, nil
}

func ListClusterTemplatesByType(ctx context.Context, c client.Client, tmplTypes []string) ([]cosmov1alpha1.ClusterTemplate, error) {
	tmplList := cosmov1alpha1.ClusterTemplateList{}

	req, _ := labels.NewRequirement(cosmov1alpha1.TemplateLabelKeyType, selection.In, tmplTypes)
	opts := &client.ListOptions{
		LabelSelector: labels.NewSelector().Add(*req),
	}

	if err := c.List(ctx, &tmplList, opts); err != nil {
		return nil, err
	}
	sort.Slice(tmplList.Items, func(i, j int) bool { return tmplList.Items[i].Name < tmplList.Items[j].Name })
	return tmplList.Items, nil
}
