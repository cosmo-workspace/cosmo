package auth

import (
	"context"
	"testing"
	// dashboardv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

func TestLdapAuthorizer_Authorize(t *testing.T) {

	tests := []struct {
		name         string
		authorizer   *LdapAuthorizer
		msg          AuthRequest
		wantVerified bool
		wantErr      bool
	}{
		//--------------------------------------------------------------------
		// Comment out because it accesses an external ldap test server.
		//--------------------------------------------------------------------
		//
		// {
		// 	name: "",
		// 	authorizer: func() (a *LdapAuthorizer) {
		// 		a = NewLdapAuthorizer("ldap://ldap.forumsys.com:389", "dc=example,dc=com", "uid")
		// 		return
		// 	}(),
		// 	msg: &dashboardv1alpha1.LoginRequest{
		// 		UserName: "newton",
		// 		Password: "password",
		// 	},
		// 	wantVerified: true,
		// 	wantErr:      false,
		// },
		// {
		// 	name: "",
		// 	authorizer: func() (a *LdapAuthorizer) {
		// 		a = NewLdapAuthorizer("ldap://ldap.forumsys.com", "dc=example,dc=com", "uid")
		// 		return
		// 	}(),
		// 	msg: &dashboardv1alpha1.LoginRequest{
		// 		UserName: "newton",
		// 		Password: "password",
		// 	},
		// 	wantVerified: true,
		// 	wantErr:      false,
		// },
		// {
		// 	name: "",
		// 	authorizer: func() (a *LdapAuthorizer) {
		// 		a = NewLdapAuthorizer("ldap://ldap.forumsys.com", "dc=example,dc=com", "uid")
		// 		return
		// 	}(),
		// 	msg: &dashboardv1alpha1.LoginRequest{
		// 		UserName: "newton",
		// 		Password: "xxxx",
		// 	},
		// 	wantVerified: false,
		// 	wantErr:      true,
		// },
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
