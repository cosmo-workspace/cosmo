package v1alpha1

// Template annotation keys for WorkspaceConfig
const (
	TemplateAnnKeyURLBase                  = "cosmo/ws-urlbase"
	TemplateAnnKeyWorkspaceDeployment      = "cosmo/ws-deployment"
	TemplateAnnKeyWorkspaceService         = "cosmo/ws-service"
	TemplateAnnKeyWorkspaceIngress         = "cosmo/ws-ingress"
	TemplateAnnKeyWorkspaceServiceMainPort = "cosmo/ws-service-main-port"
)

// Template annotation keys for user addon
const (
	TemplateAnnKeyDefaultUserAddon = "cosmo/default-user-addon"
)

// Template variables key
const (
	TemplateVarDeploymentName      = "{{WORKSPACE_DEPLOYMENT_NAME}}"
	TemplateVarServiceName         = "{{WORKSPACE_SERVICE_NAME}}"
	TemplateVarIngressName         = "{{WORKSPACE_INGRESS_NAME}}"
	TemplateVarServiceMainPortName = "{{WORKSPACE_SERVICE_MAIN_PORT_NAME}}"

	// Var for user addon
	TemplateVarUserName = "{{USER_NAME}}"
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
