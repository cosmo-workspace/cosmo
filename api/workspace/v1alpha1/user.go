package v1alpha1

import (
	"errors"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func UserNamespace(userid string) string {
	return UserNamespacePrefix + userid
}

func UserIDByNamespace(namespace string) string {
	if !strings.HasPrefix(namespace, UserNamespacePrefix) {
		return ""
	}
	return strings.TrimPrefix(namespace, UserNamespacePrefix)
}

// +kubebuilder:object:generate=true
// User is namespace for Workspace. This is not custom resource but converted to Namespace object.
type User struct {
	ID          string                `json:"id"`
	DisplayName string                `json:"displayName,omitempty"`
	Role        UserRole              `json:"role,omitempty"`
	AuthType    UserAuthType          `json:"authType,omitempty"`
	Status      corev1.NamespacePhase `json:"status,omitempty"`
}

// UserRole enums
type UserRole string

const (
	UserAdminRole UserRole = "cosmo-admin"
)

func (r UserRole) IsAdmin() bool {
	return r == UserAdminRole
}

func (r UserRole) IsValid() bool {
	switch r {
	case UserAdminRole:
		return true
	case UserRole(""):
		return true
	default:
		return false
	}
}

func (r UserRole) String() string {
	return string(r)
}

// UserAuthType enums
type UserAuthType string

const (
	UserAuthTypeKosmoSecert UserAuthType = "kosmo-secret"
	// TODO
	// UserAuthTypeLDAP    = "ldap"
	// UserAuthTypeOIDC    = "oidc"
	// UserAuthTypeWebhook = "webhook"
)

func (t UserAuthType) IsValid() bool {
	switch t {
	case UserAuthTypeKosmoSecert:
		return true
	default:
		return false
	}
}

func (t UserAuthType) String() string {
	return string(t)
}

func ConvertUserNamespaceToUser(ns corev1.Namespace) (*User, error) {
	user := User{}

	idOnName := UserIDByNamespace(ns.Name)
	if idOnName == "" {
		return nil, errors.New("not user namespace")
	}

	label := ns.GetLabels()
	if label == nil {
		return nil, errors.New("label not found")
	}
	idOnLabel, ok := label[NamespaceLabelKeyUserID]
	if !ok {
		return nil, errors.New("user id not found in label")
	}

	if idOnName != idOnLabel {
		return nil, errors.New("user id in namespace name does not match user id in the label")
	}
	user.ID = idOnName

	user.Status = ns.Status.Phase

	ann := ns.GetAnnotations()
	if ann == nil {
		return nil, errors.New("annotation not found")
	}

	user.DisplayName, ok = ann[NamespaceAnnKeyUserName]
	if !ok {
		return nil, errors.New("user name not found in annotation")
	}
	role, ok := ann[NamespaceAnnKeyUserRole]
	if !ok {
		return nil, errors.New("user role not found in annotation")
	}
	user.Role = UserRole(role)

	authtype := ann[NamespaceAnnKeyUserAuthType]
	user.AuthType = UserAuthType(authtype)

	if !UserAuthType(user.AuthType).IsValid() {
		user.AuthType = UserAuthTypeKosmoSecert
	}

	return &user, nil
}
