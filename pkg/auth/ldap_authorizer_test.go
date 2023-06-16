package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"testing"
)

type IdPass struct {
	id       string
	password string
}

func NewIdPass(id, password string) *IdPass { return &IdPass{id: id, password: password} }
func (a *IdPass) GetUserName() string       { return a.id }
func (a *IdPass) GetPassword() string       { return a.password }

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
		//*********************
		//* connection
		//*********************
		{
			name: "✅ connection: no tls",
			authorizer: func() *LdapAuthorizer {
				return &LdapAuthorizer{
					URL:    "ldap://localhost",
					BindDN: "cn=%s,ou=users,dc=cosmows,dc=dev",
				}
			}(),
			msg:          NewIdPass("ldapuser1", "xxxxxxxx"),
			wantVerified: true,
			wantErr:      false,
		},
		{
			name: "✅ connection: tls + skip verify",
			authorizer: func() *LdapAuthorizer {
				return &LdapAuthorizer{
					URL:       "ldaps://localhost",
					BindDN:    "cn=%s,ou=users,dc=cosmows,dc=dev",
					TlsConfig: &tls.Config{InsecureSkipVerify: true},
				}
			}(),
			msg:          NewIdPass("ldapuser1", "xxxxxxxx"),
			wantVerified: true,
			wantErr:      false,
		},

		{
			name: "✅ connection: tls + verify",
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
				return &LdapAuthorizer{
					URL:       "ldaps://localhost",
					BindDN:    "cn=%s,ou=users,dc=cosmows,dc=dev",
					TlsConfig: &tls.Config{InsecureSkipVerify: false, RootCAs: certPool, ServerName: "localhost:636"},
				}
			}(),
			msg:          NewIdPass("ldapuser1", "xxxxxxxx"),
			wantVerified: true,
			wantErr:      false,
		},
		{
			name: "✅ connection: start tls",
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
				return &LdapAuthorizer{
					URL:       "ldap://localhost",
					StartTLS:  true,
					BindDN:    "cn=%s,ou=users,dc=cosmows,dc=dev",
					TlsConfig: &tls.Config{InsecureSkipVerify: false, RootCAs: certPool, ServerName: "localhost:636"},
				}
			}(),
			msg:          NewIdPass("ldapuser1", "xxxxxxxx"),
			wantVerified: true,
			wantErr:      false,
		},
		{
			name: "❌ connection: dialURL error",
			authorizer: func() *LdapAuthorizer {
				return &LdapAuthorizer{
					URL:    "ldaplocalhost",
					BindDN: "cn=%s,ou=users,dc=cosmows,dc=dev",
				}
			}(),
			msg:          NewIdPass("ldapuser1", "xxxxxxxx"),
			wantVerified: false,
			wantErr:      true,
		},
		{
			name: "❌ connection: start tls error",
			authorizer: func() *LdapAuthorizer {
				return &LdapAuthorizer{
					URL:       "ldaps://localhost",
					StartTLS:  true,
					BindDN:    "cn=%s,ou=users,dc=cosmows,dc=dev",
					TlsConfig: &tls.Config{InsecureSkipVerify: true},
				}
			}(),
			msg:          NewIdPass("ldapuser1", "xxxxxxxx"),
			wantVerified: false,
			wantErr:      true,
		},
		//*********************
		//* bind mode
		//*********************
		{
			name: "✅ bind mode: admin",
			authorizer: func() *LdapAuthorizer {
				return &LdapAuthorizer{
					URL:    "ldap://localhost",
					BindDN: "cn=%s,dc=cosmows,dc=dev",
				}
			}(),
			msg:          NewIdPass("admin", "vvvvvvvv"),
			wantVerified: true,
			wantErr:      false,
		},

		{
			name: "✅ bind mode: verified",
			authorizer: func() *LdapAuthorizer {
				return &LdapAuthorizer{
					URL:    "ldap://localhost",
					BindDN: "cn=%s,ou=users,dc=cosmows,dc=dev",
				}
			}(),
			msg:          NewIdPass("ldapuser1", "xxxxxxxx"),
			wantVerified: true,
			wantErr:      false,
		},

		{
			name: "❌ bind mode: not verified",
			authorizer: func() *LdapAuthorizer {
				return &LdapAuthorizer{
					URL:    "ldap://localhost",
					BindDN: "cn=%s,ou=users,dc=cosmows,dc=dev",
				}
			}(),
			msg:          NewIdPass("ldapuser1", "hogehoge"),
			wantVerified: false,
			wantErr:      true,
		},

		//*********************
		//* search mode
		//*********************
		{
			name: "✅ search mode: verified",
			authorizer: func() *LdapAuthorizer {
				a := &LdapAuthorizer{
					URL:                "ldap://localhost",
					SearchBaseDN:       "ou=users,dc=cosmows,dc=dev",
					SearchBindDN:       "cn=admin,dc=cosmows,dc=dev",
					SearchBindPassword: "vvvvvvvv",
					SearchFilter:       "(uid=%s)",
				}
				return a
			}(),
			msg:          NewIdPass("ldapuser1", "xxxxxxxx"),
			wantVerified: true,
			wantErr:      false,
		},
		{
			name: "❌ search mode: not verified",
			authorizer: func() *LdapAuthorizer {
				a := &LdapAuthorizer{
					URL:                "ldap://localhost",
					SearchBaseDN:       "ou=users,dc=cosmows,dc=dev",
					SearchBindDN:       "cn=admin,dc=cosmows,dc=dev",
					SearchBindPassword: "vvvvvvvv",
					SearchFilter:       "(uid=%s)",
				}
				return a
			}(),
			msg:          NewIdPass("ldapuser1", "hogehoge"),
			wantVerified: false,
			wantErr:      true,
		},
		{
			name: "❌ search mode: bind fail",
			authorizer: func() *LdapAuthorizer {
				a := &LdapAuthorizer{
					URL:                "ldap://localhost",
					SearchBaseDN:       "ou=users,dc=cosmows,dc=dev",
					SearchBindDN:       "xxx",
					SearchBindPassword: "vvvvvvvv",
					SearchFilter:       "(uid=%s)",
				}
				return a
			}(),
			msg:          NewIdPass("ldapuser1", "hogehoge"),
			wantVerified: false,
			wantErr:      true,
		},
		{
			name: "❌ search mode: UnauthenticatedBind",
			authorizer: func() *LdapAuthorizer {
				a := &LdapAuthorizer{
					URL:          "ldap://localhost",
					SearchBaseDN: "ou=users,dc=cosmows,dc=dev",
					SearchFilter: "(uid=%s)",
				}
				return a
			}(),
			msg:          NewIdPass("ldapuser1", "hogehoge"),
			wantVerified: false,
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			gotVerified, err := tt.authorizer.Authorize(ctx, tt.msg)
			if err != nil {
				fmt.Println(err.Error())
			}
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
