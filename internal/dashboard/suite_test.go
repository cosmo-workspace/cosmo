package dashboard

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bufbuild/connect-go"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	//+kubebuilder:scaffold:imports

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/internal/webhooks"
	"github.com/cosmo-workspace/cosmo/pkg/auth"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo/test"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg        *rest.Config
	k8sClient  kosmo.Client
	clientMock kubeutil.ClientMock
	testUtil   test.TestUtil
	testEnv    *envtest.Environment
	ctx        context.Context
	cancel     context.CancelFunc
)

const DefaultURLBase = "https://{{NETRULE_GROUP}}-{{INSTANCE}}-{{USER_NAME}}.domain"

func init() {
	utilruntime.Must(cosmov1alpha1.AddToScheme(clientgoscheme.Scheme))
	//+kubebuilder:scaffold:scheme
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dashboard Suite")
}

var _ = BeforeSuite(func() {
	// opts := zap.Options{TimeEncoder: zapcore.ISO8601TimeEncoder}
	// logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true), zap.UseFlagOptions(&opts)))
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

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

	c, err := client.New(cfg, client.Options{Scheme: clientgoscheme.Scheme})
	Expect(err).NotTo(HaveOccurred())

	k8sClient = kosmo.NewClient(c)
	Expect(k8sClient).NotTo(BeNil())

	testUtil = test.NewTestUtil(k8sClient)

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             clientgoscheme.Scheme,
		MetricsBindAddress: "0",
		CertDir:            testEnv.WebhookInstallOptions.LocalServingCertDir,
		Port:               testEnv.WebhookInstallOptions.LocalServingPort,
	})
	Expect(err).NotTo(HaveOccurred())

	// webhook
	(&webhooks.InstanceMutationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("InstanceMutationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)

	(&webhooks.InstanceValidationWebhookHandler{
		Client:       k8sClient,
		Log:          clog.NewLogger(ctrl.Log.WithName("InstanceValidationWebhookHandler")),
		FieldManager: "cosmo-instance-controller",
	}).SetupWebhookWithManager(mgr)

	(&webhooks.WorkspaceMutationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("WorkspaceMutationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)

	(&webhooks.WorkspaceValidationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("WorkspaceValidationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)

	(&webhooks.UserMutationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("UserMutationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)

	(&webhooks.UserValidationWebhookHandler{
		Client: k8sClient,
		Log:    clog.NewLogger(ctrl.Log.WithName("UserValidationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)

	// Setup server
	By("bootstrapping server")
	clientMock = kubeutil.NewClientMock(mgr.GetClient())
	klient := kosmo.NewClient(&clientMock)

	auths := make(map[cosmov1alpha1.UserAuthType]auth.Authorizer)
	auths[cosmov1alpha1.UserAuthTypePasswordSecert] = auth.NewPasswordSecretAuthorizer(klient)
	auths[cosmov1alpha1.UserAuthTypeLDAP] = auth.NewMockAuthorizer(
		map[string]string{
			"ldap-user": "password",
		},
	)

	serv := (&Server{
		Log:                 clog.NewLogger(ctrl.Log.WithName("dashboard")),
		Klient:              klient,
		GracefulShutdownDur: time.Second * time.Duration(4),
		ResponseTimeout:     time.Second * time.Duration(4),
		StaticFileDir:       filepath.Join(".", "test"),
		Port:                8888,
		MaxAgeSeconds:       60,
		TLSPrivateKeyPath:   "",
		TLSCertPath:         "",
		Insecure:            true,
		CookieDomain:        "test.domain",
		CookieHashKey:       "----+----1----+----2----+----3----+----4----+----5----+----6----",
		CookieBlockKey:      "----+----1----+----2----+----3--",
		CookieSessionName:   "test-server",
		Authorizers:         auths,
	})
	err = mgr.Add(serv)
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err := mgr.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	}()

})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func test_Login(userName string, password string) string {
	var res *connect.Response[dashv1alpha1.LoginResponse]
	client := dashboardv1alpha1connect.NewAuthServiceClient(http.DefaultClient, "http://localhost:8888")
	Eventually(func() (err error) {
		res, err = client.Login(ctx, connect.NewRequest(&dashv1alpha1.LoginRequest{UserName: userName, Password: password}))
		return err
	}, time.Second*5, time.Millisecond*100).Should(Succeed())

	return res.Header().Get("Set-Cookie")
}

func test_CreateLoginUserSession(userName, displayName string, role []cosmov1alpha1.UserRole, password string) string {
	testUtil.CreateLoginUser(userName, displayName, role, cosmov1alpha1.UserAuthTypePasswordSecert, password)
	return test_Login(userName, password)
}

func NewRequestWithSession[T any](message *T, session string) *connect.Request[T] {
	req := connect.NewRequest(message)
	if session != "" {
		req.Header().Add("Cookie", session)
	}
	return req
}
