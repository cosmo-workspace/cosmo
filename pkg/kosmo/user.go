package kosmo

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

const (
	DefaultServiceAccount = "default"
)

func (c *Client) CreateUser(ctx context.Context, user *wsv1alpha1.User) (*wsv1alpha1.User, error) {
	log := clog.FromContext(ctx).WithCaller()

	// create namespace
	ns := corev1.Namespace{}
	ns.SetName(wsv1alpha1.UserNamespace(user.ID))

	label := make(map[string]string)
	label[wsv1alpha1.NamespaceLabelKeyUserID] = user.ID
	ns.SetLabels(label)

	ann := make(map[string]string)
	ann[wsv1alpha1.NamespaceAnnKeyUserName] = user.DisplayName
	ann[wsv1alpha1.NamespaceAnnKeyUserRole] = user.Role.String()
	ann[wsv1alpha1.NamespaceAnnKeyUserAuthType] = user.AuthType.String()
	ns.SetAnnotations(ann)

	log.Info("creating namespace", "userid", user.ID, "namespace", ns.Name)
	if err := c.Create(ctx, &ns); err != nil {
		log.Debug().Info("failed to create namespace", "error", err, "ns", ns)
		return user, err
	}

	log.Info("initializing password secret", "userid", user.ID)
	if err := c.ResetPassword(ctx, user.ID); err != nil {
		defer func() {
			log.Debug().Info("deleting namespace", "ns", ns)
			c.DeleteUser(ctx, user.ID)
		}()

		log.Info("failed to reset password", "error", err)
		return user, err
	}

	log.Info("creating role and rolebinding for default serviceaccount", "userid", user.ID, "role", wsv1alpha1.AuthProxyRoleName)
	if err := c.addAuthProxyRoleOnDefaultServiceAccount(ctx, ns.Name); err != nil {
		defer func() {
			log.Debug().Info("deleting namespace", "ns", ns)
			c.DeleteUser(ctx, user.ID)
		}()

		log.Info("failed to create rbacs", "error", err)
		return user, err
	}

	log.Debug().Info("successfully created user", "user", user, "ns", ns)
	return wsv1alpha1.ConvertUserNamespaceToUser(ns)
}

func (c Client) addAuthProxyRoleOnDefaultServiceAccount(ctx context.Context, namespace string) error {
	log := clog.FromContext(ctx).WithCaller()

	role := wsv1alpha1.AuthProxyRole(namespace)
	log.DebugAll().Info("creating role", "role", role)
	if err := c.Create(ctx, &role); err != nil {
		return err
	}

	roleb := wsv1alpha1.AuthProxyRoleBindings(DefaultServiceAccount, namespace)
	log.DebugAll().Info("creating rolebinding", "rolebinding", roleb)
	if err := c.Create(ctx, &roleb); err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdateUser(ctx context.Context, user *wsv1alpha1.User, opts ...client.UpdateOption) (*wsv1alpha1.User, error) {
	ns, err := c.GetUserNamespace(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	if ns.GetAnnotations() == nil {
		return nil, errors.New("not user namespace")
	}
	ns.Annotations[wsv1alpha1.NamespaceAnnKeyUserName] = user.DisplayName
	ns.Annotations[wsv1alpha1.NamespaceAnnKeyUserRole] = user.Role.String()

	if err := c.Update(ctx, ns, opts...); err != nil {
		return nil, err
	}

	return wsv1alpha1.ConvertUserNamespaceToUser(*ns)
}

func (c *Client) DeleteUser(ctx context.Context, userid string, opts ...client.DeleteOption) (*wsv1alpha1.User, error) {
	log := clog.FromContext(ctx).WithCaller()

	ns, err := c.GetUserNamespace(ctx, userid)
	if err != nil {
		return nil, err
	}

	log.Info("deleting namespace", "userid", userid)
	if err := c.Delete(ctx, ns, opts...); err != nil {
		return nil, err
	}

	return wsv1alpha1.ConvertUserNamespaceToUser(*ns)
}

func (c *Client) GetUser(ctx context.Context, userid string) (*wsv1alpha1.User, error) {
	ns, err := c.GetUserNamespace(ctx, userid)
	if err != nil {
		return nil, err
	}
	return wsv1alpha1.ConvertUserNamespaceToUser(*ns)
}

func (c *Client) GetUserNamespace(ctx context.Context, userid string) (*corev1.Namespace, error) {
	ns := corev1.Namespace{}
	key := types.NamespacedName{
		Name: wsv1alpha1.UserNamespace(userid),
	}
	if err := c.Get(ctx, key, &ns); err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", userid, err)
	}
	return &ns, nil
}

func (c *Client) ListUsers(ctx context.Context) ([]wsv1alpha1.User, error) {
	log := clog.FromContext(ctx).WithCaller()

	var nsList corev1.NamespaceList

	ls := labels.NewSelector()
	req, _ := labels.NewRequirement(wsv1alpha1.NamespaceLabelKeyUserID, selection.Exists, nil)
	ls = ls.Add(*req)

	opts := &client.ListOptions{
		LabelSelector: ls,
	}
	if err := c.List(ctx, &nsList, opts); err != nil {
		return nil, err
	}

	if len(nsList.Items) == 0 {
		return nil, ErrNoItems
	}

	users := make([]wsv1alpha1.User, 0)
	for _, ns := range nsList.Items {
		u, err := wsv1alpha1.ConvertUserNamespaceToUser(ns)
		if err != nil {
			log.DebugAll().Info("not user namespace", "namespace", ns.Name, "reason", err)
			continue
		}
		users = append(users, *u)
	}

	return users, nil
}
