all: user-addon

user-addon:
	kustomize build config/auth-proxy/ | cosmoctl  template generate --name cosmo-auth-proxy-role \
		--user-addon --required-vars SERVICE_ACCOUNT:default > user-addon-cosmo-auth-proxy-role.yaml