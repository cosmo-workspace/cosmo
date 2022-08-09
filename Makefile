
# Image URL to use all building/pushing image targets
VERSION ?=
PRERELEASE ?= false
QUICK_BUILD ?= no

MANAGER_VERSION   ?= $(VERSION)
DASHBOARD_VERSION ?= $(VERSION)
COSMOCTL_VERSION  ?= $(VERSION)
AUTHPROXY_VERSION ?= $(VERSION)

CHART_MANAGER_VERSION   ?= $(MANAGER_VERSION)
CHART_DASHBOARD_VERSION ?= $(DASHBOARD_VERSION)

IMG_MANAGER ?= cosmo-controller-manager:$(MANAGER_VERSION)
IMG_DASHBOARD ?= cosmo-dashboard:$(DASHBOARD_VERSION)
IMG_AUTHPROXY ?= cosmo-auth-proxy:$(AUTHPROXY_VERSION)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,generateEmbeddedObjectMeta=true,preserveUnknownFields=false"

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION ?= 1.21.x

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: manager cosmoctl dashboard auth-proxy

##---------------------------------------------------------------------
##@ General
##---------------------------------------------------------------------

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##---------------------------------------------------------------------
##@ Development
##---------------------------------------------------------------------
define WEBHOOK_CHART_SUFIX
---
{{- if not $$.Values.enableCertManager }}
apiVersion: v1
kind: Secret
metadata:
  name: webhook-server-cert
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "cosmo-controller-manager.labels" . | nindent 4 }}
type: kubernetes.io/tls
data:
  ca.crt: {{ $$tls.caCert }}
  tls.crt: {{ $$tls.clientCert }}
  tls.key: {{ $$tls.clientKey }}
{{- else }}
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  labels:
    {{- include "cosmo-controller-manager.labels" . | nindent 4 }}
  name: cosmo-serving-cert
  namespace: {{ .Release.Namespace }}
spec:
  dnsNames:
  - cosmo-webhook-service.{{ .Release.Namespace }}.svc
  - cosmo-webhook-service.{{ .Release.Namespace }}.svc.cluster.local
  issuerRef:
    kind: ClusterIssuer
    name: cosmo-selfsigned-clusterissuer
  secretName: webhook-server-cert
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  labels:
    {{- include "cosmo-controller-manager.labels" . | nindent 4 }}
  name: cosmo-selfsigned-clusterissuer
  namespace: {{ .Release.Namespace }}
spec:
  selfSigned: {}
{{- end }}
endef

WEBHOOK_CHART_YAML ?= charts/cosmo-controller-manager/templates/webhook.yaml

export WEBHOOK_CHART_SUFIX
gen-charts:
	cp config/crd/bases/* charts/cosmo-controller-manager/crds/
	kustomize build config/webhook-chart \
		| sed -e 's/namespace: system/namespace: {{ .Release.Namespace }}/g' \
		| sed -z 's;apiVersion: v1\nkind: Service\nmetadata:\n  name: cosmo-webhook-service\n  namespace: {{ .Release.Namespace }}\nspec:\n  ports:\n  - port: 443\n    targetPort: 9443\n  selector:\n    control-plane: controller-manager;{{ $$tls := fromYaml ( include "cosmo-controller-manager.gen-certs" . ) }};g' \
		| sed -z 's;creationTimestamp: null;{{- if $$.Values.enableCertManager }}\n  annotations:\n    cert-manager.io/inject-ca-from: {{ .Release.Namespace }}/cosmo-serving-cert\n  {{- end }}\n  labels:\n    {{- include "cosmo-controller-manager.labels" . | nindent 4 }};g' \
		| sed -z 's;clientConfig:;clientConfig:\n    caBundle: {{ if not $$.Values.enableCertManager -}}{{ $$tls.caCert }}{{- else -}}Cg=={{ end }};g' > $(WEBHOOK_CHART_YAML)
	echo "$$WEBHOOK_CHART_SUFIX" >> $(WEBHOOK_CHART_YAML)

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
ifeq ($(QUICK_BUILD),no)
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	make gen-charts
endif

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
ifeq ($(QUICK_BUILD),no)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."
endif

.PHONY: api-generate
api-generate:
	make -C hack/api-generate generate

.PHONY: chart-crd
chart-crd: manifests
	kustomize build ./config/crd/ > charts/stable/cosmo-controller-manager/templates/crd/crd.yaml

.PHONY: chart-check
chart-check: chart-crd
	./hack/diff-chart-kust.sh controller-manager
	./hack/diff-chart-kust.sh dashboard

.PHONY: fmt
fmt: ## Run go fmt against code.
ifeq ($(QUICK_BUILD),no)
	go fmt ./...
endif

.PHONY: vet
vet: ## Run go vet against code.
ifeq ($(QUICK_BUILD),no)
	go vet ./...
endif

##---------------------------------------------------------------------
##@ Test
##---------------------------------------------------------------------
TEST_FILES ?= ./...
COVER_PROFILE ?= cover.out
#TEST_OPTS ?= --ginkgo.focus 'Dashboard server \[User\]' -ginkgo.v -ginkgo.progress -test.v > test.out 2>&1

.PHONY: go-test.env
go-test.env: 
	@echo KUBEBUILDER_ASSETS=$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path) > ./.vscode/go-test.env

.PHONY: test
test: manifests generate fmt vet envtest go-test.env go-test ## Run tests.

.PHONY: go-test
go-test:
ifeq ($(QUICK_BUILD),no)
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" \
	go test $(TEST_FILES) -coverprofile $(COVER_PROFILE) $(TEST_OPTS) -v
endif

.PHONY: test-all-k8s-versions
test-all-k8s-versions: manifests generate fmt vet envtest ## Run tests on targeting k8s versions.
ifeq ($(QUICK_BUILD),no)
	-KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use 1.19.x -p path)" go test ./... -coverprofile $(COVER_PROFILE)
	-KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use 1.20.x -p path)" go test ./... -coverprofile $(COVER_PROFILE)
	-KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use 1.21.x -p path)" go test ./... -coverprofile $(COVER_PROFILE)
	-KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use 1.22.x -p path)" go test ./... -coverprofile $(COVER_PROFILE)
	-KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use 1.23.x -p path)" go test ./... -coverprofile $(COVER_PROFILE)
endif

.PHONY: ui-test
ui-test: ## Run UI tests.
	cd web/dashboard-ui && yarn install && yarn test  --coverage  --ci --watchAll=false

.PHONY: swaggerui
swaggerui:
	docker run --rm --name swagger -p 8090:8080 \
		-e SWAGGER_JSON=/cosmo/openapi.yaml -v `pwd`/api/openapi/dashboard/openapi-v1alpha1.yaml:/cosmo swaggerapi/swagger-ui

##---------------------------------------------------------------------
##@ Build
##---------------------------------------------------------------------
.PHONY: manager
manager: generate fmt vet ## Build manager binary.
	CGO_ENABLED=0 go build -o bin/manager ./cmd/controller-manager/main.go

.PHONY: cosmoctl
cosmoctl: generate fmt vet ## Build cosmoctl binary.
	CGO_ENABLED=0 go build -o bin/cosmoctl ./cmd/cosmoctl/main.go

.PHONY: dashboard
dashboard: generate fmt vet ## Build dashboard binary.
	CGO_ENABLED=0 go build -o bin/dashboard ./cmd/dashboard/main.go

.PHONY: auth-proxy
auth-proxy: generate fmt vet ## Build auth-proxy binary.
	CGO_ENABLED=0 go build -o bin/auth-proxy ./cmd/auth-proxy/main.go

.PHONY: update-version
update-version: ## Update version in version.go.
ifndef VERSION
	@echo "Usage: make update-version VERSION=v9.9.9"
	@exit 9
else
ifeq ($(shell expr $(VERSION) : '^v[0-9]\+\.[0-9]\+\.[0-9]\+$$'),0)
	@echo "Usage: make update-version VERSION=v9.9.9"
	@exit 9
endif
endif
	sed -i.bk -e "s/v[0-9]\+.[0-9]\+.[0-9]\+.* cosmo-workspace/${MANAGER_VERSION} cosmo-workspace/" ./cmd/controller-manager/main.go
	sed -i.bk -e "s/v[0-9]\+.[0-9]\+.[0-9]\+.* cosmo-workspace/${DASHBOARD_VERSION} cosmo-workspace/" ./cmd/dashboard/main.go
	sed -i.bk -e "s/v[0-9]\+.[0-9]\+.[0-9]\+.* cosmo-workspace/${COSMOCTL_VERSION} cosmo-workspace/" ./internal/cmd/root_cmd.go
	sed -i.bk -e "s/v[0-9]\+.[0-9]\+.[0-9]\+.* cosmo-workspace/${AUTHPROXY_VERSION} cosmo-workspace/" ./cmd/auth-proxy/main.go
	cd config/manager && kustomize edit set image controller=${IMG_MANAGER}
	cd config/dashboard && kustomize edit set image dashboard=${IMG_DASHBOARD}
	sed -i.bk \
		-e "s/version: [0-9]\+.[0-9]\+.[0-9]\+.*/version: ${CHART_MANAGER_VERSION:v%=%}/" \
		-e "s/appVersion: v[0-9]\+.[0-9]\+.[0-9]\+.*/appVersion: ${MANAGER_VERSION}/" \
		-e 's;artifacthub.io/prerelease: "\(true\|false\)";artifacthub.io/prerelease: "$(PRERELEASE)";' \
		charts/cosmo-controller-manager/Chart.yaml
	sed -i.bk \
		-e "s/version: [0-9]\+.[0-9]\+.[0-9]\+.*/version: ${CHART_DASHBOARD_VERSION:v%=%}/" \
		-e "s/appVersion: v[0-9]\+.[0-9]\+.[0-9]\+.*/appVersion: ${DASHBOARD_VERSION}/" \
		-e 's;artifacthub.io/prerelease: "\(true\|false\)";artifacthub.io/prerelease: "$(PRERELEASE)";' \
		charts/cosmo-dashboard/Chart.yaml

##---------------------------------------------------------------------
##@ Run
##---------------------------------------------------------------------
# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run-dashboard
run-dashboard: generate fmt vet manifests ## Run dashboard against the configured Kubernetes cluster in ~/.kube/config.
	go run ./cmd/dashboard/main.go \
		--zap-log-level 3 \
		--insecure

.PHONY: run-dashboard-ui
run-dashboard-ui: ## Run dashboard-ui.
	cd web/dashboard-ui && yarn install && yarn start

.PHONY: run-auth-proxy
run-auth-proxy: generate fmt vet manifests ## Run auth-proxy against the configured Kubernetes cluster in ~/.kube/config.
	go run ./cmd/auth-proxy/main.go \
		--zap-log-level 3 \
		--insecure

.PHONY: run-auth-proxy-ui
run-auth-proxy-ui: ## Run auth-proxy-ui.
	cd web/auth-proxy-ui && yarn install && PORT=3010 yarn start

.PHONY: run
run: generate fmt vet manifests ## Run controller-manager against the configured Kubernetes cluster in ~/.kube/config.
	go run ./cmd/controller-manager/main.go --metrics-bind-address :8085 --cert-dir .

##---------------------------------------------------------------------
##@ Docker build
##---------------------------------------------------------------------
.PHONY: docker-build
docker-build: docker-build-manager docker-build-dashboard docker-build-auth-proxy ## Build the docker image.

.PHONY: docker-build-manager
docker-build-manager: test ## Build the docker image for controller-manager.
	DOCKER_BUILDKIT=1 docker build . -t ${IMG_MANAGER} -f dockerfile/controller-manager.Dockerfile

.PHONY: docker-build-dashboard
docker-build-dashboard: test ## Build the docker image for dashboard.
	DOCKER_BUILDKIT=1 docker build . -t ${IMG_DASHBOARD} -f dockerfile/dashboard.Dockerfile

.PHONY: docker-build-auth-proxy
docker-build-auth-proxy: test ## Build the docker image for auth-proxy.
	DOCKER_BUILDKIT=1 docker build . -t ${IMG_AUTHPROXY} -f dockerfile/auth-proxy.Dockerfile

##---------------------------------------------------------------------
##@ Deployment
##---------------------------------------------------------------------
ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

##---------------------------------------------------------------------
##@ Build Dependencies
##---------------------------------------------------------------------
## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.6.0

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
