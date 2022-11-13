package test

import (
	"context"
	"time"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

type TestUtil struct {
	kosmoClient kosmo.Client
}

func NewTestUtil(client client.Client) TestUtil {
	k := kosmo.NewClient(client)
	Expect(k).NotTo(BeNil())
	return TestUtil{kosmoClient: k}
}

func (c *TestUtil) CreateTemplate(templateType string, templateName string) {
	ctx := context.Background()
	tmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: templateName,
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: templateType,
			},
			Annotations: map[string]string{
				wsv1alpha1.TemplateAnnKeyWorkspaceServiceMainPort: "main",
				wsv1alpha1.TemplateAnnKeyDefaultUserAddon:         "true",
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RequiredVars: []cosmov1alpha1.RequiredVarSpec{
				{Var: "{{HOGE}}", Default: "HOGEhoge"},
				{Var: "{{FUGA}}", Default: "FUGAfuga"},
			},
		},
	}
	err := c.kosmoClient.Create(ctx, &tmpl)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() error {
		err := c.kosmoClient.Get(ctx, client.ObjectKey{Name: templateName}, &cosmov1alpha1.Template{})
		return err
	}, time.Second*5, time.Millisecond*100).Should(Succeed())
}

func (c *TestUtil) DeleteTemplateAll() {
	ctx := context.Background()
	err := c.kosmoClient.DeleteAllOf(ctx, &cosmov1alpha1.Template{})
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() ([]cosmov1alpha1.Template, error) {
		var tmplList cosmov1alpha1.TemplateList
		err := c.kosmoClient.List(ctx, &tmplList)
		return tmplList.Items, err
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func (c *TestUtil) CreateCosmoUser(userName string, dispayName string, role wsv1alpha1.UserRole) {
	ctx := context.Background()
	user := wsv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: userName,
		},
		Spec: wsv1alpha1.UserSpec{
			DisplayName: dispayName,
			Role:        role,
			AuthType:    wsv1alpha1.UserAuthTypePasswordSecert,
		},
	}
	err := c.kosmoClient.Create(ctx, &user)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() error {
		_, err := c.kosmoClient.GetUser(ctx, userName)
		return err
	}, time.Second*5, time.Millisecond*100).Should(Succeed())
}

func (c *TestUtil) DeleteCosmoUserAll() {
	ctx := context.Background()
	users, err := c.kosmoClient.ListUsers(ctx)
	Expect(err).ShouldNot(HaveOccurred())
	for _, user := range users {
		c.kosmoClient.Delete(ctx, &user)
	}
	Eventually(func() ([]wsv1alpha1.User, error) {
		return c.kosmoClient.ListUsers(ctx)
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func (c *TestUtil) CreateUserNameSpaceandDefaultPasswordIfAbsent(userName string) {
	ctx := context.Background()
	var ns v1.Namespace
	key := client.ObjectKey{Name: wsv1alpha1.UserNamespace(userName)}
	err := c.kosmoClient.Get(ctx, key, &ns)
	if err != nil {
		ns = v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: wsv1alpha1.UserNamespace(userName),
			},
		}
		err = c.kosmoClient.Create(ctx, &ns)
		Expect(err).ShouldNot(HaveOccurred())
	}
	// create default password
	err = c.kosmoClient.ResetPassword(ctx, userName)
	Expect(err).ShouldNot(HaveOccurred())
}

func (c *TestUtil) CreateLoginUser(userName, displayName string, role wsv1alpha1.UserRole, password string) {
	ctx := context.Background()

	c.CreateCosmoUser(userName, displayName, role)
	c.CreateUserNameSpaceandDefaultPasswordIfAbsent(userName)
	err := c.kosmoClient.RegisterPassword(ctx, userName, []byte(password))
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() error {
		_, _, err := c.kosmoClient.VerifyPassword(ctx, userName, []byte(password))
		return err
	}, time.Second*5, time.Millisecond*100).Should(Succeed())
}

func (c *TestUtil) CreateWorkspace(userName string, name string, template string, vars map[string]string) {
	ctx := context.Background()

	cfg, err := c.kosmoClient.GetWorkspaceConfig(ctx, template)
	Expect(err).ShouldNot(HaveOccurred())

	ws := &wsv1alpha1.Workspace{}
	ws.SetName(name)
	ws.SetNamespace(wsv1alpha1.UserNamespace(userName))
	ws.Spec = wsv1alpha1.WorkspaceSpec{
		Template: cosmov1alpha1.TemplateRef{
			Name: template,
		},
		Replicas: pointer.Int64(1),
		Vars:     vars,
	}
	err = c.kosmoClient.Create(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())

	ws.Status.Phase = "Pending"
	ws.Status.Config = cfg
	err = c.kosmoClient.Status().Update(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() (*wsv1alpha1.Workspace, error) {
		return c.kosmoClient.GetWorkspaceByUserID(ctx, name, userName)
	}, time.Second*5, time.Millisecond*100).ShouldNot(BeNil())
}

func (c *TestUtil) StopWorkspace(userName string, name string) {
	ctx := context.Background()
	ws, err := c.kosmoClient.GetWorkspaceByUserID(ctx, name, userName)
	Expect(err).ShouldNot(HaveOccurred())
	ws.Spec.Replicas = pointer.Int64(0)
	err = c.kosmoClient.Update(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())
}

func (c *TestUtil) DeleteWorkspaceAllByUserName(userName string) {
	ctx := context.Background()
	err := c.kosmoClient.DeleteAllOf(ctx, &wsv1alpha1.Workspace{}, client.InNamespace(wsv1alpha1.UserNamespace(userName)))
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() ([]wsv1alpha1.Workspace, error) {
		return c.kosmoClient.ListWorkspaces(ctx, wsv1alpha1.UserNamespace(userName))
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func (c *TestUtil) DeleteWorkspaceAll() {
	ctx := context.Background()
	users, err := c.kosmoClient.ListUsers(ctx)
	Expect(err).ShouldNot(HaveOccurred())
	for _, user := range users {
		c.DeleteWorkspaceAllByUserName(user.Name)
	}
}

func (c *TestUtil) UpsertNetworkRule(userName, workspaceName, networkRuleName string, portNumber int, group, httpPath string, public bool) {
	ctx := context.Background()

	_, err := c.kosmoClient.AddNetworkRule(ctx, workspaceName, userName, networkRuleName, portNumber, &group, httpPath, public)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() bool {
		ws, _ := c.kosmoClient.GetWorkspaceByUserID(ctx, workspaceName, userName)
		for _, n := range ws.Spec.Network {
			if n.Name == networkRuleName {
				return true
			}
		}
		return false
	}, time.Second*5, time.Millisecond*100).Should(BeTrue())
}

func (c *TestUtil) DeleteNetworkRule(userName, workspaceName, networkRuleName string) {
	ctx := context.Background()

	_, err := c.kosmoClient.DeleteNetworkRule(ctx, workspaceName, userName, networkRuleName)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() bool {
		ws, _ := c.kosmoClient.GetWorkspaceByUserID(ctx, workspaceName, userName)
		for _, n := range ws.Spec.Network {
			if n.Name == networkRuleName {
				return true
			}
		}
		return false
	}, time.Second*5, time.Millisecond*100).Should(BeFalse())
}
