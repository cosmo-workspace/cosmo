package dashboard

import (
	"context"
	"errors"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_userAuthentication(t *testing.T) {
	type args struct {
		caller   *cosmov1alpha1.User
		userName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "pass",
			args: args{
				caller: &cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "harry",
					},
				},
				userName: "harry",
			},
			wantErr: false,
		},
		{
			name: "forbidden",
			args: args{
				caller: &cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "harry",
					},
				},
				userName: "harr",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newContextWithCaller(context.TODO(), tt.args.caller)
			if err := userAuthentication(ctx, tt.args.userName); (err != nil) != tt.wantErr {
				t.Errorf("userAuthentication() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_adminAuthentication(t *testing.T) {
	type args struct {
		caller            *cosmov1alpha1.User
		customAuthenFuncs []func(callerGroupRoleMap map[string]string) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "pass with privileged",
			args: args{
				caller: &cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "harry",
					},
					Spec: cosmov1alpha1.UserSpec{
						Roles: []cosmov1alpha1.UserRole{cosmov1alpha1.PrivilegedRole},
					},
				},
				customAuthenFuncs: nil,
			},
			wantErr: false,
		},
		{
			name: "deny for no roles",
			args: args{
				caller: &cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "harry",
					},
				},
				customAuthenFuncs: nil,
			},
			wantErr: true,
		},
		{
			name: "deny for not admin",
			args: args{
				caller: &cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "harry",
					},
					Spec: cosmov1alpha1.UserSpec{
						Roles: []cosmov1alpha1.UserRole{{Name: "gryffindor-dev"}},
					},
				},
				customAuthenFuncs: nil,
			},
			wantErr: true,
		},
		{
			name: "pass for custom func passed",
			args: args{
				caller: &cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "harry",
					},
					Spec: cosmov1alpha1.UserSpec{
						Roles: []cosmov1alpha1.UserRole{{Name: "gryffindor-admin"}},
					},
				},
				customAuthenFuncs: []func(callerGroupRoleMap map[string]string) error{
					func(callerGroupRoleMap map[string]string) error {
						return nil
					},
				},
			},
			wantErr: false,
		},
		{
			name: "pass for custom funcs are passed",
			args: args{
				caller: &cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "harry",
					},
					Spec: cosmov1alpha1.UserSpec{
						Roles: []cosmov1alpha1.UserRole{{Name: "gryffindor-admin"}},
					},
				},
				customAuthenFuncs: []func(callerGroupRoleMap map[string]string) error{
					func(callerGroupRoleMap map[string]string) error {
						if callerGroupRoleMap["gryffindor"] == cosmov1alpha1.AdminRoleName {
							return nil
						}
						return errors.New("no gryffindor admin")
					},
				},
			},
			wantErr: false,
		},
		{
			name: "deny for custom func failed",
			args: args{
				caller: &cosmov1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{
						Name: "draco",
					},
					Spec: cosmov1alpha1.UserSpec{
						Roles: []cosmov1alpha1.UserRole{{Name: "slytherin-admin"}},
					},
				},
				customAuthenFuncs: []func(callerGroupRoleMap map[string]string) error{
					// pass
					func(callerGroupRoleMap map[string]string) error {
						if callerGroupRoleMap["slytherin"] == cosmov1alpha1.AdminRoleName {
							return nil
						}
						return errors.New("no slytherin admin")
					},
					// fail
					func(callerGroupRoleMap map[string]string) error {
						if callerGroupRoleMap["gryffindor"] == cosmov1alpha1.AdminRoleName {
							return nil
						}
						return errors.New("no gryffindor admin")
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newContextWithCaller(context.TODO(), tt.args.caller)
			if err := adminAuthentication(ctx, tt.args.customAuthenFuncs...); (err != nil) != tt.wantErr {
				t.Errorf("adminAuthentication() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateCallerHasAdminForAllRoles(t *testing.T) {
	type args struct {
		callerGroupRoleMap map[string]string
		tryRoleNames       []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "forbidden with no-admin role",
			args: args{
				callerGroupRoleMap: map[string]string{"gryffindor": "developer"},
				tryRoleNames:       []string{"gryffindor-developer"},
			},
			wantErr: true,
		},
		{
			name: "pass with admin role",
			args: args{
				callerGroupRoleMap: map[string]string{"gryffindor": "admin"},
				tryRoleNames:       []string{"gryffindor-developer"},
			},
			wantErr: false,
		},
		{
			name: "forbidden without matching all trying roles",
			args: args{
				callerGroupRoleMap: map[string]string{"slytherin": "admin"},
				tryRoleNames:       []string{"slytherin-developer", "gryffindor-developer"},
			},
			wantErr: true,
		},
		{
			name: "pass with matching all trying roles",
			args: args{
				callerGroupRoleMap: map[string]string{"slytherin": "admin", "gryffindor": "admin"},
				tryRoleNames:       []string{"slytherin-developer", "gryffindor-developer"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := validateCallerHasAdminForAllRoles(tt.args.tryRoleNames)
			if err := f(tt.args.callerGroupRoleMap); (err != nil) != tt.wantErr {
				t.Errorf("validateCallerHasAdminForAllRoles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
