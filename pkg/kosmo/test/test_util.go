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

func (c *TestUtil) CreateCosmoUser(id string, dispayName string, role wsv1alpha1.UserRole) {
	ctx := context.Background()
	user := wsv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
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
		_, err := c.kosmoClient.GetUser(ctx, id)
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

func (c *TestUtil) CreateUserNameSpaceandDefaultPasswordIfAbsent(id string) {
	ctx := context.Background()
	var ns v1.Namespace
	key := client.ObjectKey{Name: wsv1alpha1.UserNamespace(id)}
	err := c.kosmoClient.Get(ctx, key, &ns)
	if err != nil {
		ns = v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: wsv1alpha1.UserNamespace(id),
			},
		}
		err = c.kosmoClient.Create(ctx, &ns)
		Expect(err).ShouldNot(HaveOccurred())
	}
	// create default password
	err = c.kosmoClient.ResetPassword(ctx, id)
	Expect(err).ShouldNot(HaveOccurred())
}

func (c *TestUtil) CreateLoginUser(id, displayName string, role wsv1alpha1.UserRole, password string) {
	ctx := context.Background()

	c.CreateCosmoUser(id, displayName, role)
	c.CreateUserNameSpaceandDefaultPasswordIfAbsent(id)
	err := c.kosmoClient.RegisterPassword(ctx, id, []byte(password))
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() error {
		_, _, err := c.kosmoClient.VerifyPassword(ctx, id, []byte(password))
		return err
	}, time.Second*5, time.Millisecond*100).Should(Succeed())
}

func (c *TestUtil) CreateWorkspace(userId string, name string, template string, vars map[string]string) {
	ctx := context.Background()

	cfg, err := c.kosmoClient.GetWorkspaceConfig(ctx, template)
	Expect(err).ShouldNot(HaveOccurred())

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
	err = c.kosmoClient.Create(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())

	ws.Status.Phase = "Pending"
	ws.Status.Config = cfg
	err = c.kosmoClient.Status().Update(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() (*wsv1alpha1.Workspace, error) {
		return c.kosmoClient.GetWorkspaceByUserID(ctx, name, userId)
	}, time.Second*5, time.Millisecond*100).ShouldNot(BeNil())
}

func (c *TestUtil) StopWorkspace(userId string, name string) {
	ctx := context.Background()
	ws, err := c.kosmoClient.GetWorkspaceByUserID(ctx, name, userId)
	Expect(err).ShouldNot(HaveOccurred())
	ws.Spec.Replicas = pointer.Int64(0)
	err = c.kosmoClient.Update(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())
}

func (c *TestUtil) DeleteWorkspaceAllByUserId(userId string) {
	ctx := context.Background()
	err := c.kosmoClient.DeleteAllOf(ctx, &wsv1alpha1.Workspace{}, client.InNamespace(wsv1alpha1.UserNamespace(userId)))
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() ([]wsv1alpha1.Workspace, error) {
		return c.kosmoClient.ListWorkspaces(ctx, wsv1alpha1.UserNamespace(userId))
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func (c *TestUtil) DeleteWorkspaceAll() {
	ctx := context.Background()
	users, err := c.kosmoClient.ListUsers(ctx)
	Expect(err).ShouldNot(HaveOccurred())
	for _, user := range users {
		c.DeleteWorkspaceAllByUserId(user.Name)
	}
}

func (c *TestUtil) UpsertNetworkRule(userId, workspaceName, networkRuleName string, portNumber int, group, httpPath string, public bool) {
	ctx := context.Background()

	_, err := c.kosmoClient.AddNetworkRule(ctx, workspaceName, userId, networkRuleName, portNumber, &group, httpPath, public)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() bool {
		ws, _ := c.kosmoClient.GetWorkspaceByUserID(ctx, workspaceName, userId)
		for _, n := range ws.Spec.Network {
			if n.PortName == networkRuleName {
				return true
			}
		}
		return false
	}, time.Second*5, time.Millisecond*100).Should(BeTrue())
}

func (c *TestUtil) DeleteNetworkRule(userId, workspaceName, networkRuleName string) {
	ctx := context.Background()

	_, err := c.kosmoClient.DeleteNetworkRule(ctx, workspaceName, userId, networkRuleName)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() bool {
		ws, _ := c.kosmoClient.GetWorkspaceByUserID(ctx, workspaceName, userId)
		for _, n := range ws.Spec.Network {
			if n.PortName == networkRuleName {
				return true
			}
		}
		return false
	}, time.Second*5, time.Millisecond*100).Should(BeFalse())
}
