GOBIN = $(if $(shell go env GOBIN),$(shell go env GOBIN),$(shell go env GOPATH)/bin)

all: generate

generate: go ts

go: install
	yarn go:gen_nofix
	$(GOBIN)/goimports -w ../../api/openapi/dashboard/v1alpha1/*.go

ts: install
	yarn ts:gen

install:
	yarn install
	go install golang.org/x/tools/cmd/goimports@latest
