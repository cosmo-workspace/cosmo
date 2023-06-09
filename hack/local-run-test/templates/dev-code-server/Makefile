IMAGE_REPO=ghcr.io/cosmo-workspace
IMAGE_NAME=dev-code-server
IMAGE_TAG=v0.0.2-4.13.0
#URLBASE ?= 'https://{{NETRULE_GROUP}}-{{INSTANCE}}-{{NAMESPACE}}.domain'

.PHONY: template
template: ## Create template
	cd kubernetes/ && kustomize edit set image $(IMAGE_REPO)/$(IMAGE_NAME):$(IMAGE_TAG)
	kustomize build kubernetes/ | cosmoctl tmpl generate -o cosmo-template.yaml --workspace \
		--workspace-urlbase "$(URLBASE)" \
		--disable-inject-auth-proxy \
		--required-vars CODE-SERVER_STORAGE_GB:20,DOCKER_STORAGE:20

.PHONY: apply
apply: template ## Apply template
	kubectl apply -f cosmo-template.yaml
	kubectl apply -f team-a-template.yaml
	kubectl apply -f team-shared-template.yaml
