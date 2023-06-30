# documents
# - buf          https://docs.buf.build/introduction
# - connect      https://connect.build/docs/introduction
# - validate     https://github.com/bufbuild/protoc-gen-validate

GOBIN = $(if $(shell go env GOBIN),$(shell go env GOBIN),$(shell go env GOPATH)/bin)
NODEBIN = node_modules/.bin
export PATH := $(GOBIN):$(NODEBIN):$(PATH)

.PHONY: all generate-go generate-ts

all: generate-go generate-ts

generate-go: $(GOBIN)/buf $(GOBIN)/protoc-gen-connect-go $(GOBIN)/protoc-gen-go $(GOBIN)/protoc-gen-doc $(GOBIN)/protoc-gen-validate
	-rm -rf ./gen
	buf generate --template buf.gen/buf.gen.go.yaml

generate-ts: $(GOBIN)/buf $(NODEBIN)/protoc-gen-es $(NODEBIN)/protoc-gen-connect-web
	-rm -rf ../web/dashboard-ui/src/proto/gen
	buf generate ./proto/dashboard-apis  --template buf.gen/buf.gen.ts.dashboard-ui.yaml

##---------------------------------------------------------------------
##@ Download Binary
##---------------------------------------------------------------------
$(GOBIN)/buf:
	go install github.com/bufbuild/buf/cmd/buf@latest

$(GOBIN)/protoc-gen-connect-go:
	go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest

$(GOBIN)/protoc-gen-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

$(GOBIN)/protoc-gen-validate:
	go install github.com/envoyproxy/protoc-gen-validate@latest

$(GOBIN)/protoc-gen-doc:
	go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest

$(NODEBIN)/protoc-gen-es:
	yarn install

$(NODEBIN)/protoc-gen-connect-web:
	yarn install
