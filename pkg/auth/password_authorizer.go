package auth

import (
	"context"

	"github.com/cosmo-workspace/cosmo/pkg/auth/password"
	authv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/auth-proxy/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PasswordSecretAuthorizer authorize with cosmo user's password secret
type PasswordSecretAuthorizer struct {
	client.Client
}

func NewPasswordSecretAuthorizer(c client.Client) *PasswordSecretAuthorizer {
	return &PasswordSecretAuthorizer{c}
}

func (a *PasswordSecretAuthorizer) Authorize(ctx context.Context, msg *authv1alpha1.LoginRequest) (bool, error) {
	verified, _, err := password.VerifyPassword(ctx, a.Client, msg.Id, []byte(msg.Password))
	return verified, err
}
