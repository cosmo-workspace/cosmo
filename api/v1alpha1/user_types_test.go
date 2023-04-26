package v1alpha1

import "testing"

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
