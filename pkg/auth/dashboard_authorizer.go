package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
)

// DashboardAuthorizer authorize with cosmo-dashboard login API
type DashboardAuthorizer struct {
	URL    string
	Client *http.Client
}

func NewDashboardAuthorizer(url string, ca []byte) *DashboardAuthorizer {
	auth := &DashboardAuthorizer{
		URL: url,
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(ca)

	auth.Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}
	return auth
}

func (a *DashboardAuthorizer) Authorize(ctx context.Context, msg AuthRequest) (bool, error) {
	body := new(bytes.Buffer)

	req := dashv1alpha1.LoginRequest{
		Id:       msg.GetId(),
		Password: msg.GetPassword(),
	}
	err := json.NewEncoder(body).Encode(req)
	if err != nil {
		return false, fmt.Errorf("failed to encode auth request: %w", err)
	}

	upstreamReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, a.URL, body)
	resp, err := a.Client.Do(upstreamReq)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || http.StatusMultipleChoices <= resp.StatusCode {
		return false, fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
	return true, nil
}
