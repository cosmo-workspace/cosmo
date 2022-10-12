package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"

	authv1alpha1 "github.com/cosmo-workspace/cosmo/api/auth-proxy/v1alpha1"
)

// HTTPAuthorizer authorize with cosmo-dashboard login API
type HTTPAuthorizer struct {
	URL    string
	Client *http.Client
}

func NewHTTPAuthorizer(url string, ca []byte) *HTTPAuthorizer {
	auth := &HTTPAuthorizer{
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

func (a *HTTPAuthorizer) Authorize(ctx context.Context, msg *authv1alpha1.LoginRequest) (bool, error) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(msg)
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
