
# Image URL to use all building/pushing image targets
VERSION ?=
PRERELEASE ?= false

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

COVER_PROFILE ?= cover.out

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

all: manager cosmoctl dashboard auth-proxy

##@ General

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

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

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

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	make gen-charts

generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

api-generate:
	make -C hack/api-generate generate

chart-crd: manifests
	kustomize build ./config/crd/ > charts/stable/cosmo-controller-manager/templates/crd/crd.yaml

chart-check: chart-crd
	./hack/diff-chart-kust.sh controller-manager
	./hack/diff-chart-kust.sh dashboard

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
export ACK_GINKGO_DEPRECATIONS=1.16.5 ## To silence deprecations message when you execute "go test -v"
export ACK_GINKGO_RC=true             ## To silence deprecations message when you execute "go test -v"
test: manifests generate fmt vet ## Run tests.
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile $(COVER_PROFILE)

ui-test:
	cd web/dashboard-ui && yarn install && yarn test  --coverage  --ci --watchAll=false

swaggerui:
	docker run --rm --name swagger -p 8090:8080 \
		-e SWAGGER_JSON=/cosmo/openapi.yaml -v `pwd`/api/openapi/dashboard/openapi-v1alpha1.yaml:/cosmo swaggerapi/swagger-ui

##@ Build

# Build manager binary
manager: generate fmt vet
	CGO_ENABLED=0 go build -o bin/manager ./cmd/controller-manager/main.go

# Build cosmoctl binary
cosmoctl: generate fmt vet
	CGO_ENABLED=0 go build -o bin/cosmoctl ./cmd/cosmoctl/main.go

# Build dashboard binary
dashboard: generate fmt vet
	CGO_ENABLED=0 go build -o bin/dashboard ./cmd/dashboard/main.go

# Build auth-proxy binary
auth-proxy: generate fmt vet
	CGO_ENABLED=0 go build -o bin/auth-proxy ./cmd/auth-proxy/main.go

# Update version in version.go
update-version:
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
	sed -i.bk -e "s/v[0-9]\+.[0-9]\+.[0-9]\+.* cosmo-workspace/${COSMOCTL_VERSION} cosmo-workspace/" ./cmd/cosmoctl/main.go
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


# Run against the configured Kubernetes cluster in ~/.kube/config
run-dashboard: generate fmt vet manifests
	go run ./cmd/dashboard/main.go \
		--zap-log-level 3 \
		--insecure

run-dashboard-ui:
	cd web/dashboard-ui && yarn install && yarn start

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./cmd/controller-manager/main.go --metrics-bind-address :8085 --cert-dir .


# Build the docker image
docker-build: docker-build-manager docker-build-dashboard docker-build-auth-proxy

# Build the docker image for controller-manager
docker-build-manager: test
	docker build . -t ${IMG_MANAGER} -f dockerfile/controller-manager.Dockerfile

# Build the docker image for dashboard
docker-build-dashboard: test
	docker build . -t ${IMG_DASHBOARD} -f dockerfile/dashboard.Dockerfile

# Build the docker image for auth-proxy
docker-build-auth-proxy: test
	docker build . -t ${IMG_AUTHPROXY} -f dockerfile/auth-proxy.Dockerfile

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -


CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.6.0)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
