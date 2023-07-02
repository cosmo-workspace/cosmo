package controllers

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	traefikv1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo/test"
	"github.com/cosmo-workspace/cosmo/pkg/workspace"
	//+kubebuilder:scaffold:imports
)

const (
	controllerFieldManager string = "cosmo-instance-controller"
)

const (
	instController        string = "cosmo-instance-controller"
	clusterInstController string = "cosmo-cluster-instance-controller"
	tmplController        string = "cosmo-template-controller"
	clusterTmplController string = "cosmo-cluster-template-controller"
	userController        string = "cosmo-user-controller"
	wsController          string = "cosmo-workspace-controller"
	wsStatController      string = "cosmo-workspace-status-controller"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
	testUtil  test.TestUtil
)

func init() {
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme.Scheme))
	utilruntime.Must(traefikv1.AddToScheme(scheme.Scheme))
	//+kubebuilder:scaffold:scheme
}

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(ctrl.SetupSignalHandler())

	By("bootstrapping test environment")

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd", "bases"),
			filepath.Join("..", "..", "config", "crd", "traefik")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())

	testUtil = test.NewTestUtil(k8sClient)

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
	})
	Expect(err).NotTo(HaveOccurred())

	err = (&InstanceReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(instController),
	}).SetupWithManager(mgr, controllerFieldManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&ClusterInstanceReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(instController),
	}).SetupWithManager(mgr, controllerFieldManager)
	Expect(err).NotTo(HaveOccurred())

	err = (&TemplateReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&ClusterTemplateReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&WorkspaceReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(wsController),

		TraefikIngressRouteCfg: &workspace.TraefikIngressRouteConfig{
			Entrypoints: []string{"web", "websecure"},
			TLS:         nil,
			AuthenMiddleware: traefikv1.MiddlewareRef{
				Name:      "cosmo-auth",
				Namespace: "cosmo-system",
			},
			UserNameHeaderMiddleware: traefikv1.MiddlewareRef{
				Name: "userNameHeader",
			},
			HostBase: "{{NETRULE}}-{{WORKSPACE}}-{{USER}}",
			Domain:   "domain",
		},
		URLBaseProtocol: "https",
	}).SetupWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&WorkspaceStatusReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(wsStatController),
	}).SetupWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&UserReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(userController),
	}).SetupWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

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

func ownerRef(obj runtime.Object, scheme *runtime.Scheme) metav1.OwnerReference {
	type ownerObject interface {
		GetName() string
		GetUID() types.UID
	}

	owner, ok := obj.(ownerObject)
	Expect(ok).Should(BeTrue())

	gvk, err := apiutil.GVKForObject(obj, scheme)
	Expect(err).ShouldNot(HaveOccurred())
	return metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		BlockOwnerDeletion: pointer.Bool(true),
		Controller:         pointer.Bool(true),
	}
}
