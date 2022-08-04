package kosmo

import (
	"context"

	"k8s.io/apimachinery/pkg/types"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
)

func (c *Client) ListTemplates(ctx context.Context) ([]cosmov1alpha1.Template, error) {
	tmplList := cosmov1alpha1.TemplateList{}

	err := c.List(ctx, &tmplList)
	return tmplList.Items, err
}

func (c *Client) ListWorkspaceTemplates(ctx context.Context) ([]cosmov1alpha1.Template, error) {
	log := clog.FromContext(ctx).WithCaller()
	if tmpls, err := kubeutil.ListTemplatesByType(ctx, c, []string{wsv1alpha1.TemplateTypeWorkspace}); err != nil {
		log.Error(err, "failed to list WorkspaceTemplates")
		return nil, NewInternalServerError("failed to list WorkspaceTemplates", err)
	} else {
		return tmpls, nil
	}
}

func (c *Client) ListUserAddonTemplates(ctx context.Context) ([]cosmov1alpha1.TemplateObject, error) {
	log := clog.FromContext(ctx).WithCaller()
	if tmpls, err := kubeutil.ListTemplateObjectsByType(ctx, c, []string{wsv1alpha1.TemplateTypeUserAddon}); err != nil {
		log.Error(err, "failed to list UserAddon Templates")
		return nil, NewInternalServerError("failed to list UserAddon Templates", err)
	} else {
		return tmpls, nil
	}
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
