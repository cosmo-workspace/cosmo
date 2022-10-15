GOBIN = $(if $(shell go env GOBIN),$(shell go env GOBIN),$(shell go env GOPATH)/bin)

all: generate

generate: go ts

go: install
	yarn go:gen_nofix
	cp go-server/auth_request_impl.go_ ../../api/openapi/dashboard/v1alpha1/auth_request_impl.go
	$(GOBIN)/goimports -w ../../api/openapi/dashboard/v1alpha1/*.go

ts: install
	yarn ts:gen

install:
	yarn install
	go install golang.org/x/tools/cmd/goimports@latest
