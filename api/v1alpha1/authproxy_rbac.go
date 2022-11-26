package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacv1apply "k8s.io/client-go/applyconfigurations/rbac/v1"
)

const (
	AuthProxyRoleName               = "cosmo-auth-proxy-role"
	AuthProxyClusterRoleBindingName = "cosmo-auth-proxy-rolebinding"
)

var authProxyPolicyRoles = []rbacv1.PolicyRule{
	{
		APIGroups: []string{GroupVersion.Group},
		Resources: []string{"workspaces"},
		Verbs:     []string{"patch", "update", "get", "list", "watch"},
	},
	{
		APIGroups: []string{GroupVersion.Group},
		Resources: []string{"workspaces/status"},
		Verbs:     []string{"get", "list", "watch"},
	},
	{
		APIGroups: []string{GroupVersion.Group},
		Resources: []string{"instances", "instances/status"},
		Verbs:     []string{"get", "list", "watch"},
	},
	{
		APIGroups: []string{corev1.GroupName},
		Resources: []string{"secrets"},
		Verbs:     []string{"get", "list", "watch"},
	},
	{
		APIGroups: []string{corev1.GroupName},
		Resources: []string{"events"},
		Verbs:     []string{"create", "get", "list", "watch"},
	},
}

func AuthProxyRole(namespace string) rbacv1.Role {
	return rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      AuthProxyRoleName,
			Namespace: namespace,
		},
		Rules: authProxyPolicyRoles,
	}
}

func AuthProxyRoleBindings(sa, namespace string) rbacv1.RoleBinding {
	roleb := rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: rbacv1.GroupName + "/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      AuthProxyRoleName,
			Namespace: namespace,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     AuthProxyRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sa,
				Namespace: namespace,
			},
		},
	}
	return roleb
}

func AuthProxyRoleBindingApplyConfiguration(sa, namespace string) *rbacv1apply.RoleBindingApplyConfiguration {
	return rbacv1apply.RoleBinding(AuthProxyRoleName, namespace).
		WithRoleRef(rbacv1apply.RoleRef().
			WithAPIGroup(rbacv1.GroupName).
			WithKind("Role").
			WithName(AuthProxyRoleName)).
		WithSubjects(rbacv1apply.Subject().
			WithKind("ServiceAccount").
			WithName(sa).
			WithNamespace(namespace))
}
