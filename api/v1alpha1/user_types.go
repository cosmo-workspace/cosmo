package v1alpha1

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}

const (
	// UserPasswordSecretName is a secret name for password secret
	UserPasswordSecretName = "password"
	// UserPasswordSecretDataKeyUserPasswordSecret is a secret data key for hashed password value
	UserPasswordSecretDataKeyUserPasswordSecret = "password"
	// UserPasswordSecretDataKeyUserPasswordSalt is a secret data key for hashed password salt
	UserPasswordSecretDataKeyUserPasswordSalt = "salt"
	// UserPasswordSecretAnnKeyUserPasswordIfDefault is a secret annotation key to notify if password is default
	UserPasswordSecretAnnKeyUserPasswordIfDefault = "cosmo-workspace.github.io/default-password"
)

// NamespaceLabelKeyUserName is a label key on namespace created b User
const NamespaceLabelKeyUserName = "cosmo-workspace.github.io/user"

// UserAddonTemplateAnnKeyDefault is an annotation key on UserAddon Template to notify controller to create the UserAddon for all Users
const UserAddonTemplateAnnKeyDefaultUserAddon = "useraddon.cosmo-workspace.github.io/default"

// Var for user addon
const TemplateVarUserName = "{{USER_NAME}}"

const UserNamespacePrefix = "cosmo-user-"

func UserNamespace(username string) string {
	return UserNamespacePrefix + username
}

func UserNameByNamespace(namespace string) string {
	if !strings.HasPrefix(namespace, UserNamespacePrefix) {
		return ""
	}
	return strings.TrimPrefix(namespace, UserNamespacePrefix)
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope="Cluster"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Roles",type=string,JSONPath=`.spec.roles[*].name`
// +kubebuilder:printcolumn:name="AuthType",type=string,JSONPath=`.spec.authType`
// +kubebuilder:printcolumn:name="Namespace",type=string,JSONPath=`.status.namespace.name`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Addons",type=string,JSONPath=`.spec.addons[*].template.name`
// User is the Schema for the workspaces API
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// UserList contains a list of User
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []User `json:"items"`
}

type UserSpec struct {
	DisplayName string       `json:"displayName,omitempty"`
	Roles       []UserRole   `json:"roles,omitempty"`
	AuthType    UserAuthType `json:"authType,omitempty"`
	Addons      []UserAddon  `json:"addons,omitempty"`
}

type UserStatus struct {
	Phase     corev1.NamespacePhase `json:"phase,omitempty"`
	Namespace ObjectRef             `json:"namespace,omitempty"`
	Addons    []ObjectRef           `json:"addons,omitempty"`
}

type UserAddon struct {
	Template UserAddonTemplateRef `json:"template,omitempty"`
	Vars     map[string]string    `json:"vars,omitempty"`
}

// TemplateRef defines template to use in Instance creation
type UserAddonTemplateRef struct {
	// +kubebuilder:validation:Required
	Name          string `json:"name"`
	ClusterScoped bool   `json:"clusterScoped,omitempty"`
}

type UserRole struct {
	Name string `json:"name"`
}

const (
	UserAdminRole string = "cosmo-admin"
)

func (u User) IsAdmin() bool {
	for _, role := range u.Spec.Roles {
		if role.Name == UserAdminRole {
			return true
		}
	}
	return false
}

// +kubebuilder:validation:enum=kosmo-secret
// UserAuthType enums
type UserAuthType string

const (
	UserAuthTypePasswordSecert UserAuthType = "kosmo-secret" // TODO change password-secret
	// TODO
	// UserAuthTypeLDAP    = "ldap"
	// UserAuthTypeOIDC    = "oidc"
	// UserAuthTypeWebhook = "webhook"
)

func (t UserAuthType) IsValid() bool {
	switch t {
	case UserAuthTypePasswordSecert:
		return true
	default:
		return false
	}
}

func (t UserAuthType) String() string {
	return string(t)
}
