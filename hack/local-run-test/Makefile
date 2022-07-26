BIN := $(shell pwd)/bin
export BIN_DIR := $(BIN)
export PATH := $(BIN):$(PATH)
export KUBECONFIG := $(BIN)/kubeconfig
unexport KUBERNETES_SERVICE_PORT
unexport KUBERNETES_SERVICE_HOST
unexport KUBERNETES_PORT

# $VSCODE_PROXY_URI => https://main-wsname-cosmo-user-username.domain.name/proxy/{{port}}
#                                   |-----PARENT_URLBASE-----| |--DOMAIN-|
PARENT_URLBASE ?= $(shell echo $$VSCODE_PROXY_URI | grep -Po "(?<=://main-)[^\.]*")
DOMAIN ?= $(shell echo $$VSCODE_PROXY_URI | grep -Po "(?<=\.)[^/]*")
DEFAULT_URLBASE_HOST := {{NETRULE_GROUP}}-{{INSTANCE}}-{{USERID}}-k3d-$(PARENT_URLBASE)
DASHBOARD_URL := dash-k3d-$(PARENT_URLBASE).$(DOMAIN)

##---------------------------------------------------------------------
##@ General
##---------------------------------------------------------------------
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


##---------------------------------------------------------------------
##@ Batch processing
##---------------------------------------------------------------------
.PHONY: create-all delete-all
.PHONY: docker-build-all
.PHONY: install-all uninstall-all

create-all: create-cluster docker-build-all install-all apply-template add-user add-workspace ## Create all
delete-all: delete-cluster ## Delete all

docker-build-all: docker-build-manager docker-build-dashboard docker-build-auth-proxy docker-build-cs ## Docker build all

install-all: install-manager install-dashboard ## Install cosmo resources.
uninstall-all: uninstall-manager uninstall-dashboard ## Uninstall cosmo resources.

##---------------------------------------------------------------------
##@ Kubenetes environment
##---------------------------------------------------------------------
CLUSTER_NAME := cosmo
.PHONY: create-cluster delete-cluster

create-cluster: k3d helm kubectl  ## Create k3s cluster with cert-manager.
	@echo ====== $@ ======
	@k3d cluster list $(CLUSTER_NAME) > /dev/null 2>&1 && echo "This cluster has already been created" || { \
		echo ====== create cluser ======= && \
		k3d cluster create $(CLUSTER_NAME) \
			--api-port 43000 \
			-p "15000:80@loadbalancer" \
			--registry-create cosmo.io:0.0.0.0:5000 \
			--agents 2 \
			--k3s-arg "--no-deploy=local-storage@server:*" \
			--k3s-arg "--debug@server:*" \
		&& \
		echo ====== deploy cert-manager ====== && \
		helm repo add jetstack https://charts.jetstack.io && \
		helm repo update && \
		helm upgrade --install cert-manager jetstack/cert-manager --namespace cert-manager \
			--create-namespace --version v1.7.1 --set installCRDs=true --wait && \
		echo ====== wait traefik ====== && \
		until (kubectl get po -n kube-system -l app.kubernetes.io/instance=traefik  | grep traefik) do sleep 1; done ; \
		kubectl wait po -n kube-system -l app.kubernetes.io/instance=traefik --for condition=Ready ; \
	}

delete-cluster: k3d  ## Delete k3s cluster.
	@echo ====== $@ ======
	-@k3d cluster delete $(CLUSTER_NAME)


##---------------------------------------------------------------------
##@ Docker image build
##---------------------------------------------------------------------
.PHONY: docker-build-manager docker-build-dashboard docker-build-auth-proxy docker-cache-clear

docker-build-manager: ## build & push cosmo contoller-manager image.
	@echo ====== $@ ======
	cd ../.. && make docker-build-manager VERSION=latest QUICK_BUILD=yes
	docker tag cosmo-controller-manager:latest localhost:5000/cosmo-controller-manager:latest
	docker push localhost:5000/cosmo-controller-manager:latest
	k3d image import localhost:5000/cosmo-controller-manager:latest -c $(CLUSTER_NAME)

docker-build-dashboard: ## build & push cosmo dashboard image.
	@echo ====== $@ ======
	cd ../.. && make docker-build-dashboard VERSION=latest QUICK_BUILD=yes
	docker tag cosmo-dashboard:latest localhost:5000/cosmo-dashboard:latest
	docker push localhost:5000/cosmo-dashboard:latest
	k3d image import localhost:5000/cosmo-dashboard:latest -c $(CLUSTER_NAME)

docker-build-auth-proxy: ## build & push cosmo auth-proxy image.
	@echo ====== $@ ======
	cd ../.. && make docker-build-auth-proxy VERSION=latest QUICK_BUILD=yes
	docker tag cosmo-auth-proxy:latest localhost:5000/cosmo-auth-proxy:latest
	docker push localhost:5000/cosmo-auth-proxy:latest
	k3d image import localhost:5000/cosmo-auth-proxy:latest -c $(CLUSTER_NAME)

docker-cache-clear: ## docker cache clear.
	docker system df
	docker builder prune

##---------------------------------------------------------------------
##@ Install to k8s
##---------------------------------------------------------------------
.PHONY: install-manager install-dashboard
.PHONY: uninstall-manager uninstall-dashboard

install-manager: helm kubectl create-cluster docker-build-manager ## Install cosmo contoller-manager resources.
	@echo ====== $@ ======
	helm upgrade --install cosmo-controller-manager ../../charts/cosmo-controller-manager \
		-n cosmo-system --create-namespace \
		--wait \
		--set image.repository=cosmo.io:5000/cosmo-controller-manager \
		--set image.tag=latest \
		--set image.pullPolicy=Always \
		--set defaultURLBase.host="\$(DEFAULT_URLBASE_HOST)" \
		--set defaultURLBase.domain="\$(DOMAIN)"

install-manager-local: helm kubectl create-cluster ## Install cosmo contoller-manager resources.
	@echo ====== $@ ======
	helm upgrade --install cosmo-controller-manager ../../charts/cosmo-controller-manager \
		-n cosmo-system --create-namespace \
		--wait \
		--set image.repository=cosmo.io:5000/cosmo-controller-manager \
		--set image.tag=latest \
		--set image.pullPolicy=Always \
		--set defaultURLBase.host="\$(DEFAULT_URLBASE_HOST)" \
		--set defaultURLBase.domain="\$(DOMAIN)" \
		--set replicaCount=0 \
		--set localRunTest.enabled=true \
		--set localRunTest.webhookEndpoint.ip=$(shell hostname -i) \
		--set localRunTest.webhookEndpoint.port=9443

install-dashboard: helm kubectl create-cluster docker-build-dashboard ## Install cosmo contoller-manager resources.
	@echo ====== $@ ======
	helm upgrade --install cosmo-dashboard ../../charts/cosmo-dashboard \
		-n cosmo-system \
		--wait \
		--set image.repository=cosmo.io:5000/cosmo-dashboard \
		--set image.tag=latest \
		--set image.pullPolicy=Always \
		--set ingress.enabled=true \
		--set ingress.hosts[0].host=$(DASHBOARD_URL) \
		--set ingress.hosts[0].paths[0].path=/ \
		--set ingress.hosts[0].paths[0].pathType=Prefix \
		--set cert.dnsName=$(DASHBOARD_URL) \
		--set insecure=true \
		--set logLevel=debug

uninstall-manager: helm  ## Uninstall cosmo contoller-manager resources.
	@echo ====== $@ ======
	-helm uninstall -n cosmo-system cosmo-controller-manager

uninstall-dashboard: helm  ## Uninstall cosmo contoller-manager resources.
	@echo ====== $@ ======
	-helm uninstall -n cosmo-system cosmo-dashboard

docker-build-cs:
	@echo ====== $@ ======
	cd templates/code-server-01 && $(MAKE) docker-build

apply-template: kubectl cosmoctl docker-build-cs
	@echo ====== $@ ======
	until (kubectl apply -f ../../config/user-addon/auth-proxy/user-addon-cosmo-auth-proxy-role.yaml) do sleep 1; done
	cd templates/code-server-01 && $(MAKE) apply

BASEURL := https://$(DASHBOARD_URL)

add-user: kubectl cosmoctl
	@echo ====== $@ ======
	-cosmoctl user create tom --admin 2> /dev/null
	export PASSWORD=`cosmoctl user reset-password tom | grep -Po "(?<=New password: ).*")` && \
	bash -c 'until (curl -L -c bin/cookie.txt -X POST $(BASEURL)/api/v1alpha1/auth/login -d "{\"id\":\"tom\",\"password\":\"$${PASSWORD}\"}") do sleep 1; done' && \
	curl -L -b bin/cookie.txt -X PUT $(BASEURL)/api/v1alpha1/user/tom/password -d "{\"currentPassword\":\"$${PASSWORD}\",\"newPassword\": \"vvv\"}"

add-workspace: kubectl cosmoctl
	@echo ====== $@ ======
	-cosmoctl workspace create --user=tom --template=code-server-01 ws1
	sleep 5
	-cosmoctl workspace open-port --user=tom ws1 --name=dash1 --port=7701

##---------------------------------------------------------------------
##@ Execute test
##---------------------------------------------------------------------
.PHONY: run rerun stop

run-manager:
	@echo ====== $@ ======
	@#(cd ../.. && bash hack/download-certs.sh webhook-server-cert && make run QUICK_BUILD=yes)
	cd ../.. && bash hack/download-certs.sh webhook-server-cert && \
		go run ./cmd/controller-manager/main.go \
		--metrics-bind-address :8085 \
		--cert-dir . \
		--zap-log-level=debug \
		--workspace-default-urlbase=https://$(DEFAULT_URLBASE_HOST).$(DOMAIN)

run-local: kubectl ## Run test.
	-@make uninstall-manager
	make install-manager-local
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
		tmux send-keys -t 0 'make run-manager 2>&1 | tee -a $${BIN_DIR}/local-manager.log' Enter && \
		# ======== run dashboard-ui ======== \
		tmux send-keys -t 1 '(cd ../.. && make run-dashboard-ui 2>&1 | tee -a $${BIN_DIR}/local-dashboard-ui.log)' Enter && \
		# ======== run dashboard ======== \
		tmux send-keys -t 2 '(cd ../.. && make run-dashboard QUICK_BUILD=yes 2>&1 | tee -a $${BIN_DIR}/local-dashboard.log)' Enter && \
		# ======== move to pane 3 ======== \
		tmux select-pane -t 3 && \
		# ======== attach ======== \
		tmux attach-session -d -t cosmo && \
		echo end ; \
	}

rerun: stop run ## Rerun test.

stop: ## Stop test.
	-@tmux kill-session -t cosmo 2> /dev/null

##---------------------------------------------------------------------
##@ Build Binary
##---------------------------------------------------------------------
GO_SRC := $(shell find ../../internal ../../pkg -type f -regex ".*\.go")
cosmoctl: bin bin/cosmoctl ## cosmoctl
bin/cosmoctl: ../../cmd/cosmoctl/main.go $(GO_SRC)
	@echo ====== $@ ======
	CGO_ENABLED=0 go build -o bin/cosmoctl ../../cmd/cosmoctl/main.go

##---------------------------------------------------------------------
##@ Download Binary
##---------------------------------------------------------------------
bin:
	mkdir -p $(BIN)

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


##---------------------------------------------------------------------
##@ Utility
##---------------------------------------------------------------------
bash: ## Activate kubeconfig for local k8s. 
	@bash --rcfile bashrc 

helm-ls: ## helm list
	@helm list -a -A

k: ## Get k0s resources.
	@kubectl get node
	@kubectl get po -A
	@kubectl get ing -A
	@kubectl get svc -A
	@kubectl get ep -A
