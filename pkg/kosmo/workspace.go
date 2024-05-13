package kosmo

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/workspace"
)

var (
	ErrNoItems = errors.New("no items")
)

func (c *Client) GetWorkspaceByUserName(ctx context.Context, name, username string) (*cosmov1alpha1.Workspace, error) {
	return c.GetWorkspace(ctx, name, cosmov1alpha1.UserNamespace(username))
}

func (c *Client) GetWorkspace(ctx context.Context, name, namespace string) (*cosmov1alpha1.Workspace, error) {
	log := clog.FromContext(ctx).WithCaller()

	if _, err := c.GetUser(ctx, cosmov1alpha1.UserNameByNamespace(namespace)); err != nil {
		return nil, err
	}

	ws := cosmov1alpha1.Workspace{}
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	if err := c.Get(ctx, key, &ws); err != nil {
		log.Error(err, "failed to get workspace", "namespace", namespace, "workspace", name)
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}
	log.DebugAll().Info("GetWorkspace", "ws", ws, "namespace", namespace)
	return &ws, nil
}

func (c *Client) ListWorkspacesByUserName(ctx context.Context, username string) ([]cosmov1alpha1.Workspace, error) {
	return c.ListWorkspaces(ctx, cosmov1alpha1.UserNamespace(username))
}

func (c *Client) ListWorkspaces(ctx context.Context, namespace string) ([]cosmov1alpha1.Workspace, error) {
	log := clog.FromContext(ctx).WithCaller()

	if _, err := c.GetUser(ctx, cosmov1alpha1.UserNameByNamespace(namespace)); err != nil {
		return nil, err
	}

	wsList := cosmov1alpha1.WorkspaceList{}
	opts := &client.ListOptions{Namespace: namespace}

	if err := c.List(ctx, &wsList, opts); err != nil {
		log.Error(err, "failed to list workspaces", "namespace", namespace)
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	sort.Slice(wsList.Items, func(i, j int) bool { return wsList.Items[i].Name < wsList.Items[j].Name })

	return wsList.Items, nil
}

func (c *Client) CreateWorkspace(ctx context.Context, username, wsName, tmplName string, vars map[string]string, opts ...client.CreateOption) (*cosmov1alpha1.Workspace, error) {
	log := clog.FromContext(ctx).WithCaller()

	if _, err := c.GetUser(ctx, username); err != nil {
		return nil, err
	}

	cfg, err := c.GetWorkspaceConfig(ctx, tmplName)
	if err != nil {
		log.Error(err, "failed to get workspace config in template", "template", tmplName)
		return nil, fmt.Errorf("failed to get workspace config in template: %w", err)
	}

	ws := &cosmov1alpha1.Workspace{}
	ws.SetName(wsName)
	ws.SetNamespace(cosmov1alpha1.UserNamespace(username))
	ws.Spec = cosmov1alpha1.WorkspaceSpec{
		Template: cosmov1alpha1.TemplateRef{
			Name: tmplName,
		},
		Vars: vars,
	}
	log.Debug().Info("creating workspace", "ws", ws, "dryrun", opts)

	if err := c.Create(ctx, ws, opts...); err != nil {
		log.Error(err, "failed to create workspace", "username", username, "workspace", ws.Name, "template", tmplName, "vars", fmt.Sprintf("%v", vars))
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}
	ws.Status.Phase = "Pending"
	ws.Status.Config = cfg

	return ws, nil
}

func (c *Client) DeleteWorkspace(ctx context.Context, name, username string, opts ...client.DeleteOption) (*cosmov1alpha1.Workspace, error) {
	log := clog.FromContext(ctx).WithCaller()

	ws, err := c.GetWorkspaceByUserName(ctx, name, username)
	if err != nil {
		return nil, err
	}
	if err := c.Delete(ctx, ws, opts...); err != nil {
		log.Error(err, "failed to delete workspace", "username", username, "workspace", name)
		return nil, fmt.Errorf("failed to delete workspace: %w", err)
	}
	return ws, nil
}

type UpdateWorkspaceOpts struct {
	Replicas *int64
	Vars     map[string]string
}

func (c *Client) UpdateWorkspace(ctx context.Context, name, username string, opts UpdateWorkspaceOpts) (*cosmov1alpha1.Workspace, error) {
	log := clog.FromContext(ctx).WithCaller()

	ws, err := c.GetWorkspaceByUserName(ctx, name, username)
	if err != nil {
		return nil, err
	}

	before := ws.DeepCopy()

	if opts.Replicas != nil {
		ws.Spec.Replicas = opts.Replicas
	}
	if opts.Vars != nil {
		ws.Spec.Vars = opts.Vars
	}

	if equality.Semantic.DeepEqual(before, ws) {
		return nil, apierrs.NewBadRequest("no change")
	}

	if opts.Replicas != nil {
		if *opts.Replicas == 0 {
			kubeutil.SetAnnotation(ws, cosmov1alpha1.WorkspaceAnnKeyLastStoppedAt, time.Now().Format(time.RFC3339))
		} else {
			kubeutil.SetAnnotation(ws, cosmov1alpha1.WorkspaceAnnKeyLastStartedAt, time.Now().Format(time.RFC3339))
		}
	}

	if err := c.Update(ctx, ws); err != nil {
		log.Error(err, "failed to update workspace", "username", username, "workspace", ws.Name)
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	return ws, nil
}

func (c *Client) AddNetworkRule(ctx context.Context, name, username string, r cosmov1alpha1.NetworkRule, index int) (*cosmov1alpha1.NetworkRule, error) {
	log := clog.FromContext(ctx).WithCaller()

	ws, err := c.GetWorkspaceByUserName(ctx, name, username)
	if err != nil {
		return nil, err
	}
	before := ws.DeepCopy()

	r.Default()

	// upsert
	if index < 0 || index >= len(ws.Spec.Network) {
		ws.Spec.Network = append(ws.Spec.Network, r)
		log.Debug().Info("insert network rule", "ws", ws.Name, "namespace", ws.Namespace, "netRule", r)
	} else {
		ws.Spec.Network[index] = r
		log.Debug().Info("update network rule", "ws", ws.Name, "namespace", ws.Namespace, "netRule", r)
	}

	log.DebugAll().PrintObjectDiff(before, ws)

	if equality.Semantic.DeepEqual(before, ws) {
		log.Info("no change", "username", username, "workspace", ws.Name, "netRule", r)
		return nil, apierrs.NewBadRequest("no change")
	}

	if err := c.Update(ctx, ws); err != nil {
		log.Error(err, "failed to upsert network rule", "username", username, "workspace", ws.Name, "netRule", r)
		return nil, fmt.Errorf("failed to upsert network rule: %w", err)
	}
	return r.DeepCopy(), nil
}

func (c *Client) DeleteNetworkRule(ctx context.Context, name, username string, index int) (*cosmov1alpha1.NetworkRule, error) {
	log := clog.FromContext(ctx).WithCaller()

	ws, err := c.GetWorkspaceByUserName(ctx, name, username)
	if err != nil {
		return nil, err
	}
	log.DebugAll().Info("GetWorkspace", "ws", ws, "username", username)
	before := ws.DeepCopy()

	if index < 0 || index >= len(ws.Spec.Network) {
		return nil, errors.New("index out of range")
	}

	delRule := ws.Spec.Network[index].DeepCopy()
	ws.Spec.Network = ws.Spec.Network[:index+copy(ws.Spec.Network[index:], ws.Spec.Network[index+1:])]

	log.Debug().Info("NetworkRule removing", "ws", ws, "username", username, "index", index, "netRule", delRule)
	log.DebugAll().PrintObjectDiff(before, ws)

	if equality.Semantic.DeepEqual(before, ws) {
		return nil, errors.New("no change")
	}

	if err := c.Update(ctx, ws); err != nil {
		log.Error(err, "failed to remove network rule", "username", username, "workspace", ws.Name, "index", index, "netRule", delRule)
		return nil, fmt.Errorf("failed to remove network rule: %w", err)
	}
	return delRule, nil
}

func (c *Client) GetWorkspaceConfig(ctx context.Context, tmplName string) (cfg cosmov1alpha1.Config, err error) {
	tmpl := &cosmov1alpha1.Template{}
	if err := c.Get(ctx, types.NamespacedName{Name: tmplName}, tmpl); err != nil {
		return cfg, err
	}
	return workspace.ConfigFromTemplateAnnotations(tmpl)
}
