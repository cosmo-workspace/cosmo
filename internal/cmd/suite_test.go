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
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"

	//+kubebuilder:scaffold:imports

	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/wsnet"
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

func TestAPIs(t *testing.T) {
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
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func test_CreateTemplate(templateType string, templateName string) {
	ctx := context.Background()
	tmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: templateName,
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: templateType,
			},
			Annotations: map[string]string{
				wsv1alpha1.TemplateAnnKeyWorkspaceServiceMainPort: "main",
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
		return k8sClient.ListTemplates(ctx)
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func test_CreateCosmoUser(id string, dispayName string, role wsv1alpha1.UserRole) {
	ctx := context.Background()
	user := wsv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
		},
		Spec: wsv1alpha1.UserSpec{
			DisplayName: dispayName,
			Role:        role,
			AuthType:    wsv1alpha1.UserAuthTypeKosmoSecert,
		},
	}
	err := k8sClient.Create(ctx, &user)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() error {
		_, err := k8sClient.GetUser(ctx, id)
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
	Eventually(func() ([]wsv1alpha1.User, error) {
		return k8sClient.ListUsers(ctx)
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func test_CreateUserNameSpaceandDefaultPasswordIfAbsent(id string) {
	ctx := context.Background()
	var ns v1.Namespace
	key := client.ObjectKey{Name: wsv1alpha1.UserNamespace(id)}
	err := k8sClient.Get(ctx, key, &ns)
	if err != nil {
		ns = v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: wsv1alpha1.UserNamespace(id),
			},
		}
		err = k8sClient.Create(ctx, &ns)
		Expect(err).ShouldNot(HaveOccurred())
	}
	// create default password
	err = k8sClient.ResetPassword(ctx, id)
	Expect(err).ShouldNot(HaveOccurred())
}

func test_CreateLoginUser(id, displayName string, role wsv1alpha1.UserRole, password string) {
	ctx := context.Background()

	test_CreateCosmoUser(id, displayName, role)
	test_CreateUserNameSpaceandDefaultPasswordIfAbsent(id)
	err := k8sClient.RegisterPassword(ctx, id, []byte(password))
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() error {
		_, _, err := k8sClient.VerifyPassword(ctx, id, []byte(password))
		return err
	}, time.Second*5, time.Millisecond*100).Should(Succeed())
}

func test_CreateWorkspace(userId string, name string, template string, vars map[string]string) {
	ctx := context.Background()

	ws := &wsv1alpha1.Workspace{}
	ws.SetName(name)
	ws.SetNamespace(wsv1alpha1.UserNamespace(userId))
	ws.Spec = wsv1alpha1.WorkspaceSpec{
		Template: cosmov1alpha1.TemplateRef{
			Name: template,
		},
		Replicas: pointer.Int64(1),
		Vars:     vars,
	}

	err := k8sClient.Create(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())

	ws.Status = wsv1alpha1.WorkspaceStatus{
		Instance: cosmov1alpha1.ObjectRef{},
		Phase:    "",
		URLs:     map[string]string{},
		Config: wsv1alpha1.Config{
			DeploymentName:      "",
			ServiceName:         "",
			IngressName:         "",
			ServiceMainPortName: "main",
			URLBase:             "",
		},
	}
	err = k8sClient.Status().Update(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() (*wsv1alpha1.Workspace, error) {
		return k8sClient.GetWorkspaceByUserID(ctx, name, userId)
	}, time.Second*5, time.Millisecond*100).ShouldNot(BeNil())
}

func test_StopWorkspace(userId string, name string) {
	ctx := context.Background()
	ws, err := k8sClient.GetWorkspaceByUserID(ctx, name, userId)
	Expect(err).ShouldNot(HaveOccurred())
	ws.Spec.Replicas = pointer.Int64(0)
	err = k8sClient.Update(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())
}

func test_DeleteWorkspaceAllByUserId(userId string) {
	ctx := context.Background()
	err := k8sClient.DeleteAllOf(ctx, &wsv1alpha1.Workspace{}, client.InNamespace(wsv1alpha1.UserNamespace(userId)))
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() ([]wsv1alpha1.Workspace, error) {
		return k8sClient.ListWorkspaces(ctx, wsv1alpha1.UserNamespace(userId))
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func test_DeleteWorkspaceAll() {
	ctx := context.Background()
	users, err := k8sClient.ListUsers(ctx)
	Expect(err).ShouldNot(HaveOccurred())
	for _, user := range users {
		test_DeleteWorkspaceAllByUserId(user.Name)
	}
}

func test_createNetworkRule(userId, workspaceName, networkRuleName string, portNumber int, group, httpPath string) {
	ctx := context.Background()

	ws, err := k8sClient.GetWorkspaceByUserID(ctx, workspaceName, userId)
	Expect(err).ShouldNot(HaveOccurred())

	netRule := wsv1alpha1.NetworkRule{
		PortName:   networkRuleName,
		PortNumber: portNumber,
		Group:      pointer.String(group),
		HTTPPath:   httpPath,
	}

	ws.Spec.Network, err = wsnet.UpsertNetRule(ws.Spec.Network, netRule)
	Expect(err).ShouldNot(HaveOccurred())

	err = k8sClient.Update(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() []wsv1alpha1.NetworkRule {
		ws, _ := k8sClient.GetWorkspaceByUserID(ctx, workspaceName, userId)
		return ws.Spec.Network
	}, time.Second*5, time.Millisecond*100).Should(ContainElement(netRule))
}
