package kosmo

import (
	"context"
	"sort"
	"time"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (c *Client) GetUser(ctx context.Context, name string) (*wsv1alpha1.User, error) {
	log := clog.FromContext(ctx).WithCaller()
	user := wsv1alpha1.User{}

	key := types.NamespacedName{Name: name}
	if err := c.Get(ctx, key, &user); err != nil {
		if apierrs.IsNotFound(err) {
			return nil, NewNotFoundError("user is not found", err)
		} else {
			log.Error(err, "failed to get user", "id", name)
			return nil, NewInternalServerError("failed to get user", err)
		}
	}
	return &user, nil
}

func (c *Client) ListUsers(ctx context.Context) ([]wsv1alpha1.User, error) {
	log := clog.FromContext(ctx).WithCaller()
	userList := wsv1alpha1.UserList{}
	if err := c.List(ctx, &userList); err != nil {
		log.Error(err, "failed to list users")
		return nil, NewInternalServerError("failed to list users", err)
	}
	users := userList.Items
	sort.Slice(users, func(i, j int) bool { return users[i].Name < users[j].Name })

	return users, nil
}

func (c *Client) CreateUser(ctx context.Context, userId string, displayName string,
	role string, authType string, addons []wsv1alpha1.UserAddon) (*wsv1alpha1.User, error) {

	log := clog.FromContext(ctx).WithCaller()
	log.Info("creating user", "id", userId, "displayName", displayName, "role", role, "authType", authType, "addons", addons)

	if displayName == "" {
		displayName = userId
	}

	userrole := wsv1alpha1.UserRole(role)
	if !userrole.IsValid() {
		log.Info("invalid request", "id", userId, "role", userrole)
		return nil, NewBadRequestError("'userrole' is invalid", nil)
	}

	authtype := wsv1alpha1.UserAuthType(authType)
	if authtype != "" && !authtype.IsValid() {
		log.Info("invalid request", "id", userId, "authtype", authtype)
		return nil, NewBadRequestError("'authtype' is invalid", nil)
	}

	user := &wsv1alpha1.User{}
	user.SetName(userId)
	user.Spec = wsv1alpha1.UserSpec{
		DisplayName: displayName,
		Role:        userrole,
		AuthType:    authtype,
		Addons:      addons,
	}

	log.Debug().Info("creating user object", "user", user)

	err := c.Create(ctx, user)
	if err != nil {
		if apierrs.IsAlreadyExists(err) {
			return nil, NewIsAlreadyExistsError("user already exists", err)
		} else {
			log.Error(err, "failed to create user", "user", userId)
			return nil, NewServiceUnavailableError("failed to create user", err)
		}
	}

	return user, nil
}

func (c *Client) GetDefaultPasswordAwait(ctx context.Context, userId string) (*string, error) {
	log := clog.FromContext(ctx).WithCaller()

	tk := time.NewTicker(time.Second)
	defer tk.Stop()
	log.Debug().Info("wait for default password creation", "user", userId)

	for {
		defaultPassword, err := c.GetDefaultPassword(ctx, userId)
		if err == nil {
			tk.Stop()
			return defaultPassword, nil
		}

		select {
		case <-ctx.Done():
			tk.Stop()
			log.Error(err, "reached to timeout in user creation", "user", userId)
			return nil, NewInternalServerError("reached to timeout in user creation", nil)
		default:
			<-tk.C
		}
	}
}

func (c *Client) DeleteUser(ctx context.Context, userId string) (*wsv1alpha1.User, error) {
	log := clog.FromContext(ctx).WithCaller()
	user, err := c.GetUser(ctx, userId)
	if err != nil {
		return nil, err
	}

	if err := c.Delete(ctx, user); err != nil {
		log.Error(err, "failed to delete user")
		return nil, NewInternalServerError("failed to delete user", err)
	}

	return user, nil
}

type UpdateUserOpts struct {
	DisplayName *string
	UserRole    *string
}

func (c *Client) UpdateUser(ctx context.Context, userId string, opts UpdateUserOpts) (*wsv1alpha1.User, error) {
	logr := clog.FromContext(ctx).WithCaller()

	user, err := c.GetUser(ctx, userId)
	if err != nil {
		return nil, err
	}

	before := user.DeepCopy()

	if opts.DisplayName != nil && *opts.DisplayName != "" {
		user.Spec.DisplayName = *opts.DisplayName
	}
	if opts.UserRole != nil && *opts.UserRole != "-" {
		user.Spec.Role = wsv1alpha1.UserRole(*opts.UserRole)
		if !user.Spec.Role.IsValid() {
			logr.Debug().Info("'userrole' is invalid", "user", userId, "role", *opts.UserRole)
			return nil, NewBadRequestError("'userrole' is invalid", nil)
		}
	}

	if equality.Semantic.DeepEqual(before, user) {
		logr.Debug().Info("no change", "user", before)
		return nil, NewBadRequestError("no change", err)
	}
	if before.Spec.DisplayName != user.Spec.DisplayName {
		logr.Debug().Info("name changed", "name", *opts.DisplayName)
	}
	if before.Spec.Role != user.Spec.Role {
		logr.Debug().Info("role changed", "role", *opts.UserRole)
	}

	if err := c.Update(ctx, user); err != nil {
		logr.Error(err, "failed to update user", "user", user)
		return nil, NewInternalServerError("failed to update user", err)
	}

	return user, nil
}
