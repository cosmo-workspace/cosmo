package kosmo

import (
	"context"

	"github.com/cosmo-workspace/cosmo/pkg/auth/password"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

func (c *Client) VerifyPassword(ctx context.Context, userid string, pass []byte) (verified bool, isDefault bool, err error) {
	log := clog.FromContext(ctx).WithCaller()

	if _, err := c.GetUser(ctx, userid); err != nil {
		return false, isDefault, err
	}

	verified, isDefault, err = password.VerifyPassword(ctx, c, userid, pass)
	if err != nil {
		log.Error(err, "failed to verify password", "userid", userid)
		return false, isDefault, NewInternalServerError("failed to verify password", err)
	}
	return verified, isDefault, nil
}

func (c *Client) IsDefaultPassword(ctx context.Context, userid string) (bool, error) {
	isDefault, err := password.IsDefaultPassword(ctx, c, userid)
	if err != nil {
		return false, NewInternalServerError("failed to get password secret", err)
	}
	return isDefault, nil
}

func (c *Client) GetDefaultPassword(ctx context.Context, userid string) (*string, error) {
	p, err := password.GetDefaultPassword(ctx, c, userid)
	if err != nil {
		return nil, NewInternalServerError("failed to get default password", err)
	}
	return p, nil
}

func (c *Client) ResetPassword(ctx context.Context, userid string) error {
	err := password.ResetPassword(ctx, c, userid)
	if err != nil {
		return NewInternalServerError("failed to reset password", err)
	}
	return nil
}

func (c *Client) RegisterPassword(ctx context.Context, userid string, passwd []byte) error {
	err := password.RegisterPassword(ctx, c, userid, passwd)
	if err != nil {
		return NewInternalServerError("failed to register password", err)
	}
	return nil
}
