module github.com/cosmo-workspace/cosmo

go 1.16

require (
	github.com/evanphx/json-patch/v5 v5.3.0
	github.com/go-logr/logr v0.4.0
	github.com/google/go-cmp v0.5.6
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.1
	github.com/mattn/go-isatty v0.0.12
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	github.com/sethvargo/go-password v0.2.0
	github.com/spf13/cobra v1.1.3
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	k8s.io/api v0.22.3
	k8s.io/apiextensions-apiserver v0.21.3 // indirect
	k8s.io/apimachinery v0.22.3
	k8s.io/cli-runtime v0.22.3
	k8s.io/client-go v0.22.3
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a
	sigs.k8s.io/controller-runtime v0.9.2
	sigs.k8s.io/kustomize/api v0.10.0
	sigs.k8s.io/yaml v1.2.0
)
