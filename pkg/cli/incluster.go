package cli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bufbuild/connect-go"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

const (
	InClusterCAFile   string = "ca.crt"
	InClusterCertFile string = "tls.crt"
)

const (
	ServiceAccountTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

const (
	DefaultDashboardNamespace     = "cosmo-system"
	DefaultDashboardPort          = 8443
	DefaultDashboardInclusterPort = 8080
)

var (
	InClusterDashboardURL = inClusterDashboardURL()
	caURL                 = fmt.Sprintf("%s/%s", inClusterCAServerURL(), InClusterCAFile)
)

func inClusterDashboardURL() string {
	namespace := DefaultDashboardNamespace
	if envNamespace := os.Getenv("COSMO_DASHBOARD_NAMESPACE"); envNamespace != "" {
		namespace = envNamespace
	}

	port := fmt.Sprint(DefaultDashboardPort)
	if envPort := os.Getenv("COSMO_DASHBOARD_PORT"); envPort != "" {
		port = envPort
	}
	return fmt.Sprintf("https://cosmo-dashboard.%s.svc.cluster.local:%s", namespace, port)
}

func inClusterCAServerURL() string {
	namespace := DefaultDashboardNamespace
	if envNamespace := os.Getenv("COSMO_DASHBOARD_NAMESPACE"); envNamespace != "" {
		namespace = envNamespace
	}

	port := fmt.Sprint(DefaultDashboardInclusterPort)
	if envPort := os.Getenv("COSMO_DASHBOARD_INCLUSTER_PORT"); envPort != "" {
		port = envPort
	}
	return fmt.Sprintf("http://cosmo-dashboard.%s.svc.cluster.local:%s", namespace, port)
}

func UseServiceAccount(cfg *Config) bool {
	if _, err := os.Stat(ServiceAccountTokenFile); err != nil {
		return false
	}
	// use service account
	if cfg.UseServiceAccount {
		return true
	}
	// empty config
	if cfg.Endpoint == "" && cfg.Token == "" && cfg.User == "" {
		return true
	}
	return false
}

func ServiceAccountLogin(ctx context.Context, cfg *Config) error {
	ca, err := downloadFile(caURL) // TODO: cache
	if err != nil {
		return fmt.Errorf("serviceAccountLogin: failed to download CA: %w", err)
	}

	httpClient, err := InClusterHTTPClient(ca)
	if err != nil {
		return fmt.Errorf("serviceAccountLogin: failed to create http client: %w", err)
	}

	c, err := NewCosmoDashClient(httpClient, InClusterDashboardURL)
	if err != nil {
		return fmt.Errorf("serviceAccountLogin: failed to create cosmo client: %w", err)
	}

	res, err := c.AuthServiceClient.ServiceAccountLogin(ctx,
		connect.NewRequest(&dashv1alpha1.ServiceAccountLoginRequest{
			Token: mustGetServiceAccountTokenFromFile(),
		}))
	if err != nil {
		return fmt.Errorf("serviceAccountLogin: api error: %w", err)
	}
	cfg.Token = ExtractSessionToken(res)
	cfg.Endpoint = InClusterDashboardURL
	cfg.User = res.Msg.UserName
	cfg.UseServiceAccount = true
	cfg.SetCACert(ca)

	return cfg.Save()
}

func InClusterHTTPClient(ca []byte) (*http.Client, error) {
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(ca); !ok {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}
	client := &http.Client{
		Transport: transport,
	}
	return client, nil
}

func downloadFile(fileURL string) ([]byte, error) {
	res, err := http.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get incluster ca: %w", err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read incluster ca: %w", err)
	}

	return data, nil
}

func mustGetServiceAccountTokenFromFile() string {
	token, err := os.ReadFile(ServiceAccountTokenFile)
	if err != nil {
		panic(err)
	}
	return string(token)
}
