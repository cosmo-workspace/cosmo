package kosmo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
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
		if apierrs.IsNotFound(err) {
			return nil, NewNotFoundError("workspace is not found", err)
		} else {
			log.Error(err, "failed to get workspace", "namespace", namespace, "workspace", name)
			return nil, NewInternalServerError("failed to get workspace", err)
		}
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
		return nil, NewInternalServerError("failed to list workspaces", err)
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
		return nil, NewBadRequestError("failed to get workspace config in template", err)
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
		if apierrs.IsAlreadyExists(err) {
			return nil, NewIsAlreadyExistsError("Workspace already exists", err)
		} else {
			log.Error(err, "failed to create workspace", "username", username, "workspace", ws.Name, "template", tmplName, "vars", fmt.Sprintf("%v", vars))
			return nil, NewInternalServerError("failed to create workspace", err)
		}
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
		return nil, NewInternalServerError("failed to delete workspace", err)
	}
	return ws, nil
}

type UpdateWorkspaceOpts struct {
	Replicas *int64
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

	if equality.Semantic.DeepEqual(before, ws) {
		return nil, NewBadRequestError("no change", nil)
	}

	if err := c.Update(ctx, ws); err != nil {
		message := "failed to update workspace"
		log.Error(err, message, "username", username, "workspace", ws.Name)
		return nil, NewInternalServerError(message, err)
	}

	return ws, nil
}

func (c *Client) AddNetworkRule(ctx context.Context, name, username,
	networkRuleName string, portNumber int32, group *string, httpPath string, public bool) (*cosmov1alpha1.NetworkRule, error) {
	log := clog.FromContext(ctx).WithCaller()

	ws, err := c.GetWorkspaceByUserName(ctx, name, username)
	if err != nil {
		return nil, err
	}

	for _, netRule := range ws.Spec.Network {
		if netRule.Name != networkRuleName &&
			reflect.DeepEqual(netRule.Group, group) &&
			netRule.HTTPPath == httpPath {
			message := fmt.Sprintf("group '%s' and http path '%s' is already used", *group, httpPath)
			log.Error(err, message, "username", username, "workspace", ws.Name, "netRuleName", networkRuleName)
			return nil, NewBadRequestError(message, nil)
		}
	}

	before := ws.DeepCopy()

	// upsert
	index := getNetRuleIndex(ws.Spec.Network, networkRuleName)
	if index == -1 {
		index = len(ws.Spec.Network)
		ws.Spec.Network = append(ws.Spec.Network, cosmov1alpha1.NetworkRule{})
	}
	var netRule = &ws.Spec.Network[index]
	netRule.Name = networkRuleName
	netRule.PortNumber = portNumber
	netRule.Group = group
	netRule.HTTPPath = httpPath
	netRule.Public = public

	log.Debug().Info("upserting network rule", "ws", ws.Name, "namespace", ws.Namespace, "netRule", netRule)
	log.DebugAll().PrintObjectDiff(before, ws)

	if equality.Semantic.DeepEqual(before, ws) {
		log.Info("no change", "username", username, "workspace", ws.Name, "netRuleName", networkRuleName)
		return nil, NewBadRequestError("no change", nil)
	}

	if err := c.Update(ctx, ws); err != nil {
		message := "failed to upsert network rule"
		log.Error(err, message, "username", username, "workspace", ws.Name, "netRuleName", networkRuleName)
		return nil, NewInternalServerError(message, err)
	}
	return netRule.DeepCopy(), nil
}

func (c *Client) DeleteNetworkRule(ctx context.Context, name, username, networkRuleName string) (*cosmov1alpha1.NetworkRule, error) {
	log := clog.FromContext(ctx).WithCaller()

	ws, err := c.GetWorkspaceByUserName(ctx, name, username)
	if err != nil {
		return nil, err
	}
	log.DebugAll().Info("GetWorkspace", "ws", ws, "username", username)

	if networkRuleName == ws.Status.Config.ServiceMainPortName {
		return nil, NewBadRequestError("main port cannot be removed", nil)
	}

	index := getNetRuleIndex(ws.Spec.Network, networkRuleName)
	if index == -1 {
		message := fmt.Sprintf("port name %s is not found", networkRuleName)
		log.Info(message, "username", username, "workspace", ws.Name, "netRuleName", networkRuleName)
		return nil, NewBadRequestError(message, nil)
	}

	before := ws.DeepCopy()

	delRule := ws.Spec.Network[index].DeepCopy()
	ws.Spec.Network = ws.Spec.Network[:index+copy(ws.Spec.Network[index:], ws.Spec.Network[index+1:])]

	log.DebugAll().Info("NetworkRule removed", "ws", ws, "username", username, "netRuleName", networkRuleName)
	log.DebugAll().PrintObjectDiff(before, ws)

	if equality.Semantic.DeepEqual(before, ws) {
		return nil, errors.New("no change")
	}

	if err := c.Update(ctx, ws); err != nil {
		message := "failed to remove network rule"
		log.Error(err, message, "username", username, "workspace", ws.Name, "netRuleName", networkRuleName)
		return nil, NewInternalServerError(message, err)
	}
	return delRule, nil
}

func getNetRuleIndex(netRules []cosmov1alpha1.NetworkRule, netRuleName string) int {
	for i, netRule := range netRules {
		if netRule.Name == netRuleName {
			return i
		}
	}
	return -1
}

func (c *Client) GetWorkspaceConfig(ctx context.Context, tmplName string) (cfg cosmov1alpha1.Config, err error) {
	tmpl := &cosmov1alpha1.Template{}
	if err := c.Get(ctx, types.NamespacedName{Name: tmplName}, tmpl); err != nil {
		return cfg, err
	}
	return workspace.ConfigFromTemplateAnnotations(tmpl)
}
