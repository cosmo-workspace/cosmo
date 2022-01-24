package dashboard

import (
	"net/http"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"

	//+kubebuilder:scaffold:imports

	"github.com/cosmo-workspace/cosmo/pkg/auth"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient kosmo.Client
	testEnv   *envtest.Environment

	userSession  []*http.Cookie
	adminSession []*http.Cookie
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
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

	err = cosmov1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = wsv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	c, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())

	k8sClient = kosmo.NewClient(c)

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
	})
	Expect(err).NotTo(HaveOccurred())

	// Setup server
	By("bootstrapping server")
	klient := kosmo.NewClient(mgr.GetClient())

	auths := make(map[wsv1alpha1.UserAuthType]auth.Authorizer)
	auths[wsv1alpha1.UserAuthTypeKosmoSecert] = auth.NewKosmoSecretAuthorizer(klient)

	serv := (&Server{
		Log:                 clog.NewLogger(ctrl.Log.WithName("dashboard")),
		Klient:              klient,
		GracefulShutdownDur: time.Second * time.Duration(5),
		ResponseTimeout:     time.Second * time.Duration(5),
		StaticFileDir:       filepath.Join(".", "test"),
		Port:                8888,
		MaxAgeSeconds:       60,
		SessionName:         "test-server",
		Insecure:            true,
		Authorizers:         auths,
	})
	err = mgr.Add(serv)
	Expect(err).NotTo(HaveOccurred())

	ctx := ctrl.SetupSignalHandler()

	go func() {
		err = mgr.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	}()

	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("creating default user")

	user := wsv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "usertest",
		},
		Spec: wsv1alpha1.UserSpec{
			DisplayName: "お名前",
			AuthType:    wsv1alpha1.UserAuthTypeKosmoSecert,
		},
	}
	err = k8sClient.Create(ctx, &user)
	Expect(err).ShouldNot(HaveOccurred())

	ns := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cosmo-user-usertest",
		},
	}
	err = k8sClient.Create(ctx, &ns)
	Expect(err).ShouldNot(HaveOccurred())

	err = klient.RegisterPassword(ctx, "usertest", []byte("password"))
	Expect(err).ShouldNot(HaveOccurred())

	By("creating default admin user")

	user2 := wsv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "usertest-admin",
		},
		Spec: wsv1alpha1.UserSpec{
			DisplayName: "アドミン",
			Role:        wsv1alpha1.UserAdminRole,
			AuthType:    wsv1alpha1.UserAuthTypeKosmoSecert,
		},
	}
	err = k8sClient.Create(ctx, &user2)
	Expect(err).ShouldNot(HaveOccurred())

	ns2 := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cosmo-user-usertest-admin",
		},
	}
	err = k8sClient.Create(ctx, &ns2)
	Expect(err).ShouldNot(HaveOccurred())

	err = klient.RegisterPassword(ctx, "usertest-admin", []byte("password"))
	Expect(err).ShouldNot(HaveOccurred())

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
