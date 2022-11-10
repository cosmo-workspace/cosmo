package auth

import (
	"context"

	"github.com/cosmo-workspace/cosmo/pkg/auth/password"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PasswordSecretAuthorizer authorize with cosmo user's password secret
type PasswordSecretAuthorizer struct {
	client.Client
}

func NewPasswordSecretAuthorizer(c client.Client) *PasswordSecretAuthorizer {
	return &PasswordSecretAuthorizer{c}
}

func (a *PasswordSecretAuthorizer) Authorize(ctx context.Context, msg AuthRequest) (bool, error) {
	verified, _, err := password.VerifyPassword(ctx, a.Client, msg.GetUserName(), []byte(msg.GetPassword()))
	return verified, err
}
