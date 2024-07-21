package kosmo

import (
	"context"
	"fmt"
	"sort"
	"time"

	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
)

func (c *Client) GetUser(ctx context.Context, name string) (*cosmov1alpha1.User, error) {
	log := clog.FromContext(ctx).WithCaller()
	user := cosmov1alpha1.User{}

	key := types.NamespacedName{Name: name}
	if err := c.Get(ctx, key, &user); err != nil {
		log.Error(err, "failed to get user", "username", name)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (c *Client) ListUsers(ctx context.Context) ([]cosmov1alpha1.User, error) {
	log := clog.FromContext(ctx).WithCaller()
	userList := cosmov1alpha1.UserList{}
	if err := c.List(ctx, &userList); err != nil {
		log.Error(err, "failed to list users")
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	users := userList.Items
	sort.Slice(users, func(i, j int) bool { return users[i].Name < users[j].Name })

	return users, nil
}

func (c *Client) CreateUser(ctx context.Context, username string, displayName string,
	roles []string, authType string, addons []cosmov1alpha1.UserAddon) (*cosmov1alpha1.User, error) {

	log := clog.FromContext(ctx).WithCaller()
	log.Info("creating user", "username", username, "displayName", displayName, "role", roles, "authType", authType, "addons", addons)

	if displayName == "" {
		displayName = username
	}

	userrole := make([]cosmov1alpha1.UserRole, 0)
	for _, v := range roles {
		if v != "" {
			userrole = append(userrole, cosmov1alpha1.UserRole{Name: v})
		}
	}

	authtype := cosmov1alpha1.UserAuthType(authType)
	if authtype != "" && !authtype.IsValid() {
		log.Info("invalid request", "username", username, "authtype", authtype)
		return nil, apierrs.NewBadRequest(fmt.Sprintf("invalid authtype: %s", authType))
	}

	user := &cosmov1alpha1.User{}
	user.SetName(username)
	user.Spec = cosmov1alpha1.UserSpec{
		DisplayName: displayName,
		Roles:       userrole,
		AuthType:    authtype,
		Addons:      addons,
	}

	log.Debug().Info("creating user object", "user", user)

	err := c.Create(ctx, user)
	if err != nil {
		log.Error(err, "failed to create user", "user", username)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (c *Client) GetDefaultPasswordAwait(ctx context.Context, username string) (*string, error) {
	log := clog.FromContext(ctx).WithCaller()

	tk := time.NewTicker(time.Second)
	defer tk.Stop()
	log.Debug().Info("wait for default password creation", "user", username)

	for {
		defaultPassword, err := c.GetDefaultPassword(ctx, username)
		if err == nil {
			tk.Stop()
			return defaultPassword, nil
		}

		select {
		case <-ctx.Done():
			tk.Stop()
			log.Error(err, "reached to timeout in user creation", "user", username)
			return nil, fmt.Errorf("reached to timeout in user creation")
		default:
			<-tk.C
		}
	}
}

func (c *Client) DeleteUser(ctx context.Context, username string) (*cosmov1alpha1.User, error) {
	log := clog.FromContext(ctx).WithCaller()
	user, err := c.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}

	if cosmov1alpha1.KeepResourceDeletePolicy(user) {
		return nil, fmt.Errorf("protected: keep resource policy is set")
	}

	if err := c.Delete(ctx, user); err != nil {
		log.Error(err, "failed to delete user")
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return user, nil
}

type UpdateUserOpts struct {
	DisplayName  *string
	UserRoles    []cosmov1alpha1.UserRole
	UserAddons   []cosmov1alpha1.UserAddon
	DeletePolicy *string
}

func (c *Client) UpdateUser(ctx context.Context, username string, opts UpdateUserOpts) (*cosmov1alpha1.User, error) {
	logr := clog.FromContext(ctx).WithCaller()

	user, err := c.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}

	before := user.DeepCopy()

	if opts.DisplayName != nil {
		user.Spec.DisplayName = *opts.DisplayName
	}

	// `nil` means caller would not like to change roles.
	if opts.UserRoles != nil {
		user.Spec.Roles = opts.UserRoles
	}

	if opts.UserAddons != nil {
		user.Spec.Addons = opts.UserAddons
	}

	if opts.DeletePolicy != nil {
		kubeutil.SetAnnotation(user, cosmov1alpha1.ResourceAnnKeyDeletePolicy, *opts.DeletePolicy)
	}

	if equality.Semantic.DeepEqual(before, user) {
		logr.Debug().Info("no change", "user", before)
		return nil, apierrs.NewBadRequest("no change")
	}
	if before.Spec.DisplayName != user.Spec.DisplayName {
		logr.Debug().Info("name changed", "name", *opts.DisplayName)
	}
	if !equality.Semantic.DeepEqual(before.Spec.Roles, user.Spec.Roles) {
		logr.Debug().Info("role changed", "role", opts.UserRoles)
	}
	if !equality.Semantic.DeepEqual(before.Spec.Addons, user.Spec.Addons) {
		logr.Debug().Info("addons changed", "addons", opts.UserAddons)
	}

	if err := c.Update(ctx, user); err != nil {
		logr.Error(err, "failed to update user", "user", user)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}
