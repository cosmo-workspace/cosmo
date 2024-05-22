package kubeutil

import (
	"context"
	"errors"
	"fmt"
	"strings"

	v1 "k8s.io/api/authentication/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TokenReviewResult struct {
	ServiceAccount    string
	ServiceAccountUID string
	Namespace         string
	PodName           string
	PodUID            string
}

func TokenReview(ctx context.Context, c client.Client, token string) (*TokenReviewResult, error) {
	t := v1.TokenReview{}
	t.Spec.Token = token

	err := c.Create(ctx, &t)
	if err != nil {
		return nil, err
	}

	if !t.Status.Authenticated {
		return nil, errors.New("not authenticated")
	}

	username := strings.Split(t.Status.User.Username, ":") // system:serviceaccount:cosmo-user-jinu:default
	if len(username) != 4 {
		return nil, fmt.Errorf("invalid username format: %s", username)
	}
	if username[0] != "system" || username[1] != "serviceaccount" {
		return nil, fmt.Errorf("invalid username format: %s", username)
	}

	res := &TokenReviewResult{
		Namespace:         username[2],
		ServiceAccount:    username[3],
		ServiceAccountUID: t.Status.User.UID,
		PodName:           t.Status.User.Extra["authentication.kubernetes.io/pod-name"].String(),
		PodUID:            t.Status.User.Extra["authentication.kubernetes.io/pod-uid"].String(),
	}

	return res, err
}
