package kosmo

import (
	"context"
	"sort"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

func (c *Client) ListTemplates(ctx context.Context) ([]cosmov1alpha1.Template, error) {
	tmplList := cosmov1alpha1.TemplateList{}

	err := c.List(ctx, &tmplList)
	return tmplList.Items, err
}

func (c *Client) ListWorkspaceTemplates(ctx context.Context) ([]cosmov1alpha1.Template, error) {
	log := clog.FromContext(ctx).WithCaller()
	if tmpls, err := c.ListTemplatesByType(ctx, []string{wsv1alpha1.TemplateTypeWorkspace}); err != nil {
		log.Error(err, "failed to list WorkspaceTemplates")
		return nil, NewInternalServerError("failed to list WorkspaceTemplates", err)
	} else {
		return tmpls, nil
	}
}

func (c *Client) ListUserAddonTemplates(ctx context.Context) ([]cosmov1alpha1.Template, error) {
	log := clog.FromContext(ctx).WithCaller()
	if tmpls, err := c.ListTemplatesByType(ctx, []string{wsv1alpha1.TemplateTypeUserAddon}); err != nil {
		log.Error(err, "failed to list UserAddon Templates")
		return nil, NewInternalServerError("failed to list UserAddon Templates", err)
	} else {
		return tmpls, nil
	}
}

func (c *Client) ListTemplatesByType(ctx context.Context, tmplTypes []string) ([]cosmov1alpha1.Template, error) {
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

func (c *Client) GetTemplate(ctx context.Context, tmplName string) (*cosmov1alpha1.Template, error) {
	tmpl := cosmov1alpha1.Template{}

	key := types.NamespacedName{
		Name: tmplName,
	}

	if err := c.Get(ctx, key, &tmpl); err != nil {
		return nil, err
	}
	return &tmpl, nil
}
