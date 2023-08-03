package kosmo

import (
	"context"
	"fmt"

	"github.com/cosmo-workspace/cosmo/pkg/auth/password"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

func (c *Client) VerifyPassword(ctx context.Context, username string, pass []byte) (verified bool, isDefault bool, err error) {
	log := clog.FromContext(ctx).WithCaller()

	verified, isDefault, err = password.VerifyPassword(ctx, c, username, pass)
	if err != nil {
		log.Error(err, "failed to verify password", "username", username)
		return false, isDefault, fmt.Errorf("failed to verify password: %w", err)
	}
	return verified, isDefault, nil
}

func (c *Client) IsDefaultPassword(ctx context.Context, username string) (bool, error) {
	isDefault, err := password.IsDefaultPassword(ctx, c, username)
	if err != nil {
		return false, fmt.Errorf("failed to get password secret: %w", err)
	}
	return isDefault, nil
}

func (c *Client) GetDefaultPassword(ctx context.Context, username string) (*string, error) {
	p, err := password.GetDefaultPassword(ctx, c, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get default password: %w", err)
	}
	return p, nil
}

func (c *Client) ResetPassword(ctx context.Context, username string) error {
	err := password.ResetPassword(ctx, c, username)
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}
	return nil
}

func (c *Client) RegisterPassword(ctx context.Context, username string, passwd []byte) error {
	err := password.RegisterPassword(ctx, c, username, passwd)
	if err != nil {
		return fmt.Errorf("failed to register password: %w", err)
	}
	return nil
}
