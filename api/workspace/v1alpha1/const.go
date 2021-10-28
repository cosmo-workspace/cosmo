package v1alpha1

import "fmt"

// Instance annotation keys for WorkspaceConfig
const (
	InstanceAnnKeyURLBase                  = "cosmo/ws-urlbase"
	InstanceAnnKeyWorkspaceDeployment      = "cosmo/ws-deployment"
	InstanceAnnKeyWorkspaceService         = "cosmo/ws-service"
	InstanceAnnKeyWorkspaceIngress         = "cosmo/ws-ingress"
	InstanceAnnKeyWorkspaceServiceMainPort = "cosmo/ws-service-main-port"
)

// Key of NetworkRule Group
const (
	InstanceIngressAnnKeyNetRuleGroupPrefix = "cosmo/ws-netrule-group"
)

func InstanceIngressAnnKeyNetRuleGroup(portName string) string {
	return fmt.Sprintf("%s.%s", InstanceIngressAnnKeyNetRuleGroupPrefix, portName)
}

// Template variables key
const (
	TemplateVarDeploymentName      = "{{WORKSPACE_DEPLOYMENT_NAME}}"
	TemplateVarServiceName         = "{{WORKSPACE_SERVICE_NAME}}"
	TemplateVarIngressName         = "{{WORKSPACE_INGRESS_NAME}}"
	TemplateVarServiceMainPortName = "{{WORKSPACE_SERVICE_MAIN_PORT_NAME}}"
)

// UserPasswordSecret name and keys
const (
	UserPasswordSecretName                        = "password"
	UserPasswordSecretDataKeyUserPasswordSecret   = "password"
	UserPasswordSecretDataKeyUserPasswordSalt     = "salt"
	UserPasswordSecretAnnKeyUserPasswordIfDefault = "cosmo/default"
)

// TemplateType enum for Workspace
const (
	TemplateTypeWorkspace = "workspace"
	TemplateTypeUserAddon = "user-addon"
)

// AuthProxy RBAC names
const (
	AuthProxyRoleName               = "cosmo-auth-proxy-role"
	AuthProxyClusterRoleBindingName = "cosmo-auth-proxy-rolebinding"
)
