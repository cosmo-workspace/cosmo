package authproxy

import (
	"context"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/internal/authproxy/proxy"
	"github.com/cosmo-workspace/cosmo/pkg/auth"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
)

var _ = Describe("auth-proxy controller", func() {

	var (
		ctx          context.Context
		cancel       context.CancelFunc
		clientMock   kubeutil.ClientMock
		proxyManager *proxy.Manager
	)

	startManager := func(username, workspaceName string) {
		By("---- Manager setting start ----")
		logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: "0",
			LeaderElection:     false,
			Namespace:          "cosmo-user-" + username,
			//		CertDir:            testEnv.WebhookInstallOptions.LocalServingCertDir,
			Port: testEnv.WebhookInstallOptions.LocalServingPort,
		})
		Expect(err).NotTo(HaveOccurred())

		clientMock = kubeutil.NewClientMock(mgr.GetClient())
		klient := kosmo.NewClient(&clientMock)

		authorizer := auth.NewPasswordSecretAuthorizer(mgr.GetClient())
		proxyManager, err = (&proxy.Manager{
			Log:                      clog.NewLogger(ctrl.Log.WithName("proxy-manager")),
			ProxyBackendScheme:       "http",
			ProxyGracefulShutdownDur: time.Second * time.Duration(10),
			ProxyStartupCheckTimeout: time.Second * time.Duration(10),
			Insecure:                 true,
			TLSCertPath:              "",
			TLSPrivateKeyPath:        "",
			User:                     username,
			MaxAgeSeconds:            60 * 720,
			Authorizer:               authorizer,
		}).Initialize()
		Expect(err).NotTo(HaveOccurred())

		reconciler := NetworkRuleReconciler{
			Client:        klient,
			Recorder:      mgr.GetEventRecorderFor("cosmo-auth-proxy"),
			Scheme:        mgr.GetScheme(),
			ProxyManager:  proxyManager,
			WorkspaceName: workspaceName,
		}
		err = reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		err = mgr.AddHealthzCheck("healthz", healthz.Ping)
		Expect(err).NotTo(HaveOccurred())
		err = mgr.AddReadyzCheck("readyz", healthz.Ping)
		Expect(err).NotTo(HaveOccurred())

		ctx, cancel = context.WithCancel(context.Background())
		By("---- Manager setting complete ----")

		go func() {
			By("----Start Manager----")
			err := mgr.Start(ctx)
			Expect(err).NotTo(HaveOccurred())
		}()
	}

	stopManager := func() {
		By("----Stop Manager in porgress---")
		cancel()
		time.Sleep(100 * time.Millisecond)
	}

	workspaceSnap := func(ws *cosmov1alpha1.Workspace) struct{ Name, Namespace, Spec, Status interface{} } {
		if ws == nil {
			return struct{ Name, Namespace, Spec, Status interface{} }{}
		}
		ws = ws.DeepCopy()
		for i, nw := range ws.Spec.Network {
			if nw.TargetPortNumber != nil && *nw.TargetPortNumber != int32(nw.PortNumber) {
				ws.Spec.Network[i].TargetPortNumber = pointer.Int32(99999)
			}
		}
		return struct{ Name, Namespace, Spec, Status interface{} }{
			Name:      ws.Name,
			Namespace: ws.Namespace,
			Spec:      ws.Spec,
			Status:    ws.Status,
		}
	}

	proxySnap := func(proxies []proxy.LocalPortProxyInfo) []proxy.LocalPortProxyInfo {
		snapProxies := make([]proxy.LocalPortProxyInfo, 0, len(proxies))
		for _, p := range proxies {
			if p.LocalPort != 0 {
				p.LocalPort = 99999
			}
			snapProxies = append(snapProxies, p)
		}
		return snapProxies
	}

	//==================================================================================

	AfterEach(func() {
		clientMock.Clear()
		testUtil.DeleteWorkspaceAll()
		testUtil.DeleteCosmoUserAll()
		testUtil.DeleteTemplateAll()
	})

	//==================================================================================

	DescribeTable("✅ success in normal context:",

		func(initialFunc func(), modifyFuncs ...func()) {
			testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeWorkspace, "template1")
			testUtil.CreateLoginUser("test-user", "", nil, "password")
			initialFunc()
			By("---------------test start----------------")
			defer stopManager()
			startManager("test-user", "test-workspace")
			time.Sleep(time.Second * 3)

			for _, modifyFunc := range modifyFuncs {
				modifyFunc()
				time.Sleep(time.Second * 3)

				wsv1Workspace, _ := kosmoClient.GetWorkspaceByUserName(context.Background(), "test-workspace", "test-user")
				Ω(workspaceSnap(wsv1Workspace)).To(MatchSnapShot())

				proxies := proxyManager.GetRunningProxies()
				Ω(proxySnap(proxies)).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		},

		Entry("01 when starting a reconcile after creating network rules",
			func() {
				testUtil.CreateWorkspace("test-user", "test-workspace", "template1", nil)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw1", 1111, "gp1", "/", false)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw2", 2222, "gp2", "/", true)
			},
			func() {},
		),

		Entry("02 when starting a reconcile before creating network rules",
			func() {},
			func() {
				testUtil.CreateWorkspace("test-user", "test-workspace", "template1", nil)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw1", 1111, "gp1", "/", false)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw2", 2222, "gp2", "/", true)
			},
		),

		Entry("03 when changing the port Number",
			func() {
				testUtil.CreateWorkspace("test-user", "test-workspace", "template1", nil)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw1", 1111, "gp1", "/", false)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw2", 2222, "gp2", "/", true)
			},
			func() {
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw1", 3333, "gp1", "/", false)
			},
			func() {
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw2", 4444, "gp2", "/", true)
			},
		),

		Entry("04 when changing the path ",
			func() {
				testUtil.CreateWorkspace("test-user", "test-workspace", "template1", nil)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw1", 1111, "gp1", "/", false)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw2", 2222, "gp2", "/", true)
			},
			func() {
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw1", 1111, "gp1", "/abc", false)
			},
			func() {
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw2", 2222, "gp2", "/cdf", true)
			},
		),

		Entry("05 when changing the public flag",
			func() {
				testUtil.CreateWorkspace("test-user", "test-workspace", "template1", nil)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw1", 1111, "gp1", "/", false)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw2", 2222, "gp2", "/", true)
			},
			func() {
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw1", 1111, "gp1", "/", true)
			},
			func() {
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw2", 2222, "gp2", "/", false)
			},
		),

		Entry("06 wheh deleting the network rule",
			func() {
				testUtil.CreateWorkspace("test-user", "test-workspace", "template1", nil)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw1", 1111, "gp1", "/", false)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw2", 2222, "gp2", "/", true)
			},
			func() {
				testUtil.DeleteNetworkRule("test-user", "test-workspace", "nw2")
			},
			func() {
				testUtil.DeleteNetworkRule("test-user", "test-workspace", "nw1")
			},
		),

		Entry("07 when deleting the workspace",
			func() {},
			func() {
				testUtil.CreateWorkspace("test-user", "test-workspace", "template1", nil)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw1", 1111, "gp1", "/", false)
				testUtil.UpsertNetworkRule("test-user", "test-workspace", "nw2", 2222, "gp2", "/", true)
			},
			func() {
				testUtil.DeleteWorkspaceAll()
			},
		),
	)

})
