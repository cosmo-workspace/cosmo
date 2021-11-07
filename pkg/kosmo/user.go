package kosmo

import (
	"context"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

func (c *Client) GetUser(ctx context.Context, name string) (*wsv1alpha1.User, error) {
	user := wsv1alpha1.User{}

	key := types.NamespacedName{Name: name}
	if err := c.Get(ctx, key, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) ListUsers(ctx context.Context) ([]wsv1alpha1.User, error) {
	userList := wsv1alpha1.UserList{}
	err := c.List(ctx, &userList)
	return userList.Items, err
}
