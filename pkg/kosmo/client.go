package kosmo

import (
	"context"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
)

type Client struct {
	client.Client
}

func NewClient(c client.Client) Client {
	return Client{Client: c}
}

func NewClientByRestConfig(cfg *rest.Config, scheme *runtime.Scheme) (Client, error) {
	clientOptions := client.Options{Scheme: scheme}
	client, err := client.New(cfg, clientOptions)
	if err != nil {
		return Client{}, err
	}

	return NewClient(client), nil
}

func (c *Client) GetInstance(ctx context.Context, name, namespace string) (*cosmov1alpha1.Instance, error) {
	inst := cosmov1alpha1.Instance{}
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}

	if err := c.Get(ctx, key, &inst); err != nil {
		return nil, err
	}
	return &inst, nil
}

func (c *Client) GetClusterInstance(ctx context.Context, name string) (*cosmov1alpha1.ClusterInstance, error) {
	inst := cosmov1alpha1.ClusterInstance{}
	key := types.NamespacedName{
		Name: name,
	}

	if err := c.Get(ctx, key, &inst); err != nil {
		return nil, err
	}
	return &inst, nil
}

func (c *Client) ListInstances(ctx context.Context, namespace string) ([]cosmov1alpha1.Instance, error) {
	instList := cosmov1alpha1.InstanceList{}
	opts := &client.ListOptions{Namespace: namespace}

	err := c.List(ctx, &instList, opts)
	return instList.Items, err
}

func (c *Client) ListTemplates(ctx context.Context) ([]cosmov1alpha1.Template, error) {
	tmplList := cosmov1alpha1.TemplateList{}

	err := c.List(ctx, &tmplList)
	return tmplList.Items, err
}

func (c *Client) ListTemplatesByType(ctx context.Context, tmplTypes []string) ([]cosmov1alpha1.Template, error) {
	tmplList := cosmov1alpha1.TemplateList{}

	ls := labels.NewSelector()
	req, _ := labels.NewRequirement(cosmov1alpha1.TemplateLabelKeyType, selection.In, tmplTypes)
	ls = ls.Add(*req)

	opts := &client.ListOptions{
		LabelSelector: ls,
	}

	err := c.List(ctx, &tmplList, opts)
	return tmplList.Items, err
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

func (c *Client) GetClusterTemplate(ctx context.Context, tmplName string) (*cosmov1alpha1.ClusterTemplate, error) {
	tmpl := cosmov1alpha1.ClusterTemplate{}

	key := types.NamespacedName{
		Name: tmplName,
	}

	if err := c.Get(ctx, key, &tmpl); err != nil {
		return nil, err
	}
	return &tmpl, nil
}
