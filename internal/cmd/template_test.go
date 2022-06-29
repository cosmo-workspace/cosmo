package cmd

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
)

var _ = Describe("cosmoctl [template]", func() {

	var (
		clientMock kosmo.ClientMock
		rootCmd    *cobra.Command
		options    *cmdutil.CliOptions
		outBuf     *bytes.Buffer
		inBuf      *bytes.Buffer
	)
	consoleOut := func() string {
		out, _ := ioutil.ReadAll(outBuf)
		return string(out)
	}

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		_ = clientgoscheme.AddToScheme(scheme)
		_ = cosmov1alpha1.AddToScheme(scheme)
		_ = wsv1alpha1.AddToScheme(scheme)

		baseclient, err := kosmo.NewClientByRestConfig(cfg, scheme)
		Expect(err).NotTo(HaveOccurred())
		clientMock = kosmo.NewClientMock(baseclient)
		klient := kosmo.NewClient(&clientMock)

		options = cmdutil.NewCliOptions()
		options.Client = &klient
		inBuf = bytes.NewBufferString("")
		outBuf = bytes.NewBufferString("")
		options.In = inBuf
		options.Out = outBuf
		options.ErrOut = outBuf
		options.Scheme = scheme
		rootCmd = NewRootCmd(options)
		By("---------------BeforeEach end----------------")
	})

	AfterEach(func() {
		By("---------------AfterEach start---------------")
		clientMock.Clear()
		test_DeleteTemplateAll()
	})

	//==================================================================================
	Describe("[generate]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					inBuf.WriteString(yamlData)
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "template", "generate", "--workspace", "--workspace-main-service-port-name", "main", "--serviceaccount", "hoge", "--required-vars", "HOGE:HOGEHOGE,FUGA:FUGAFUGA"),
				Entry(nil, "template", "generate", "--workspace", "--workspace-main-service-port-name", "main", "-o", "/tmp/test-cosmo-template"),
				Entry(nil, "template", "generate", "--user-addon", "--set-default-user-addon", "--set-sysns-user-addon", "cosmo-system", "--disable-nameprefix"),
			)
		})

		Describe("fail", func() {

			DescribeTable("with invalid args:",
				func(args ...string) {
					inBuf.WriteString(yamlData)
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "template", "generate", "--workspace", "--user-addon", "--workspace-main-service-port-name", "main"),
			)
		})
	})

	//==================================================================================
	Describe("[get]", func() {

		BeforeEach(func() {
			test_CreateTemplate(wsv1alpha1.TemplateTypeWorkspace, "template1")
			test_CreateTemplate(wsv1alpha1.TemplateTypeWorkspace, "template2")
			test_CreateTemplate(wsv1alpha1.TemplateTypeUserAddon, "template3")
		})

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "template", "get"),
				Entry(nil, "template", "get", "--workspace"),
				Entry(nil, "template", "get", "template2"),
				Entry(nil, "template", "get", "template2", "--workspace"),
			)
		})

		Describe("fail", func() {

			DescribeTable("with invalid args:",
				func(args ...string) {
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "template", "get", "xxxxx"),
			)

			DescribeTable("with an unexpected error at list users:",
				func(args ...string) {
					clientMock.ListMock = func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.RunE$") {
							return true, errors.New("list error")
						}
						return false, nil
					}
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "template", "get"),
				Entry(nil, "template", "get", "--workspace"),
			)
		})
	})

	//==================================================================================
	Describe("[validate]", func() {

		createFile := func(data, fname string) string {
			f, err := os.Create(filepath.Join(os.TempDir(), fname))
			defer func() {
				Ω(f.Close()).ShouldNot(HaveOccurred())
			}()
			Ω(err).ShouldNot(HaveOccurred())
			_, err = f.Write([]byte(data))
			Ω(err).ShouldNot(HaveOccurred())
			return f.Name()
		}

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					inBuf.WriteString(tmplData)
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					o := consoleOut()
					o = regexp.MustCompile(`cosmoctl-validate-[^-]+-`).ReplaceAllString(o, "cosmoctl-validate-XXXXXXXX-")
					Expect(o).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "template", "validate", "--file", createFile(tmplData, "test-template.yaml"), "--vars", "HOGE:hoge,FUGA:fuga"),
				Entry(nil, "template", "validate", "--file", "-"),
				Entry(nil, "template", "validate", "--file", "-", "--client"),
			)
		})

		Describe("fail", func() {

			DescribeTable("with invalid args:",
				func(args ...string) {
					inBuf.WriteString(tmplData)
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "template", "validate"),
				Entry(nil, "template", "validate", "--file"),
				Entry(nil, "template", "validate", "--file", "/tmp/(xx*xx)"),
				Entry(nil, "template", "validate", "--file", createFile("", "test-empty-template.yaml")),
				Entry(nil, "template", "validate", "--file", createFile("hoge", "test-invalid-template.yaml")),
				Entry(nil, "template", "validate", "--file", "-", "--vars", "HOGE"),
				Entry(nil, "template", "validate", "--file", createFile(userAddonTmplData, "test-user-addon-template.yaml")),
			)

			DescribeTable("with an unexpected error at list users:",
				func(args ...string) {
					clientMock.ListMock = func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.RunE$") {
							return true, errors.New("list error")
						}
						return false, nil
					}
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				// Entry(nil, "template", "get"),
				// Entry(nil, "template", "get", "--workspace"),
			)
		})
	})

})

const yamlData = `apiVersion: v1
kind: Service
metadata:
  name: 'workspace'
spec:
  ports:
  - name: main
    port: 3000
    protocol: TCP
  - name: sub
    port: 3001
    protocol: TCP
  type: ClusterIP
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: 'workspace'
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: 'workspace'
spec:
  replicas: 1
  template:
    spec:
      containers:
      - image: theiaide/theia
        imagePullPolicy: IfNotPresent
        name: theia
        ports:
        - containerPort: 3000
          name: http
          protocol: TCP
        volumeMounts:
        - mountPath: /home/project
          name: data
      serviceAccountName: default
      volumes:
      - emptyDir: {}
        name: data
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: 'workspace'
spec:
  rules:
  - host: main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}
    http:
      paths:
      - backend:
          service:
            name: 'workspace'
            port:
              name: main
        path: /*
        pathType: Exact
`

const tmplData = `# Generated by cosmoctl template command
apiVersion: cosmo.cosmo-workspace.github.io/v1alpha1
kind: Template
metadata:
  annotations:
    cosmo/ws-deployment: workspace
    cosmo/ws-ingress: workspace
    cosmo/ws-service: workspace
    cosmo/ws-service-main-port: main
    cosmo/ws-urlbase: \"\"
  creationTimestamp: null
  labels:
    cosmo/type: workspace
  name: cmd
spec:
  rawYaml: |
    apiVersion: rbac.authorization.k8s.io/v1
    kind: RoleBinding
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: '{{TEMPLATE}}'
      name: '{{INSTANCE}}-cosmo-auth-proxy-role'
      namespace: '{{NAMESPACE}}'
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: Role
      name: cosmo-auth-proxy-role
    subjects:
    - kind: ServiceAccount
      name: hoge
      namespace: '{{NAMESPACE}}'
    ---
    apiVersion: v1
    kind: Service
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: '{{TEMPLATE}}'
      name: '{{INSTANCE}}-workspace'
      namespace: '{{NAMESPACE}}'
    spec:
      ports:
      - name: main
        port: 3000
        protocol: TCP
      - name: sub
        port: 3001
        protocol: TCP
      selector:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: '{{TEMPLATE}}'
      type: ClusterIP
    ---
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: '{{TEMPLATE}}'
      name: '{{INSTANCE}}-workspace'
      namespace: '{{NAMESPACE}}'
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 10Gi
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: '{{TEMPLATE}}'
      name: '{{INSTANCE}}-workspace'
      namespace: '{{NAMESPACE}}'
    spec:
      replicas: 1
      selector:
        matchLabels:
          cosmo/instance: '{{INSTANCE}}'
          cosmo/template: '{{TEMPLATE}}'
      template:
        metadata:
          labels:
            cosmo/instance: '{{INSTANCE}}'
            cosmo/template: '{{TEMPLATE}}'
        spec:
          containers:
          - args:
            - --insecure
            env:
            - name: COSMO_AUTH_PROXY_INSTANCE
              value: '{{INSTANCE}}'
            - name: COSMO_AUTH_PROXY_NAMESPACE
              value: '{{NAMESPACE}}'
            image: ghcr.io/cosmo-workspace/cosmo-auth-proxy:latest
            name: cosmo-auth-proxy
          - image: theiaide/theia
            imagePullPolicy: IfNotPresent
            name: theia
            ports:
            - containerPort: 3000
              name: http
              protocol: TCP
            volumeMounts:
            - mountPath: /home/project
              name: data
          serviceAccountName: default
          volumes:
          - emptyDir: {}
            name: data
    ---
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: '{{TEMPLATE}}'
      name: '{{INSTANCE}}-workspace'
      namespace: '{{NAMESPACE}}'
    spec:
      rules:
      - host: main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}
        http:
          paths:
          - backend:
              service:
                name: '{{INSTANCE}}-workspace'
                port:
                  name: main
            path: /*
            pathType: Exact
  requiredVars:
  - default: HOGEHOGE
    var: HOGE
  - default: FUGAFUGA
    var: FUGA`

const userAddonTmplData = `
# Generated by cosmoctl template command
apiVersion: cosmo.cosmo-workspace.github.io/v1alpha1
kind: Template
metadata:
  annotations:
    cosmo/default-user-addon: "true"
  creationTimestamp: null
  labels:
    cosmo/type: user-addon
  name: cosmo-auth-proxy-role
spec:
  description: Role and Rolebinding for COSMO Auth Proxy. By default, it is bound
    to the service account named default in the user namespace.
  rawYaml: |
    apiVersion: rbac.authorization.k8s.io/v1
    kind: Role
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: '{{TEMPLATE}}'
      name: '{{INSTANCE}}-role'
      namespace: '{{NAMESPACE}}'
    rules:
    - apiGroups:
      - workspace.cosmo-workspace.github.io
      resources:
      - workspaces
      verbs:
      - patch
      - update
      - get
      - list
      - watch
    - apiGroups:
      - workspace.cosmo-workspace.github.io
      resources:
      - workspaces/status
      verbs:
      - get
      - list
      - watch
    - apiGroups:
      - cosmo.cosmo-workspace.github.io
      resources:
      - instances
      verbs:
      - patch
      - update
      - get
      - list
      - watch
    - apiGroups:
      - cosmo.cosmo-workspace.github.io
      resources:
      - instances/status
      verbs:
      - get
      - list
      - watch
    - apiGroups:
      - ""
      resources:
      - events
      verbs:
      - create
    - apiGroups:
      - ""
      resources:
      - services
      - secrets
      verbs:
      - get
      - list
      - watch
    - apiGroups:
      - networking.k8s.io
      resources:
      - ingresses
      verbs:
      - get
      - list
      - watch
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: RoleBinding
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: '{{TEMPLATE}}'
      name: '{{INSTANCE}}-rolebinding'
      namespace: '{{NAMESPACE}}'
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: Role
      name: '{{INSTANCE}}-role'
    subjects:
    - kind: ServiceAccount
      name: '{{SERVICE_ACCOUNT}}'
      namespace: '{{NAMESPACE}}'
  requiredVars:
  - default: default
    var: SERVICE_ACCOUNT
  - var: REQUIRED_VAR
`
