package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
)

var _ = Describe("cosmoctl [template]", func() {

	var (
		clientMock kubeutil.ClientMock
		rootCmd    *cobra.Command
		options    *cmdutil.CliOptions
		outBuf     *bytes.Buffer
		inBuf      *bytes.Buffer
	)
	consoleOut := func() string {
		out, _ := io.ReadAll(outBuf)
		return string(out)
	}

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		utilruntime.Must(clientgoscheme.AddToScheme(scheme))
		utilruntime.Must(cosmov1alpha1.AddToScheme(scheme))
		// +kubebuilder:scaffold:scheme

		baseclient, err := kosmo.NewClientByRestConfig(cfg, scheme)
		Expect(err).NotTo(HaveOccurred())
		clientMock = kubeutil.NewClientMock(baseclient)
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
	})

	AfterEach(func() {
		clientMock.Clear()
		testUtil.DeleteTemplateAll()
		testUtil.DeleteClusterTemplateAll()
	})

	//==================================================================================
	desc := func(args ...string) string { return strings.Join(args, " ") }
	errSnap := func(err error) string {
		if err == nil {
			return "success"
		} else {
			return err.Error()
		}
	}

	//==================================================================================
	Describe("[generate]", func() {

		var versionRegexp = regexp.MustCompile(`v[0-9]+.[0-9]+.[0-9]+.* cosmo-workspace`)

		templateOutputSnapshot := func(output string) string {
			return versionRegexp.ReplaceAllString(output, "vX.X.X cosmo-workspace")
		}

		run_test := func(args ...string) {
			inBuf.WriteString(yamlData)
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			Expect(templateOutputSnapshot(consoleOut())).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(desc, "template", "generate", "--workspace", "--workspace-main-service-port-name", "main", "--required-vars", "HOGE:HOGEHOGE,FUGA:FUGAFUGA"),
			Entry(desc, "template", "generate", "--workspace", "--workspace-main-service-port-name", "main", "-o", "/tmp/test-cosmo-template"),
			Entry(desc, "template", "generate", "--user-addon", "--set-default-user-addon", "--disable-nameprefix"),
			Entry(desc, "template", "generate", "--user-addon", "--set-default-user-addon", "--cluster-scope", "--disable-nameprefix"),
			Entry(desc, "template", "generate", "--workspace", "--userroles", "teama-*", "--forbidden-userroles", "teama-operator,teama-testuser"),
		)

		DescribeTable("❌ fail with invalid args:",
			run_test,
			Entry(desc, "template", "generate", "--workspace", "--user-addon", "--workspace-main-service-port-name", "main"),
		)
	})

	//==================================================================================
	Describe("[get]", func() {

		run_test := func(args ...string) {
			testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeWorkspace, "template1")
			testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeWorkspace, "template2")
			testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeUserAddon, "template3")
			testUtil.CreateClusterTemplate(cosmov1alpha1.TemplateLabelEnumTypeUserAddon, "cluster-template1")
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			Expect(consoleOut()).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(desc, "template", "get"),
			Entry(desc, "template", "get", "--workspace"),
			Entry(desc, "template", "get", "template2"),
			Entry(desc, "template", "get", "template2", "--workspace"),
			Entry(desc, "template", "get", "template2", "template3"),
			Entry(desc, "template", "get", "template2", "cluster-template1", "notfound"),
			Entry(desc, "template", "get", "notfound"),
		)

		DescribeTable("❌ fail with an unexpected error at list users:",
			func(args ...string) {
				clientMock.SetListError("\\.RunE$", errors.New("mock list error"))
				run_test(args...)
			},
			Entry(desc, "template", "get"),
			Entry(desc, "template", "get", "--workspace"),
		)
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

		run_test := func(args ...string) {
			inBuf.WriteString(tmplData)
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			o := consoleOut()
			o = regexp.MustCompile(`cosmoctl-validate-[^-]+-`).ReplaceAllString(o, "cosmoctl-validate-XXXXXXXX-")
			Expect(o).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(desc, "template", "validate", "--file", createFile(tmplData, "test-template.yaml"), "--vars", "DOMAIN:example.com"),
			Entry(desc, "template", "validate", "--file", "-"),
			Entry(desc, "template", "validate", "--file", "-", "--client", "-v", "10"),
		)

		DescribeTable("❌ fail with invalid args:",
			run_test,
			Entry(desc, "template", "validate"),
			Entry(desc, "template", "validate", "--file"),
			Entry(desc, "template", "validate", "--file", "/tmp/(xx*xx)"),
			Entry(desc, "template", "validate", "--file", createFile("", "test-empty-template.yaml")),
			Entry(desc, "template", "validate", "--file", createFile("hoge", "test-invalid-template.yaml")),
			Entry(desc, "template", "validate", "--file", "-", "--vars", "HOGE"),
			Entry(desc, "template", "validate", "--file", createFile(userAddonTmplData, "test-user-addon-template.yaml")),
		)

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
apiVersion: cosmo-workspace.github.io/v1alpha1
kind: Template
metadata:
  annotations:
    workspace.cosmo-workspace.github.io/deployment: workspace
    workspace.cosmo-workspace.github.io/ingress: workspace
    workspace.cosmo-workspace.github.io/service: workspace
    workspace.cosmo-workspace.github.io/service-main-port: main
    workspace.cosmo-workspace.github.io/urlbase: \"\"
  creationTimestamp: null
  labels:
    cosmo-workspace.github.io/type: workspace
  name: cmd
spec:
  rawYaml: |
    apiVersion: rbac.authorization.k8s.io/v1
    kind: RoleBinding
    metadata:
      labels:
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
        cosmo-workspace.github.io/template: '{{TEMPLATE}}'
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
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
        cosmo-workspace.github.io/template: '{{TEMPLATE}}'
      name: '{{INSTANCE}}-workspace'
      namespace: '{{NAMESPACE}}'
    spec:
      ports:
      - name: main
        port: 3000
        protocol: TCP
      selector:
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
        cosmo-workspace.github.io/template: '{{TEMPLATE}}'
      type: ClusterIP
    ---
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      labels:
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
        cosmo-workspace.github.io/template: '{{TEMPLATE}}'
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
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
        cosmo-workspace.github.io/template: '{{TEMPLATE}}'
      name: '{{INSTANCE}}-workspace'
      namespace: '{{NAMESPACE}}'
    spec:
      replicas: 1
      selector:
        matchLabels:
          cosmo-workspace.github.io/instance: '{{INSTANCE}}'
          cosmo-workspace.github.io/template: '{{TEMPLATE}}'
      template:
        metadata:
          labels:
            cosmo-workspace.github.io/instance: '{{INSTANCE}}'
            cosmo-workspace.github.io/template: '{{TEMPLATE}}'
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
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
        cosmo-workspace.github.io/template: '{{TEMPLATE}}'
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
apiVersion: cosmo-workspace.github.io/v1alpha1
kind: Template
metadata:
  annotations:
    useraddon.cosmo-workspace.github.io/default: "true"
  creationTimestamp: null
  labels:
    cosmo-workspace.github.io/type: useraddon
  name: cosmo-auth-proxy-role
spec:
  description: Role and Rolebinding for COSMO Auth Proxy. By default, it is bound
    to the service account named default in the user namespace.
  rawYaml: |
    apiVersion: rbac.authorization.k8s.io/v1
    kind: Role
    metadata:
      labels:
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
        cosmo-workspace.github.io/template: '{{TEMPLATE}}'
      name: '{{INSTANCE}}-role'
      namespace: '{{NAMESPACE}}'
    rules:
    - apiGroups:
      - cosmo-workspace.github.io
      resources:
      - workspaces
      verbs:
      - patch
      - update
      - get
      - list
      - watch
    - apiGroups:
      - cosmo-workspace.github.io
      resources:
      - workspaces/status
      verbs:
      - get
      - list
      - watch
    - apiGroups:
      - cosmo-workspace.github.io
      resources:
      - instances
      verbs:
      - patch
      - update
      - get
      - list
      - watch
    - apiGroups:
      - cosmo-workspace.github.io
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
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
        cosmo-workspace.github.io/template: '{{TEMPLATE}}'
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
