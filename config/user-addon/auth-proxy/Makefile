all: user-addon

user-addon:
	kustomize build . | cosmoctl  template generate --name cosmo-auth-proxy-role \
		--user-addon --required-vars SERVICE_ACCOUNT:default \
		--desc 'Role and Rolebinding for COSMO Auth Proxy. By default, it is bound to the service account named `default` in the user namespace.' \
		--set-default-user-addon > user-addon-cosmo-auth-proxy-role.yaml