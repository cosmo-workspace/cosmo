GOBIN = $(if $(shell go env GOBIN),$(shell go env GOBIN),$(shell go env GOPATH)/bin)

all: generate

generate:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOBIN)/buf generate
