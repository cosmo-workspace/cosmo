package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"

	"github.com/bufbuild/connect-go"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
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

	client := dashboardv1alpha1connect.NewAuthServiceClient(a.Client, a.URL)

	req := dashv1alpha1.LoginRequest{
		UserName: msg.GetUserName(),
		Password: msg.GetPassword(),
	}

	_, err := client.Login(ctx, connect.NewRequest(&req))
	if err != nil {
		return false, err
	}

	return true, nil
}
