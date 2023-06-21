package authproxy

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"

	"github.com/cosmo-workspace/cosmo/internal/webhooks"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo/test"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg         *rest.Config
	kosmoClient kosmo.Client
	k8sClient   client.Client
	testUtil    test.TestUtil
	testEnv     *envtest.Environment
	scheme      = runtime.NewScheme()

	wsMgrCtx    context.Context
	WsMgrCancel context.CancelFunc
)

// func TestAuthproxyController(t *testing.T) {
// 	RegisterFailHandler(Fail)
// 	RunSpecs(t, "Authproxy Controller Suite")
// }

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "config", "webhook")},
		},
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	testUtil = test.NewTestUtil(k8sClient)

	kosmoClient = kosmo.NewClient(k8sClient)
	Expect(kosmoClient).NotTo(BeNil())

	wsMgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
		CertDir:            testEnv.WebhookInstallOptions.LocalServingCertDir,
		Port:               testEnv.WebhookInstallOptions.LocalServingPort,
	})
	Expect(err).NotTo(HaveOccurred())

	// webhook
	(&webhooks.InstanceMutationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("InstanceMutationWebhookHandler")),
	}).SetupWebhookWithManager(wsMgr)

	(&webhooks.InstanceValidationWebhookHandler{
		Client:       k8sClient,
		Log:          clog.NewLogger(ctrl.Log.WithName("InstanceValidationWebhookHandler")),
		FieldManager: "cosmo-instance-controller",
	}).SetupWebhookWithManager(wsMgr)

	(&webhooks.WorkspaceMutationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("WorkspaceMutationWebhookHandler")),
	}).SetupWebhookWithManager(wsMgr)

	(&webhooks.WorkspaceValidationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("WorkspaceValidationWebhookHandler")),
	}).SetupWebhookWithManager(wsMgr)

	(&webhooks.UserMutationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("UserMutationWebhookHandler")),
	}).SetupWebhookWithManager(wsMgr)

	(&webhooks.UserValidationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("UserValidationWebhookHandler")),
	}).SetupWebhookWithManager(wsMgr)

	wsMgrCtx, WsMgrCancel = context.WithCancel(ctrl.SetupSignalHandler())

	go func() {
		defer GinkgoRecover()
		err := wsMgr.Start(wsMgrCtx)
		Expect(err).NotTo(HaveOccurred())
	}()
})

var _ = AfterSuite(func() {
	WsMgrCancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
