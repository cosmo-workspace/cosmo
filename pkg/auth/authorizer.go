package auth

import (
	"context"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
)

type Authorizer interface {
	Authorize(ctx context.Context, req dashv1alpha1.LoginRequest) (bool, error)
}
