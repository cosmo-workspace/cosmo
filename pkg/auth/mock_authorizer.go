package auth

import (
	"context"
	"errors"
)

type MockAuthorizer struct {
	UserPassMap map[string]string // key: user, value: password
}

func NewMockAuthorizer(userPassMap map[string]string) *MockAuthorizer {
	return &MockAuthorizer{
		UserPassMap: userPassMap,
	}
}

func (a *MockAuthorizer) Authorize(ctx context.Context, msg AuthRequest) (bool, error) {
	if pass, ok := a.UserPassMap[msg.GetUserName()]; !ok || pass != msg.GetPassword() {
		return false, errors.New("[mock] not authorized")
	}
	return true, nil
}
