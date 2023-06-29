package webhooks

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"

	//+kubebuilder:scaffold:imports

	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

const (
	instControllerFieldManager string = "cosmo-instance-controller"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

const DefaultURLBase = "https://{{NETRULE_GROUP}}-{{INSTANCE}}-{{USER_NAME}}.domain"

func init() {
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme.Scheme))
	//+kubebuilder:scaffold:scheme
}

func TestWebhooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhooks Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	// logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter),
	// 	zap.UseFlagOptions(&zap.Options{
	// 		Development: true,
	// 		Level:       zapcore.Level(-clog.LEVEL_DEBUG_ALL)})))

	ctx, cancel = context.WithCancel(ctrl.SetupSignalHandler())

	By("bootstrapping test environment")

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "config", "webhook")},
		},
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
		CertDir:            testEnv.WebhookInstallOptions.LocalServingCertDir,
		Port:               testEnv.WebhookInstallOptions.LocalServingPort,
	})
	Expect(err).NotTo(HaveOccurred())

	k8sClient = mgr.GetClient()
	Expect(k8sClient).NotTo(BeNil())

	(&InstanceMutationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("InstanceMutationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)

	(&InstanceValidationWebhookHandler{
		Client:       k8sClient,
		Log:          clog.NewLogger(ctrl.Log.WithName("InstanceValidationWebhookHandler")),
		FieldManager: instControllerFieldManager,
	}).SetupWebhookWithManager(mgr)

	(&WorkspaceMutationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("WorkspaceMutationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)

	(&WorkspaceValidationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("WorkspaceValidationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)

	(&UserMutationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("UserMutationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)

	(&UserValidationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("UserValidationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)

	go func() {
		defer GinkgoRecover()
		err := mgr.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	}()

	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
