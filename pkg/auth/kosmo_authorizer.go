package auth

import (
	"context"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

// KosmoSecretAuthorizer authorize with cosmo user's password secret
type KosmoSecretAuthorizer struct {
	kosmo.Client
}

func NewKosmoSecretAuthorizer(c kosmo.Client) *KosmoSecretAuthorizer {
	return &KosmoSecretAuthorizer{c}
}

func (a *KosmoSecretAuthorizer) Authorize(ctx context.Context, req dashv1alpha1.LoginRequest) (bool, error) {
	verified, _, err := a.VerifyPassword(ctx, req.Id, []byte(req.Password))
	return verified, err
}
