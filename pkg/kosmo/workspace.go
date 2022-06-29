package kosmo

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/wscfg"
)

var (
	ErrNoItems = errors.New("no items")
)

func (c *Client) GetWorkspaceByUserID(ctx context.Context, name, userid string) (*wsv1alpha1.Workspace, error) {
	return c.GetWorkspace(ctx, name, wsv1alpha1.UserNamespace(userid))
}

func (c *Client) GetWorkspace(ctx context.Context, name, namespace string) (*wsv1alpha1.Workspace, error) {
	ws := wsv1alpha1.Workspace{}
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}

	if err := c.Get(ctx, key, &ws); err != nil {
		return nil, err
	}
	return &ws, nil
}

func (c *Client) ListWorkspacesByUserID(ctx context.Context, userid string) ([]wsv1alpha1.Workspace, error) {
	return c.ListWorkspaces(ctx, wsv1alpha1.UserNamespace(userid))
}

func (c *Client) ListWorkspaces(ctx context.Context, namespace string) ([]wsv1alpha1.Workspace, error) {
	wsList := wsv1alpha1.WorkspaceList{}
	opts := &client.ListOptions{Namespace: namespace}

	err := c.List(ctx, &wsList, opts)
	return wsList.Items, err
}

func (c *Client) GetWorkspaceConfig(ctx context.Context, tmplName string) (cfg wsv1alpha1.Config, err error) {
	tmpl := &cosmov1alpha1.Template{}
	if err := c.Get(ctx, types.NamespacedName{Name: tmplName}, tmpl); err != nil {
		return cfg, err
	}
	return wscfg.ConfigFromTemplateAnnotations(tmpl)
}
