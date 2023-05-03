package cmdutil

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/client-go/tools/clientcmd/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/yaml"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

const kubeconfigFile = "kubeconfig-test"
const kubeconfigFile2 = "kubeconfig-test2"

var testEnv *envtest.Environment

func TestCmdutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmdutil Suite")
}

var _ = BeforeSuite(func() {
	z := zap.New(zap.WriteTo(os.Stdout), zap.UseDevMode(true))
	logf.SetLogger(z)

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	envtestCfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(envtestCfg).NotTo(BeNil())

	// envtestClient, err := client.New(envtestCfg, client.Options{Scheme: k8scheme.Scheme})
	// Expect(err).NotTo(HaveOccurred())
	// Expect(envtestClient).NotTo(BeNil())

	// var sa *corev1.ServiceAccount
	// err = envtestClient.Get(context.TODO(), types.NamespacedName{Name: "default", Namespace: "default"}, sa)
	// Expect(err).NotTo(HaveOccurred())
	// Expect(sa).NotTo(BeNil())

	// var secret *corev1.Secret
	// err = envtestClient.Get(context.TODO(), types.NamespacedName{Name: sa.Secrets[0].Name, Namespace: "default"}, secret)
	// Expect(err).NotTo(HaveOccurred())
	// Expect(secret).NotTo(BeNil())

	// envtestSAToken = secret.Data["token"]

	envtestClusterData := v1.Cluster{
		Server:                envtestCfg.Host,
		InsecureSkipTLSVerify: true,
	}

	By("creating kubeconfig files")

	cfg := v1.Config{
		Clusters: []v1.NamedCluster{
			{
				Name:    "envtest",
				Cluster: envtestClusterData,
			},
		},
		Contexts: []v1.NamedContext{
			{
				Name: "foo-cluster",
				Context: v1.Context{
					Cluster:   "envtest",
					Namespace: "cosmo-user-foo",
				},
			},
			{
				Name: "bar-cluster",
				Context: v1.Context{
					Cluster:   "envtest",
					Namespace: "bar",
				},
			},
		},
		CurrentContext: "foo-cluster",
	}
	b, err := yaml.Marshal(cfg)
	Expect(err).ShouldNot(HaveOccurred())
	CreateFile(".", kubeconfigFile, b)

	cfg2 := v1.Config{
		Clusters: []v1.NamedCluster{
			{
				Name:    "envtest",
				Cluster: envtestClusterData,
			},
		},
		Contexts: []v1.NamedContext{
			{
				Name: "foo-cluster",
				Context: v1.Context{
					Cluster:   "envtest",
					Namespace: "cosmo-user-default",
				},
			},
		},
		CurrentContext: "foo-cluster",
	}
	b2, err := yaml.Marshal(cfg2)
	Expect(err).ShouldNot(HaveOccurred())
	CreateFile(".", kubeconfigFile2, b2)

})

var _ = AfterSuite(func() {
	By("removing kubeconfig file")
	var err error
	err = RemoveFile(".", kubeconfigFile)
	Expect(err).ShouldNot(HaveOccurred())
	err = RemoveFile(".", kubeconfigFile2)
	Expect(err).ShouldNot(HaveOccurred())

	By("tearing down the test environment")
	err = testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("CliOptions", func() {
	Context("when using default kubeconfig", func() {
		It("should create client and logger", func() {
			os.Setenv("KUBECONFIG", kubeconfigFile)

			logLevel := clog.LEVEL_DEBUG_ALL
			o := NewCliOptions()
			o.LogLevel = logLevel
			o.Out = GinkgoWriter
			o.ErrOut = GinkgoWriter

			var err error
			err = o.Validate(nil, []string{})
			Expect(err).ShouldNot(HaveOccurred())

			err = o.Complete(nil, []string{})
			// Expect(err).ShouldNot(HaveOccurred())

			Expect(o.Logr).ShouldNot(BeNil())
			// Expect(o.Client).ShouldNot(BeNil())
		})
	})

	Context("when kubeconfig is specified", func() {
		It("should create client and logger with given kubeconfig", func() {
			os.Setenv("KUBECONFIG", "notfound")

			logLevel := clog.LEVEL_DEBUG_ALL
			kubeconfigFilePath := path.Join(".", kubeconfigFile)
			o := NewCliOptions()
			o.LogLevel = logLevel
			o.Out = GinkgoWriter
			o.ErrOut = GinkgoWriter
			o.KubeConfigPath = kubeconfigFilePath

			var err error
			err = o.Validate(nil, []string{})
			Expect(err).ShouldNot(HaveOccurred())

			err = o.Complete(nil, []string{})
			// Expect(err).ShouldNot(HaveOccurred())

			Expect(o.Logr).ShouldNot(BeNil())
			// Expect(o.Client).ShouldNot(BeNil())
		})
	})
})

var _ = Describe("NamespacedCliOptions", func() {
	Context("when namespace is specified", func() {
		It("should use given namespace", func() {
			os.Setenv("KUBECONFIG", kubeconfigFile)

			logLevel := clog.LEVEL_DEBUG_ALL
			o := NewNamespacedCliOptions(NewCliOptions())
			o.LogLevel = logLevel
			o.Out = GinkgoWriter
			o.ErrOut = GinkgoWriter

			o.Namespace = "testtest"

			var err error
			err = o.Validate(nil, []string{})
			Expect(err).ShouldNot(HaveOccurred())

			err = o.Complete(nil, []string{})
			// Expect(err).ShouldNot(HaveOccurred())

			Expect(o.Logr).ShouldNot(BeNil())
			// Expect(o.Client).ShouldNot(BeNil())

			Expect(o.Namespace).Should(Equal("testtest"))
			Expect(o.AllNamespace).Should(BeFalse())
		})
	})

	Context("when all-namespaces is specified", func() {
		It("should use all-namespaces", func() {
			os.Setenv("KUBECONFIG", kubeconfigFile)

			logLevel := clog.LEVEL_DEBUG_ALL
			o := NewNamespacedCliOptions(NewCliOptions())
			o.LogLevel = logLevel
			o.Out = GinkgoWriter
			o.ErrOut = GinkgoWriter

			o.AllNamespace = true

			var err error
			err = o.Validate(nil, []string{})
			Expect(err).ShouldNot(HaveOccurred())

			err = o.Complete(nil, []string{})
			// Expect(err).ShouldNot(HaveOccurred())

			Expect(o.Logr).ShouldNot(BeNil())
			// Expect(o.Client).ShouldNot(BeNil())

			Expect(o.Namespace).Should(BeEmpty())
			Expect(o.AllNamespace).Should(BeTrue())
		})
	})

	Context("when all-namespaces nor name specified and kubeconfig current context found", func() {
		It("should use kubeconfig current context namespace", func() {
			os.Setenv("KUBECONFIG", kubeconfigFile2)

			logLevel := clog.LEVEL_DEBUG_ALL
			o := NewNamespacedCliOptions(NewCliOptions())
			o.LogLevel = logLevel
			o.Out = GinkgoWriter
			o.ErrOut = GinkgoWriter

			var err error
			err = o.Validate(nil, []string{})
			Expect(err).ShouldNot(HaveOccurred())

			err = o.Complete(nil, []string{})
			// Expect(err).ShouldNot(HaveOccurred())

			Expect(o.Logr).ShouldNot(BeNil())
			// Expect(o.Client).ShouldNot(BeNil())

			Expect(o.Namespace).Should(BeEquivalentTo("cosmo-user-default"))
			Expect(o.AllNamespace).Should(BeFalse())
		})
	})

	Context("when all-namespaces nor name specified and given context found in kubeconfig", func() {
		It("should use kubeconfig given context namespace", func() {
			os.Setenv("KUBECONFIG", kubeconfigFile)

			logLevel := clog.LEVEL_DEBUG_ALL
			o := NewNamespacedCliOptions(NewCliOptions())
			o.LogLevel = logLevel
			o.Out = GinkgoWriter
			o.ErrOut = GinkgoWriter

			kubecontext := "bar-cluster"
			o.KubeContext = kubecontext

			var err error
			err = o.Validate(nil, []string{})
			Expect(err).ShouldNot(HaveOccurred())

			err = o.Complete(nil, []string{})
			// Expect(err).ShouldNot(HaveOccurred())

			Expect(o.Logr).ShouldNot(BeNil())
			// Expect(o.Client).ShouldNot(BeNil())

			Expect(o.Namespace).Should(BeEquivalentTo("bar"))
			Expect(o.AllNamespace).Should(BeFalse())
		})
	})

	Context("when allnamespaces and namespace are specified", func() {
		It("should return error", func() {
			os.Setenv("KUBECONFIG", kubeconfigFile)

			logLevel := clog.LEVEL_DEBUG_ALL
			o := NewNamespacedCliOptions(NewCliOptions())
			o.LogLevel = logLevel
			o.Out = GinkgoWriter
			o.ErrOut = GinkgoWriter

			o.Namespace = "testtest"
			o.AllNamespace = true

			err := o.Validate(nil, []string{})
			Expect(err).Should(HaveOccurred())
		})
	})
})

var _ = Describe("UserNamespacedCliOptions", func() {
	Context("when user is specified", func() {
		It("should use kubeconfig current context namespace", func() {
			os.Setenv("KUBECONFIG", kubeconfigFile2)

			logLevel := clog.LEVEL_DEBUG_ALL
			o := NewNamespacedCliOptions(NewCliOptions())
			o.LogLevel = logLevel
			o.Out = GinkgoWriter
			o.ErrOut = GinkgoWriter

			var err error
			err = o.Validate(nil, []string{})
			Expect(err).ShouldNot(HaveOccurred())

			err = o.Complete(nil, []string{})
			// Expect(err).ShouldNot(HaveOccurred()) // UnexpectedServerResponse

			Expect(o.Logr).ShouldNot(BeNil())
			// Expect(o.Client).ShouldNot(BeNil())

			Expect(o.Namespace).Should(Equal("cosmo-user-default"))
			Expect(o.AllNamespace).Should(BeFalse())
		})
	})
})
