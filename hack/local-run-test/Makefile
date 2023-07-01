SHELL := /bin/bash
BIN := $(shell pwd)/bin
PROJ_BIN := $(abspath $(shell pwd)/../../bin)
export BIN_DIR := $(BIN)
export PATH := $(BIN):$(PROJ_BIN):$(PATH)
export KUBECONFIG := $(BIN)/kubeconfig
unexport KUBERNETES_SERVICE_PORT
unexport KUBERNETES_SERVICE_HOST
unexport KUBERNETES_PORT

##---------------------------------------------------------------------
##@ General
##---------------------------------------------------------------------
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##---------------------------------------------------------------------
##@ show variables
##---------------------------------------------------------------------
# To access services on k3d in code-server from outside the cluster,
# DNS `*.domain.name` must be configured to be name-resolved to Ingress for code-server
# and to be routed to code-sever Pod's 15000 port

# In code-server, the URL can be got in the env $VSCODE_PROXY_URI (e.g. `https://example.domain.name/proxy/{{port}}`)
# COSMO Dashboard on k3d is configured to be available at `https://dash-k3d-example.domain.name`
VSCODE_PROXY_URI ?= $$VSCODE_PROXY_URI
PARENT_URLBASE ?= $(shell echo $(VSCODE_PROXY_URI) | grep -Po "(?<=://)[^\.]*")
DOMAIN ?= $(shell echo $(VSCODE_PROXY_URI) | grep -Po "(?<=\.)[^/]*")
DEFAULT_URLBASE_HOST := {{NETRULE}}-{{WORKSPACE}}-{{USER}}-k3d-$(PARENT_URLBASE)
K3D_URLBASE := k3d-$(PARENT_URLBASE).$(DOMAIN)
DASHBOARD_HOST := dash-$(K3D_URLBASE)
DASHBOARD_URL := https://$(DASHBOARD_HOST)
SIGNIN_URL := https://$(DASHBOARD_HOST)/\#/signin
TRAEFIK_DASHBOARD_HOST := traefik-dashboard-$(K3D_URLBASE)
TRAEFIK_DASHBOARD_URL := https://$(TRAEFIK_DASHBOARD_HOST)/
ARGOCD_HOST := argocd-$(K3D_URLBASE)
ARGOCD_URL := https://$(ARGOCD_HOST)
LDAP_SSL_HOST := ssl-ldap2-$(K3D_URLBASE)
LDAP_SSL_URL := https://ssl-ldap2-$(K3D_URLBASE)
PHPLDAPADMIN_HOST := phpldapadmin-$(K3D_URLBASE)
PHPLDAPADMIN_URL := https://phpldapadmin-$(K3D_URLBASE)
AUTHENTIK_HOST := authentik-$(K3D_URLBASE)
AUTHENTIK_URL := https://$(AUTHENTIK_HOST)

show-url: ## show-url
	@echo ----------------------------------------------------
	@echo "DOMAIN                : $(DOMAIN)"
	@echo "PARENT_URLBASE        : $(PARENT_URLBASE)"
	@echo "DEFAULT_URLBASE_HOST  : $(DEFAULT_URLBASE_HOST)"
	@echo "DASHBOARD_URL         : $(DASHBOARD_URL)"
	@echo "SIGNIN_URL            : $(SIGNIN_URL)"
	@echo "TRAEFIK_DASHBOARD_URL : $(TRAEFIK_DASHBOARD_URL)"
	@echo "ARGOCD_URL            : $(ARGOCD_URL)"
	@echo "LDAP_SSL_URL          : $(LDAP_SSL_URL)"
	@echo "PHPLDAPADMIN_URL      : $(PHPLDAPADMIN_URL)"
	@echo "AUTHENTIK_URL         : $(AUTHENTIK_URL)"
	@echo ----------------------------------------------------

##---------------------------------------------------------------------
##@ Batch processing
##---------------------------------------------------------------------
.PHONY: create-all delete-all
.PHONY: docker-build-all
.PHONY: install-all uninstall-all

create-all: create-cluster docker-build-all install-all apply-template add-user add-workspace ## Create all
delete-all: delete-cluster ## Delete all

docker-build-all: docker-build-manager docker-build-dashboard ## Docker build all

install-all: install-manager install-dashboard ## Install cosmo resources.
uninstall-all: uninstall-manager uninstall-dashboard ## Uninstall cosmo resources.

##---------------------------------------------------------------------
##@ Kubernetes environment
##---------------------------------------------------------------------
CLUSTER_NAME := cosmo
.PHONY: create-cluster delete-cluster create-k3d-cluster deploy-cert-manager deploy-traefik deploy-argocd undeploy-argocd install-authentik

# ports
#   389  : ldap port.
#   636  : ldaps port.
#   15000: traefik web entrypoint 80.
#   15443: traefik websecure entrypoint 443.
#   30080: argocd nodeport
create-k3d-cluster: k3d kubectl  ## Create k3s cluster.
	@echo ====== $@ ======
	@k3d cluster list $(CLUSTER_NAME) > /dev/null 2>&1 && echo "This cluster has already been created" || { \
		k3d cluster create $(CLUSTER_NAME) \
			-i rancher/k3s:v1.26.3-k3s1 \
			--api-port 43000 \
			-p "389:30389@loadbalancer" \
			-p "636:30636@loadbalancer" \
			-p "15000:80@loadbalancer" \
			-p "15443:443@loadbalancer" \
			-p "30080:30080@loadbalancer" \
			--registry-create cosmo.io:0.0.0.0:5000 \
			--agents 2 \
			--k3s-arg "-v=2@server:*" \
			--k3s-arg "--disable=local-storage@server:*" \
			--k3s-arg "--disable=traefik@server:*" ; \
	}
deploy-cert-manager: create-k3d-cluster helm ## Deploy cert-manager.
	@echo ====== $@ ======
	helm repo add jetstack https://charts.jetstack.io
	helm repo update
	helm upgrade --install cert-manager jetstack/cert-manager --namespace cert-manager \
		--create-namespace --version v1.10.0 --set installCRDs=true --wait

deploy-traefik: create-k3d-cluster helm docker-build-traefik-plugins ## Deploy traefik.
	@echo ====== $@ ======
	helm dependency update ../../charts/cosmo-traefik 
	helm template traefik ../../charts/cosmo-traefik -n cosmo-system \
		-f <(sed -e 's/\S*\/cosmo-traefik-plugins:\S*/cosmo.io:5000\/traefik-plugins:latest/g' ../../charts/cosmo-traefik/values.yaml ) \
	 	-f ./charts/traefik.values.yaml \
		--set traefik.ingressRoute.dashboard.matchRule="Host(\`$(TRAEFIK_DASHBOARD_HOST)\`)" \
		--set cosmoAuth.signinUrl=$(SIGNIN_URL) \
		--set domain=$(DOMAIN) \
		> ./bin/traefik.manifest.yaml
	helm upgrade --install traefik ../../charts/cosmo-traefik -n cosmo-system --create-namespace \
		--wait \
		-f <(sed -e 's/\S*\/cosmo-traefik-plugins:\S*/cosmo.io:5000\/traefik-plugins:latest/g' ../../charts/cosmo-traefik/values.yaml ) \
	 	-f ./charts/traefik.values.yaml \
		--set traefik.ingressRoute.dashboard.matchRule="Host(\`$(TRAEFIK_DASHBOARD_HOST)\`)" \
		--set cosmoAuth.signinUrl=$(SIGNIN_URL) \
		--set domain=$(DOMAIN)
	-kubectl rollout restart deploy -n cosmo-system traefik
	-kubectl rollout restart deploy -n cosmo-system cosmo-dashboard 


ARGOCD_LOCAL_URL := localhost:30080
# argocd value.yaml https://github.com/argoproj/argo-helm/blob/main/charts/argo-cd/values.yaml
deploy-argocd: create-k3d-cluster helm argocd ## Deploy argocd.
	@echo ====== $@ ======
	helm repo add argo https://argoproj.github.io/argo-helm
	helm repo update
	helm template -n argocd --create-namespace argocd argo/argo-cd \
		--set configs.params."server\.insecure"=true \
		--set server.service.type=NodePort \
		--set server.service.nodePortHttp=30080 \
		> ./bin/argocd.manifest.yaml
	helm upgrade --install -n argocd --create-namespace argocd argo/argo-cd \
		--set configs.params."server\.insecure"=true \
		--set server.service.type=NodePort \
		--set server.service.nodePortHttp=30080 \
		--wait
	@echo ====== wait argocd ======
	until (kubectl get po -n argocd -l app.kubernetes.io/name=argocd-server  | grep argocd-server) do sleep 1; done
	kubectl wait po -n argocd -l app.kubernetes.io/name=argocd-server --for condition=Ready
	kubectl get po -n argocd
	@echo ====== argocd-setting ======
	bash -c "[[ -f ~/.config/argocd/config ]] && chmod 600 ~/.config/argocd/config || echo init"
	@argocd login $(ARGOCD_LOCAL_URL) --grpc-web --plaintext --username admin --password vvvvvvvv > /dev/null 2>&1 && \
		echo "Initial password has already been changed" || { \
		echo ====== login argocd ====== && \
		INITIAL_PW=$$(kubectl -n argocd get secret/argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d) && \
		argocd login $(ARGOCD_LOCAL_URL) --grpc-web --plaintext --username admin --password $${INITIAL_PW} && \
		echo ====== argocd password change ====== && \
		argocd account update-password --account admin --current-password $${INITIAL_PW} --new-password vvvvvvvv && \
		echo "Initial password has been changed" \
		; \
	}
	@echo ----------------------------------------------------
	@echo "url  : $(ARGOCD_URL)"
	@echo "user : admin"
	@echo "pass : vvvvvvvv"
	@echo ----------------------------------------------------

undeploy-argocd: ## Undeploy argocd.
	helm uninstall -n argocd argocd


install-openldap: deploy-argocd ## Install openldap.
	@echo ====== $@ ======
	kubectl apply -f ./manifests/openldap.cert.yaml
	kubectl apply -f <(sed \
		-e 's/$${LDAP_SSL_HOST}/$(LDAP_SSL_HOST)/g' \
		-e 's/$${PHPLDAPADMIN_HOST}/$(PHPLDAPADMIN_HOST)/g' \
		./manifests/openldap.app.yaml )
	@echo ====== wait openldap ======
	until (kubectl get po -n cosmo-system -l app.kubernetes.io/name=openldap-stack-ha | grep open-ldap) do sleep 1; done
	kubectl wait po -n cosmo-system -l app.kubernetes.io/name=openldap-stack-ha --for condition=Ready --timeout=90s
	kubectl get po -n cosmo-system
	@echo ----------------------------------------------------
	@echo "url  : $(PHPLDAPADMIN_URL)"
	@echo "user : admin"
	@echo "pass : vvvvvvvv"
	@echo ----------------------------------------------------


install-authentik: deploy-argocd ## Install authentik. (optional)
	kubectl apply -f <(sed -e 's/$${AUTHENTIK_HOST}/$(AUTHENTIK_HOST)/g' ./manifests/authentik.app.yaml )
	@echo ----------------------------------------------------
	@echo "initial setup url : $(AUTHENTIK_URL)/if/flow/initial-setup/"
	@echo "user : admin"
	@echo "pass : vvvvvvvv"
	@echo ----------------------------------------------------

create-cluster: create-k3d-cluster deploy-cert-manager deploy-traefik deploy-argocd install-openldap  ## Create k3s cluster.

delete-cluster: k3d  ## Delete k3s cluster.
	@echo ====== $@ ======
	-@k3d cluster delete $(CLUSTER_NAME)


##---------------------------------------------------------------------
##@ Docker image build
##---------------------------------------------------------------------
IMAGE_TAG ?= latest
MANAGER_IMAGE_TAG   ?= $(IMAGE_TAG)
DASHBOARD_IMAGE_TAG ?= $(IMAGE_TAG)
TRAEFIK_PLUGINS_IMAGE_TAG ?= $(IMAGE_TAG)

.PHONY: docker-build-manager docker-build-dashboard docker-build-traefik-plugins docker-cache-clear

docker-build-manager: ## build & push cosmo contoller-manager image.
	@echo ====== $@ ======
	cd ../.. && make docker-build-manager VERSION=$(MANAGER_IMAGE_TAG) QUICK_BUILD=yes
	docker tag cosmo-controller-manager:$(MANAGER_IMAGE_TAG) localhost:5000/cosmo-controller-manager:$(MANAGER_IMAGE_TAG)
	docker push localhost:5000/cosmo-controller-manager:$(MANAGER_IMAGE_TAG)
	docker rmi cosmo-controller-manager:$(MANAGER_IMAGE_TAG)
	k3d image import localhost:5000/cosmo-controller-manager:$(MANAGER_IMAGE_TAG) -c $(CLUSTER_NAME)
	-kubectl rollout restart deploy -n cosmo-system cosmo-controller-manager

docker-build-dashboard: ## build & push cosmo dashboard image.
	@echo ====== $@ ======
	cd ../.. && make docker-build-dashboard VERSION=$(DASHBOARD_IMAGE_TAG) QUICK_BUILD=yes
	docker tag cosmo-dashboard:$(DASHBOARD_IMAGE_TAG) localhost:5000/cosmo-dashboard:$(DASHBOARD_IMAGE_TAG)
	docker push localhost:5000/cosmo-dashboard:$(DASHBOARD_IMAGE_TAG)
	docker rmi cosmo-dashboard:$(DASHBOARD_IMAGE_TAG)
	k3d image import localhost:5000/cosmo-dashboard:$(DASHBOARD_IMAGE_TAG) -c $(CLUSTER_NAME)
	-kubectl rollout restart deploy -n cosmo-system cosmo-dashboard 

docker-build-traefik-plugins: ## build & push cosmo traefik-plugins image.
	@echo ====== $@ ======
	cd ../.. && make docker-build-traefik-plugins VERSION=$(TRAEFIK_PLUGINS_IMAGE_TAG) QUICK_BUILD=yes
	docker tag cosmo-traefik-plugins:$(TRAEFIK_PLUGINS_IMAGE_TAG) localhost:5000/traefik-plugins:$(TRAEFIK_PLUGINS_IMAGE_TAG)
	docker push localhost:5000/traefik-plugins:$(TRAEFIK_PLUGINS_IMAGE_TAG)
	docker rmi cosmo-traefik-plugins:$(TRAEFIK_PLUGINS_IMAGE_TAG)
	k3d image import localhost:5000/traefik-plugins:$(TRAEFIK_PLUGINS_IMAGE_TAG) -c $(CLUSTER_NAME)
	-kubectl rollout restart deploy -n cosmo-system traefik


docker-cache-clear: ## docker cache clear.
	-docker system df
	-docker images
	-docker rmi `docker images -f "dangling=true" -q`
	-docker builder prune -f
	-docker system df
	-docker images

##---------------------------------------------------------------------
##@ Install cosmo to k8s
##---------------------------------------------------------------------
.PHONY: install-manager   uninstall-manager
.PHONY: install-dashboard uninstall-dashboard

LOGLEVEL ?= info

install-manager: helm kubectl docker-build-manager ## Install cosmo contoller-manager resources.
	@echo ====== $@ ======
	helm upgrade --install cosmo-controller-manager ../../charts/cosmo-controller-manager \
		-n cosmo-system --create-namespace \
		--wait \
		--set image.repository=cosmo.io:5000/cosmo-controller-manager \
		--set image.tag=$(MANAGER_IMAGE_TAG) \
		--set image.pullPolicy=Always \
		--set defaultURLBase.host="\$(DEFAULT_URLBASE_HOST)" \
		--set defaultURLBase.domain="\$(DOMAIN)" \
		--set logLevel=$(LOGLEVEL)

install-manager-local: helm kubectl ## Install cosmo contoller-manager resources.
	@echo ====== $@ ======
	helm upgrade --install cosmo-controller-manager ../../charts/cosmo-controller-manager \
		-n cosmo-system --create-namespace \
		--wait \
		--set image.repository=cosmo.io:5000/cosmo-controller-manager \
		--set image.tag=$(MANAGER_IMAGE_TAG) \
		--set image.pullPolicy=Always \
		--set defaultURLBase.host="\$(DEFAULT_URLBASE_HOST)" \
		--set defaultURLBase.domain="\$(DOMAIN)" \
		--set replicaCount=0 \
		--set localRunTest.enabled=true \
		--set localRunTest.webhookEndpoint.ip=$(shell hostname -i) \
		--set localRunTest.webhookEndpoint.port=9443 \
		--set logLevel="$(LOGLEVEL)"

install-dashboard: helm kubectl docker-build-dashboard ## Install cosmo dashboard resources.
	@echo ====== $@ ======
	helm upgrade --install cosmo-dashboard ../../charts/cosmo-dashboard \
		-n cosmo-system \
		--wait \
		--set image.repository=cosmo.io:5000/cosmo-dashboard \
		--set image.tag=$(DASHBOARD_IMAGE_TAG) \
		--set image.pullPolicy=Always \
		--set ingress.enabled=true \
		--set ingress.hostName=$(DASHBOARD_HOST) \
		--set cert.dnsName=$(DASHBOARD_HOST) \
		--set insecure=false \
		--set auth.ldap.enabled=true \
		--set auth.ldap.url="ldaps://open-ldap.cosmo-system.svc.cluster.local" \
		--set XXXXXauth.ldap.bindDn="cn=%s\,ou=users\,dc=cosmows\,dc=dev" \
		--set auth.ldap.tls.secretName=openldap-cert \
    	--set auth.ldap.searchBindDN="cn=admin\,dc=cosmows\,dc=dev" \
    	--set auth.ldap.searchBindPassword="vvvvvvvv" \
    	--set auth.ldap.searchBaseDN="dc=cosmows\,dc=dev" \
    	--set auth.ldap.searchFilter="(uid=\%s)" \
		--set logLevel=$(LOGLEVEL)

install-dashboard-local: helm kubectl ## Install cosmo dashboard resources.
	@echo ====== $@ ======
	helm upgrade --install cosmo-dashboard ../../charts/cosmo-dashboard \
		-n cosmo-system \
		--wait \
		--set image.repository=cosmo.io:5000/cosmo-dashboard \
		--set image.tag=$(DASHBOARD_IMAGE_TAG) \
		--set image.pullPolicy=Always \
		--set ingress.enabled=true \
		--set ingress.hostName=$(DASHBOARD_HOST) \
		--set cert.dnsName=$(DASHBOARD_HOST) \
		--set auth.ldap.enabled=false \
		--set insecure=true \
		--set replicaCount=0 \
		--set localRunTest.enabled=true \
		--set localRunTest.dashboardEndpoint.ip=$(shell hostname -i) \
		--set localRunTest.dashboardEndpoint.uiServerPort=3000 \
		--set localRunTest.dashboardEndpoint.serverPort=8443 \
		--set logLevel=$(LOGLEVEL)

uninstall-manager: helm  ## Uninstall cosmo contoller-manager resources.
	@echo ====== $@ ======
	-helm uninstall -n cosmo-system cosmo-controller-manager

uninstall-dashboard: helm  ## Uninstall cosmo dashboard resources.
	@echo ====== $@ ======
	-helm uninstall -n cosmo-system cosmo-dashboard


apply-template: kubectl cosmoctl ## Apply template.
	@echo ====== $@ ======
	until (kubectl apply -f ../../config/user-addon/traefik-middleware/useraddon-cosmo-username-headers.yaml) do sleep 1; done
	for i in `ls ../../example/useraddons/*/*.yaml`; do until (kubectl apply -f $$i) do sleep 1; done; done
	for i in `ls ../../example/workspaces/*.yaml`; do until (kubectl apply -f $$i) do sleep 1; done; done
	$(MAKE) -C templates/dev-code-server apply


add-user: kubectl cosmoctl ## add user
	@echo ====== $@ ======
	-cosmoctl user create tom --admin 2> /dev/null
	-cosmoctl user create team-a-dev --role "team-a-dev" 2> /dev/null
	-cosmoctl user create team-a-admin --role "team-a-admin" 2> /dev/null
	-cosmoctl user create team-b-dev --role "team-b-dev" 2> /dev/null
	-cosmoctl user create team-b-admin --role "team-b-admin" 2> /dev/null
	-cosmoctl user create team-ab-dev --role "team-a-dev,team-b-dev" 2> /dev/null
	-cosmoctl user create ldapuser1 --admin --auth-type ldap 2> /dev/null
	-cosmoctl user reset-password tom --password vvv
	-cosmoctl user reset-password team-a-dev   --password xxxxxxxx
	-cosmoctl user reset-password team-a-admin --password xxxxxxxx
	-cosmoctl user reset-password team-b-dev   --password xxxxxxxx
	-cosmoctl user reset-password team-b-admin --password xxxxxxxx
	-cosmoctl user reset-password team-ab-dev  --password xxxxxxxx


add-workspace: kubectl cosmoctl ## add workspace
	@echo ====== $@ ======
	-cosmoctl workspace create --user=tom --template=dev-code-server ws1
	-cosmoctl workspace create --user=ldapuser1 --template=dev-code-server ldapws1
	sleep 5
	-cosmoctl networkrule add --user=tom --workspace=ws1 --port=7701 --host-prefix proxy1 --path /
	sleep 1
	-cosmoctl networkrule add --user=tom --workspace=ws1 --port=7701 --host-prefix proxy1 --path /aaa
	sleep 1
	-cosmoctl networkrule add --user=tom --workspace=ws1 --port=7701 --host-prefix proxy1 --path /bbb --public
	sleep 1
	-cosmoctl networkrule add --user=tom --workspace=ws1 --port=7701 --path /

##---------------------------------------------------------------------
##@ Execute test
##---------------------------------------------------------------------
.PHONY: run-manager-local run-dashboard-local run-local stop

run-manager-local: go
	@echo ====== $@ ======
	@#(cd ../.. && bash hack/download-certs.sh webhook-server-cert && make run QUICK_BUILD=yes)
	cd ../.. && bash hack/download-certs.sh webhook-server-cert && \
		$(GO) run ./cmd/controller-manager/main.go \
		--metrics-bind-address :8085 \
		--cert-dir . \
		--zap-log-level=$(LOGLEVEL) \
		--zap-devel=false \
		--workspace-urlbase-protocol=https \
		--workspace-urlbase-host=$(DEFAULT_URLBASE_HOST) \
		--workspace-urlbase-domain=$(DOMAIN)

run-dashboard-local:
	@echo ====== $@ ======
	cd ../.. && \
	$(GO) run ./cmd/dashboard/main.go \
		--zap-log-level=$(LOGLEVEL) \
		--zap-time-encoding=iso8601 \
		--cookie-domain=$(DOMAIN) \
		--cookie-session-name="$(shell kubectl get secret -n cosmo-system cosmo-auth-env -o=jsonpath={.data.COOKIE_SESSION_NAME} | base64 -d)" \
		--cookie-hashkey="$(shell kubectl get secret -n cosmo-system cosmo-auth-env -o=jsonpath={.data.COOKIE_HASHKEY} | base64 -d)" \
		--cookie-blockkey="$(shell kubectl get secret -n cosmo-system cosmo-auth-env -o=jsonpath={.data.COOKIE_BLOCKKEY} | base64 -d)" \
		--ldap-url=ldap://localhost:389 \
		--ldap-binddn="cn=%s,ou=users,dc=cosmows,dc=dev" \
		--insecure

run-local: kubectl ## Run test.
	-@make uninstall-manager
	make install-manager-local
	-@make uninstall-dashboard
	make install-dashboard-local
	@# if session presents then attach, not presents then creat session
	@tmux -2 attach-session -d -t cosmo 2> /dev/null || { \
		# ======== create session ======== \
		tmux -2 new-session -d -s cosmo -n w-run && \
		tmux set -g mouse on && \
		tmux bind-key m if-shell "tmux show-options -g mouse | grep off" "set-option -g mouse on" "set-option -g mouse off" && \
		tmux set-window-option -g mode-keys vi && \
		# ======== create pane layout ======== \
		# +----------------------------------------------+ \
		# | 0 control-manager                            | \
		# +------------------+---------------------------+ \
		# | 1 dashboard-ui   | 2 dashboard server        | \
		# +------------------+---------------------------+ \
		# | 3                                            | \
		# +----------------------------------------------+ \
		tmux splitw -v \; splitw -v \; select-layout even-vertical \
			\; splitw -h -t 1 \
			\; resizep -t 1 -x 20% \; resizep -t 3 -x 20% && \
		# ======== run controller-manager ======== \
		tmux send-keys -t 0 'make run-manager-local 2>&1 | tee -a $${BIN_DIR}/local-manager.log' Enter && \
		# ======== run dashboard-ui ======== \
		tmux send-keys -t 1 '(cd ../.. && make run-dashboard-ui 2>&1 | tee -a $${BIN_DIR}/local-dashboard-ui.log)' Enter && \
		# ======== run dashboard ======== \
		tmux send-keys -t 2 'make run-dashboard-local 2>&1 | tee -a $${BIN_DIR}/local-dashboard.log' Enter && \
		# ======== move to pane 3 ======== \
		tmux select-pane -t 3 && \
		# ======== attach ======== \
		tmux attach-session -d -t cosmo && \
		echo end ; \
	}

stop: ## Stop test.
	-@tmux kill-session -t cosmo 2> /dev/null

logs-traefik-plugins:
	-kubectl logs deploy/traefik -n cosmo-system -c traefik -f | while read line; do echo $${line} | grep cosmo-auth; done

##---------------------------------------------------------------------
##@ Build Binary
##---------------------------------------------------------------------
GO_SRC := $(shell find ../../internal ../../pkg -type f -regex ".*\.go")
cosmoctl: bin bin/cosmoctl ## cosmoctl
bin/cosmoctl: ../../cmd/cosmoctl/main.go $(GO_SRC)
	@echo ====== $@ ======
	make -C ../../ cosmoctl
	mv ../../bin/cosmoctl bin/cosmoctl

##---------------------------------------------------------------------
##@ Download Binary
##---------------------------------------------------------------------
bin:
	mkdir -p $(BIN)

GO_VERSION ?= 1.20.4
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif
GO ?= $(GOBIN)/go$(GO_VERSION)


go: $(GO) ## go
$(GO):
	go install golang.org/dl/go$(GO_VERSION)@latest
	$(GO) download


k3d: bin bin/k3d ## k3d
bin/k3d:
	@echo ====== $@ ======
	export K3D_INSTALL_DIR=$(BIN) && \
	curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash


helm: bin bin/helm ## helm
bin/helm:
	@echo ====== $@ ======
	export HELM_INSTALL_DIR=$(BIN) && \
	curl -s curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash


kubectl: bin bin/kubectl ## kubectl
bin/kubectl:
	@echo ====== $@ ======
	$(eval KUBE_VER := $(shell curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt))
	curl -L -o $@ "https://storage.googleapis.com/kubernetes-release/release/$(KUBE_VER)/bin/linux/amd64/kubectl"
	chmod +x $@

argocd: bin bin/argocd ## argocd
bin/argocd:
	curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
	sudo install -m 555 argocd-linux-amd64 bin/argocd
	rm -f argocd-linux-amd64

##---------------------------------------------------------------------
##@ Utility
##---------------------------------------------------------------------
console: ## Activate kubeconfig for local k8s. 
	@bash -rcfile <(echo ". ~/.bashrc;PS1='\[\033[01;32m\]\u@test-env\[\033[00m\]:\[\033[01;35m\]\W\[\033[00m\]$$ '")

helm-ls: ## helm list
	-@helm list -a -A

kg: ## Get k0s resources.
	-@kubectl get node --show-kind
	-@kubectl get po -A --show-kind
	-@kubectl get ing -A --show-kind
	-@kubectl get ingressroute.traefik.io -A --show-kind
	-@kubectl get middleware.traefik.io -A --show-kind
	-@kubectl get svc -A --show-kind
	-@kubectl get ep -A --show-kind
	-@kubectl get application -A --show-kind
