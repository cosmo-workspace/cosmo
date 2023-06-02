package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUserNamespace(t *testing.T) {
	type args struct {
		username string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "✅ ok",
			args: args{"user1"},
			want: "cosmo-user-user1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UserNamespace(tt.args.username); got != tt.want {
				t.Errorf("UserNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserNameByNamespace(t *testing.T) {
	type args struct {
		namespace string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "✅ valid namespace",
			args: args{"cosmo-user-user1"},
			want: "user1",
		},
		{
			name: "✅ invalid namespace",
			args: args{"xxcosmo-user-user1"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UserNameByNamespace(tt.args.namespace); got != tt.want {
				t.Errorf("UserNameByNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserRole_GetGroupAndRole(t *testing.T) {
	type fields struct {
		Name string
	}
	tests := []struct {
		name      string
		fields    fields
		wantGroup string
		wantRole  string
	}{
		{
			name: "admin",
			fields: fields{
				Name: "cosmo-admin",
			},
			wantGroup: "cosmo",
			wantRole:  AdminRoleName,
		},
		{
			name: "non-admin",
			fields: fields{
				Name: "cosmo-developer",
			},
			wantGroup: "cosmo",
			wantRole:  "developer",
		},
		{
			name: "non-admin: middle",
			fields: fields{
				Name: "cosmo-admin-developer",
			},
			wantGroup: "cosmo-admin",
			wantRole:  "developer",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := UserRole{
				Name: tt.fields.Name,
			}
			gotGroup, gotRole := r.GetGroupAndRole()
			if gotGroup != tt.wantGroup {
				t.Errorf("UserRole.GetGroupAndRole() gotGroup = %v, want %v", gotGroup, tt.wantGroup)
			}
			if gotRole != tt.wantRole {
				t.Errorf("UserRole.GetGroupAndRole() gotRole = %v, want %v", gotRole, tt.wantRole)
			}
		})
	}
}

func TestUser_GetGroupRoleMap(t *testing.T) {
	type fields struct {
		TypeMeta   metav1.TypeMeta
		ObjectMeta metav1.ObjectMeta
		Spec       UserSpec
		Status     UserStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		{
			name: "✅ ok",
			fields: fields{
				Spec: UserSpec{
					Roles: []UserRole{
						{
							Name: "xxx-yyy",
						},
					},
				},
			},
			want: map[string]string{
				"xxx": "yyy",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if got := u.GetGroupRoleMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("User.GetGroupRoleMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasPrivilegedRole(t *testing.T) {
	type args struct {
		roles []UserRole
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "✅ HasPrivilegedRole",
			args: args{
				roles: []UserRole{
					{Name: "cosmo-adminxx"},
					{Name: "cosmo-admin"},
				},
			},
			want: true,
		},
		{
			name: "✅ not HasPrivilegedRole",
			args: args{
				roles: []UserRole{
					{Name: "cosmo-adminxx"},
					{Name: "cosmo-adminyy"},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasPrivilegedRole(tt.args.roles); got != tt.want {
				t.Errorf("HasPrivilegedRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserAuthType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		tr   UserAuthType
		want bool
	}{
		{
			name: "✅ password-secret",
			tr:   "password-secret",
			want: true,
		},
		{
			name: "✅ ldap",
			tr:   "ldap",
			want: true,
		},
		{
			name: "❌ xxxx is invalid",
			tr:   "xxxx",
			want: false,
		},
		{
			name: "❌ '' is invalid",
			tr:   "",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.IsValid(); got != tt.want {
				t.Errorf("UserAuthType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserAuthType_String(t *testing.T) {
	tests := []struct {
		name string
		tr   UserAuthType
		want string
	}{
		{
			name: "✅ password-secret",
			tr:   "password-secret",
			want: "password-secret",
		},
		{
			name: "✅ ldap",
			tr:   "ldap",
			want: "ldap",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.String(); got != tt.want {
				t.Errorf("UserAuthType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
