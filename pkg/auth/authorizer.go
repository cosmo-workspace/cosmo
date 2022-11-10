package auth

import (
	"context"
)

type Authorizer interface {
	Authorize(ctx context.Context, msg AuthRequest) (bool, error)
}

type AuthRequest interface {
	GetPassword() string
	GetUserName() string
}
