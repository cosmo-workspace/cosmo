package auth

import (
	"context"

	authv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/auth-proxy/v1alpha1"
)

type Authorizer interface {
	Authorize(ctx context.Context, msg *authv1alpha1.LoginRequest) (bool, error)
}
