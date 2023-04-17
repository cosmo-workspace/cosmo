package cmd

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/internal/webhooks"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient kosmo.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
)

const DefaultURLBase = "https://{{NETRULE_GROUP}}-{{INSTANCE}}-{{USER_NAME}}.domain"

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme.Scheme))
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme.Scheme))
	//+kubebuilder:scaffold:scheme
}

func TestCommandsl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cosmoctl cmd Suite")
}

var _ = BeforeSuite(func() {
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

	c, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())

	k8sClient = kosmo.NewClient(c)
	Expect(k8sClient).NotTo(BeNil())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
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

	(&webhooks.TemplateMutationWebhookHandler{
		Client:         k8sClient,
		Log:            clog.NewLogger(ctrl.Log.WithName("TemplateMutationWebhookHandler")),
		DefaultURLBase: DefaultURLBase,
	}).SetupWebhookWithManager(mgr)

	(&webhooks.TemplateValidationWebhookHandler{
		Client:       k8sClient,
		Log:          clog.NewLogger(ctrl.Log.WithName("TemplateValidationWebhookHandler")),
		FieldManager: "cosmo-instance-controller",
	}).SetupWebhookWithManager(mgr)

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

func test_CreateClusterTemplate(templateType string, templateName string) {
	ctx := context.Background()
	tmpl := cosmov1alpha1.ClusterTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: templateName,
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: templateType,
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RequiredVars: []cosmov1alpha1.RequiredVarSpec{
				{Var: "{{HOGE}}", Default: "HOGEhoge"},
				{Var: "{{FUGA}}", Default: "FUGAfuga"},
			},
		},
	}
	err := k8sClient.Create(ctx, &tmpl)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() error {
		err := k8sClient.Get(ctx, client.ObjectKey{Name: templateName}, &cosmov1alpha1.ClusterTemplate{})
		return err
	}, time.Second*5, time.Millisecond*100).Should(Succeed())
}

func test_CreateTemplate(templateType string, templateName string) {
	ctx := context.Background()
	tmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: templateName,
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: templateType,
			},
			Annotations: map[string]string{
				cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort:  "main",
				cosmov1alpha1.UserAddonTemplateAnnKeyDefaultUserAddon: "true",
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RequiredVars: []cosmov1alpha1.RequiredVarSpec{
				{Var: "{{HOGE}}", Default: "HOGEhoge"},
				{Var: "{{FUGA}}", Default: "FUGAfuga"},
			},
		},
	}
	err := k8sClient.Create(ctx, &tmpl)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() error {
		err := k8sClient.Get(ctx, client.ObjectKey{Name: templateName}, &cosmov1alpha1.Template{})
		return err
	}, time.Second*5, time.Millisecond*100).Should(Succeed())
}

func test_DeleteTemplateAll() {
	ctx := context.Background()
	err := k8sClient.DeleteAllOf(ctx, &cosmov1alpha1.Template{})
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() ([]cosmov1alpha1.Template, error) {
		var tmplList cosmov1alpha1.TemplateList
		err := k8sClient.List(ctx, &tmplList)
		return tmplList.Items, err
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func test_DeleteClusterTemplateAll() {
	ctx := context.Background()
	err := k8sClient.DeleteAllOf(ctx, &cosmov1alpha1.ClusterTemplate{})
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() ([]cosmov1alpha1.ClusterTemplate, error) {
		var l cosmov1alpha1.ClusterTemplateList
		err := k8sClient.List(ctx, &l)
		return l.Items, err
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func test_CreateCosmoUser(username string, dispayName string, role []cosmov1alpha1.UserRole) {
	ctx := context.Background()
	user := cosmov1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: username,
		},
		Spec: cosmov1alpha1.UserSpec{
			DisplayName: dispayName,
			Roles:       role,
			AuthType:    cosmov1alpha1.UserAuthTypePasswordSecert,
		},
	}
	err := k8sClient.Create(ctx, &user)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() error {
		_, err := k8sClient.GetUser(ctx, username)
		return err
	}, time.Second*5, time.Millisecond*100).Should(Succeed())
}

func test_DeleteCosmoUserAll() {
	ctx := context.Background()
	users, err := k8sClient.ListUsers(ctx)
	Expect(err).ShouldNot(HaveOccurred())
	for _, user := range users {
		k8sClient.Delete(ctx, &user)
	}
	Eventually(func() ([]cosmov1alpha1.User, error) {
		return k8sClient.ListUsers(ctx)
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func test_CreateUserNameSpaceandDefaultPasswordIfAbsent(username string) {
	ctx := context.Background()
	var ns v1.Namespace
	key := client.ObjectKey{Name: cosmov1alpha1.UserNamespace(username)}
	err := k8sClient.Get(ctx, key, &ns)
	if err != nil {
		ns = v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: cosmov1alpha1.UserNamespace(username),
			},
		}
		err = k8sClient.Create(ctx, &ns)
		Expect(err).ShouldNot(HaveOccurred())
	}
	// create default password
	err = k8sClient.ResetPassword(ctx, username)
	Expect(err).ShouldNot(HaveOccurred())
}

func test_CreateLoginUser(username, displayName string, role []cosmov1alpha1.UserRole, password string) {
	ctx := context.Background()

	test_CreateCosmoUser(username, displayName, role)
	test_CreateUserNameSpaceandDefaultPasswordIfAbsent(username)
	err := k8sClient.RegisterPassword(ctx, username, []byte(password))
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() error {
		_, _, err := k8sClient.VerifyPassword(ctx, username, []byte(password))
		return err
	}, time.Second*5, time.Millisecond*100).Should(Succeed())
}

func test_CreateWorkspace(username string, name string, template string, vars map[string]string) {
	ctx := context.Background()

	cfg, err := k8sClient.GetWorkspaceConfig(ctx, template)
	Expect(err).ShouldNot(HaveOccurred())

	ws := &cosmov1alpha1.Workspace{}
	ws.SetName(name)
	ws.SetNamespace(cosmov1alpha1.UserNamespace(username))
	ws.Spec = cosmov1alpha1.WorkspaceSpec{
		Template: cosmov1alpha1.TemplateRef{
			Name: template,
		},
		Replicas: pointer.Int64(1),
		Vars:     vars,
	}
	err = k8sClient.Create(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())

	ws.Status.Phase = "Pending"
	ws.Status.Config = cfg
	err = k8sClient.Status().Update(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() (*cosmov1alpha1.Workspace, error) {
		return k8sClient.GetWorkspaceByUserName(ctx, name, username)
	}, time.Second*5, time.Millisecond*100).ShouldNot(BeNil())
}

func test_StopWorkspace(username string, name string) {
	ctx := context.Background()
	ws, err := k8sClient.GetWorkspaceByUserName(ctx, name, username)
	Expect(err).ShouldNot(HaveOccurred())
	ws.Spec.Replicas = pointer.Int64(0)
	err = k8sClient.Update(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())
}

func test_DeleteWorkspaceAllByusername(username string) {
	ctx := context.Background()
	err := k8sClient.DeleteAllOf(ctx, &cosmov1alpha1.Workspace{}, client.InNamespace(cosmov1alpha1.UserNamespace(username)))
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() ([]cosmov1alpha1.Workspace, error) {
		return k8sClient.ListWorkspaces(ctx, cosmov1alpha1.UserNamespace(username))
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func test_DeleteWorkspaceAll() {
	ctx := context.Background()
	users, err := k8sClient.ListUsers(ctx)
	Expect(err).ShouldNot(HaveOccurred())
	for _, user := range users {
		test_DeleteWorkspaceAllByusername(user.Name)
	}
}

func test_createNetworkRule(username, workspaceName, networkRuleName string, portNumber int32, group, httpPath string) {
	ctx := context.Background()

	_, err := k8sClient.AddNetworkRule(ctx, workspaceName, username, networkRuleName, portNumber, &group, httpPath, false)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() bool {
		ws, _ := k8sClient.GetWorkspaceByUserName(ctx, workspaceName, username)
		for _, n := range ws.Spec.Network {
			if n.Name == networkRuleName {
				return true
			}
		}
		return false
	}, time.Second*5, time.Millisecond*100).Should(BeTrue())
}
