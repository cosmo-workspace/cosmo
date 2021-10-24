package kosmo

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
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

func (c *Client) ListWorkspacePods(ctx context.Context, ws wsv1alpha1.Workspace) ([]corev1.Pod, error) {
	var podList corev1.PodList

	ls := labels.NewSelector()
	req, _ := labels.NewRequirement(cosmov1alpha1.LabelKeyInstance, selection.Equals, []string{ws.GetName()})
	ls = ls.Add(*req)

	opts := &client.ListOptions{
		LabelSelector: ls,
		Namespace:     ws.GetNamespace(),
	}
	if err := c.List(ctx, &podList, opts); err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func (c *Client) GetWorkspaceServicesAndIngress(ctx context.Context, ws wsv1alpha1.Workspace) (svc corev1.Service, ing netv1.Ingress, err error) {
	var svcList corev1.ServiceList
	var ingList netv1.IngressList

	ls := labels.NewSelector()
	req, _ := labels.NewRequirement(cosmov1alpha1.LabelKeyInstance, selection.In, []string{ws.GetName()})
	ls = ls.Add(*req)

	opts := &client.ListOptions{
		LabelSelector: ls,
		Namespace:     ws.GetNamespace(),
	}

	if err := c.List(ctx, &svcList, opts); err != nil {
		return svc, ing, err
	}

	if len(svcList.Items) == 0 {
		return svc, ing, errors.New("no services")
	}

	for _, v := range svcList.Items {
		if cosmov1alpha1.EqualInstanceResourceName(ws.GetName(), v.Name, ws.Status.Config.ServiceName) {
			svc = v
		}
	}

	if err := c.List(ctx, &ingList, opts); err != nil {
		return svc, ing, err
	}

	for _, v := range ingList.Items {
		if cosmov1alpha1.EqualInstanceResourceName(ws.GetName(), v.Name, ws.Status.Config.IngressName) {
			ing = v
		}
	}

	return svc, ing, nil
}

func (c *Client) GetWorkspaceConfig(ctx context.Context, tmplName string) (cfg wsv1alpha1.Config, err error) {
	tmpl, err := c.GetTemplate(ctx, tmplName)
	if err != nil {
		return cfg, err
	}
	return wsv1alpha1.ConfigFromTemplateAnnotations(tmpl)
}
