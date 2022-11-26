package kosmo

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

var etmpl1 *cosmov1alpha1.Template
var einst1 *cosmov1alpha1.Instance
var etmpl2 *cosmov1alpha1.Template
var einst2 *cosmov1alpha1.Instance
var einst2Pod *corev1.Pod

var ectmpl1 *cosmov1alpha1.ClusterTemplate
var ecinst1 *cosmov1alpha1.ClusterInstance

func init() {
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme.Scheme))
	//+kubebuilder:scaffold:scheme
}

func TestKosmo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kosmo Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = NewClientByRestConfig(cfg, scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	ctx := context.Background()
	Expect(k8sClient.Create(ctx, etmpl1)).ShouldNot(HaveOccurred())
	Expect(k8sClient.Create(ctx, einst1)).ShouldNot(HaveOccurred())
	Expect(k8sClient.Create(ctx, etmpl2)).ShouldNot(HaveOccurred())
	Expect(k8sClient.Create(ctx, einst2)).ShouldNot(HaveOccurred())
	Expect(k8sClient.Create(ctx, einst2Pod)).ShouldNot(HaveOccurred())

	Expect(k8sClient.Create(ctx, ecinst1)).ShouldNot(HaveOccurred())
	Expect(k8sClient.Create(ctx, ectmpl1)).ShouldNot(HaveOccurred())

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
