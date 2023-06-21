FROM_TAG=4.9.1

#URLBASE ?= 'https://{{NETRULE_GROUP}}-{{INSTANCE}}-{{NAMESPACE}}.domain'

.PHONY: template
template:
	cd kubernetes/ && kustomize edit set image codercom/code-server=cosmo.io:5000/my-code-server:latest
	kustomize build kubernetes/ | cosmoctl tmpl generate -o cosmo-template.yaml --workspace \
		--required-vars CODE-SERVER_STORAGE_GB:20,DOCKER_STORAGE:20

.PHONY: apply
apply: template ## Apply template
	kubectl apply -f cosmo-template.yaml

.PHONY: docker-build
docker-build: ## Build docker image
	DOCKER_BUILDKIT=1 docker build . -t my-code-server:latest --build-arg FROM_TAG=${FROM_TAG}
	docker tag my-code-server:latest localhost:5000/my-code-server:latest
	docker push localhost:5000/my-code-server:latest
