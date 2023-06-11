package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"testing"

	dashboardv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

//  This test is commented out because it requires an ldap server.
//  To run it, follow the steps below.
//	  cd hack/local-run-test
//	  make install-openldap
//    cd bin;../../download-certs.sh openldap-cert cosmo-system
//    remove "XXX" from test function name

func XXXTestLdapAuthorizer_Authorize(t *testing.T) {
	tests := []struct {
		name         string
		authorizer   *LdapAuthorizer
		msg          AuthRequest
		wantVerified bool
		wantErr      bool
	}{
		{
			name: "admin",
			authorizer: func() *LdapAuthorizer {
				a, _ := NewLdapAuthorizer("ldap://localhost", "dc=cosmows,dc=dev", "cn", nil, false)
				return a
			}(),
			msg: &dashboardv1alpha1.LoginRequest{
				UserName: "admin",
				Password: "vvvvvvvv",
			},
			wantVerified: true,
			wantErr:      false,
		},

		{
			name: "not tls",
			authorizer: func() *LdapAuthorizer {
				a, _ := NewLdapAuthorizer("ldap://localhost", "ou=users,dc=cosmows,dc=dev", "cn", nil, false)
				return a
			}(),
			msg: &dashboardv1alpha1.LoginRequest{
				UserName: "ldapuser1",
				Password: "xxxxxxxx",
			},
			wantVerified: true,
			wantErr:      false,
		},

		{
			name: "tls + skip verify",
			authorizer: func() *LdapAuthorizer {
				tlsConfig := &tls.Config{
					InsecureSkipVerify: true,
				}
				a, _ := NewLdapAuthorizer("ldaps://localhost", "ou=users,dc=cosmows,dc=dev", "cn", tlsConfig, false)
				return a
			}(),
			msg: &dashboardv1alpha1.LoginRequest{
				UserName: "ldapuser1",
				Password: "xxxxxxxx",
			},
			wantVerified: true,
			wantErr:      false,
		},

		{
			name: "tls + verify",
			authorizer: func() *LdapAuthorizer {
				caCert, err := os.ReadFile("../../hack/local-run-test/bin/ca.crt")
				if err != nil {
					return nil
				}
				certPool, err := x509.SystemCertPool()
				if err != nil {
					fmt.Printf("%v", err)
					certPool = x509.NewCertPool()
				}
				certPool.AppendCertsFromPEM(caCert)
				tlsConfig := &tls.Config{
					InsecureSkipVerify: false,
					RootCAs:            certPool,
					ServerName:         "localhost:636",
				}
				a, _ := NewLdapAuthorizer("ldaps://localhost", "ou=users,dc=cosmows,dc=dev", "cn", tlsConfig, false)
				return a
			}(),
			msg: &dashboardv1alpha1.LoginRequest{
				UserName: "ldapuser1",
				Password: "xxxxxxxx",
			},
			wantVerified: true,
			wantErr:      false,
		},

		{
			name: "start tls",
			authorizer: func() *LdapAuthorizer {
				caCert, err := os.ReadFile("../../hack/local-run-test/bin/ca.crt")
				if err != nil {
					return nil
				}
				certPool, err := x509.SystemCertPool()
				if err != nil {
					fmt.Printf("%v", err)
					certPool = x509.NewCertPool()
				}
				certPool.AppendCertsFromPEM(caCert)
				tlsConfig := &tls.Config{
					InsecureSkipVerify: false,
					RootCAs:            certPool,
					ServerName:         "localhost",
				}
				a, _ := NewLdapAuthorizer("ldap://localhost", "ou=users,dc=cosmows,dc=dev", "cn", tlsConfig, true)
				return a
			}(),
			msg: &dashboardv1alpha1.LoginRequest{
				UserName: "ldapuser1",
				Password: "xxxxxxxx",
			},
			wantVerified: true,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			gotVerified, err := tt.authorizer.Authorize(ctx, tt.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("LdapAuthorizer.Authorize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotVerified != tt.wantVerified {
				t.Errorf("LdapAuthorizer.Authorize() = %v, want %v", gotVerified, tt.wantVerified)
			}
		})
	}
}
