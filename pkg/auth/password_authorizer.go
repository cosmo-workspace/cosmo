package auth

import (
	"context"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
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

func (a *PasswordSecretAuthorizer) Authorize(ctx context.Context, req dashv1alpha1.LoginRequest) (bool, error) {
	verified, _, err := password.VerifyPassword(ctx, a.Client, req.Id, []byte(req.Password))
	return verified, err
}
