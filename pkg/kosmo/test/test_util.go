package test

import (
	"context"
	"time"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
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
				cosmov1alpha1.WorkspaceTemplateAnnKeyServiceName:      "workspace",
				cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort:  "main",
				cosmov1alpha1.UserAddonTemplateAnnKeyDefaultUserAddon: "true",
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RequiredVars: []cosmov1alpha1.RequiredVarSpec{
				{Var: "{{HOGE}}", Default: "HOGEhoge"},
				{Var: "{{FUGA}}", Default: "FUGAfuga"},
			},
			RawYaml: `---
apiVersion: v1
kind: Service
metadata:
  name: workspace
spec:
  ports:
  - name: main
    port: 18080
`,
		},
	}
	err := c.kosmoClient.Create(ctx, &tmpl)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() error {
		err := c.kosmoClient.Get(ctx, client.ObjectKey{Name: templateName}, &cosmov1alpha1.Template{})
		return err
	}, time.Second*5, time.Millisecond*100).Should(Succeed())
}

func (c *TestUtil) CreateClusterTemplate(templateType string, templateName string) {
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
	err := c.kosmoClient.Create(ctx, &tmpl)
	Expect(err).ShouldNot(HaveOccurred())

	Eventually(func() error {
		err := c.kosmoClient.Get(ctx, client.ObjectKey{Name: templateName}, &cosmov1alpha1.ClusterTemplate{})
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

func (c *TestUtil) DeleteClusterTemplateAll() {
	ctx := context.Background()
	err := c.kosmoClient.DeleteAllOf(ctx, &cosmov1alpha1.ClusterTemplate{})
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() ([]cosmov1alpha1.ClusterTemplate, error) {
		var l cosmov1alpha1.ClusterTemplateList
		err := c.kosmoClient.List(ctx, &l)
		return l.Items, err
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func (c *TestUtil) CreateCosmoUser(userName string, dispayName string, role []cosmov1alpha1.UserRole, authType cosmov1alpha1.UserAuthType) {
	ctx := context.Background()
	user := cosmov1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: userName,
		},
		Spec: cosmov1alpha1.UserSpec{
			DisplayName: dispayName,
			Roles:       role,
			AuthType:    authType,
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
	Eventually(func() ([]cosmov1alpha1.User, error) {
		return c.kosmoClient.ListUsers(ctx)
	}, time.Second*5, time.Millisecond*100).Should(BeEmpty())
}

func (c *TestUtil) CreateUserNameSpaceandDefaultPasswordIfAbsent(userName string) {
	ctx := context.Background()
	var ns v1.Namespace
	key := client.ObjectKey{Name: cosmov1alpha1.UserNamespace(userName)}
	err := c.kosmoClient.Get(ctx, key, &ns)
	if err != nil {
		ns = v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: cosmov1alpha1.UserNamespace(userName),
			},
		}
		err = c.kosmoClient.Create(ctx, &ns)
		Expect(err).ShouldNot(HaveOccurred())
	}
	// create default password
	err = c.kosmoClient.ResetPassword(ctx, userName)
	Expect(err).ShouldNot(HaveOccurred())
}

func (c *TestUtil) CreateLoginUser(userName, displayName string, role []cosmov1alpha1.UserRole, authType cosmov1alpha1.UserAuthType, password string) {
	ctx := context.Background()

	c.CreateCosmoUser(userName, displayName, role, authType)
	c.CreateUserNameSpaceandDefaultPasswordIfAbsent(userName)

	if authType == cosmov1alpha1.UserAuthTypePasswordSecert {
		err := c.kosmoClient.RegisterPassword(ctx, userName, []byte(password))
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(func() error {
			_, _, err := c.kosmoClient.VerifyPassword(ctx, userName, []byte(password))
			return err
		}, time.Second*5, time.Millisecond*100).Should(Succeed())
	}
}

func (c *TestUtil) CreateWorkspace(userName string, name string, template string, vars map[string]string) {
	ctx := context.Background()

	cfg, err := c.kosmoClient.GetWorkspaceConfig(ctx, template)
	Expect(err).ShouldNot(HaveOccurred())

	ws := &cosmov1alpha1.Workspace{}
	ws.SetName(name)
	ws.SetNamespace(cosmov1alpha1.UserNamespace(userName))
	ws.Spec = cosmov1alpha1.WorkspaceSpec{
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
	Eventually(func() (*cosmov1alpha1.Workspace, error) {
		return c.kosmoClient.GetWorkspaceByUserName(ctx, name, userName)
	}, time.Second*5, time.Millisecond*100).ShouldNot(BeNil())
}

func (c *TestUtil) StopWorkspace(userName string, name string) {
	ctx := context.Background()
	ws, err := c.kosmoClient.GetWorkspaceByUserName(ctx, name, userName)
	Expect(err).ShouldNot(HaveOccurred())
	ws.Spec.Replicas = pointer.Int64(0)
	err = c.kosmoClient.Update(ctx, ws)
	Expect(err).ShouldNot(HaveOccurred())
}

func (c *TestUtil) DeleteWorkspaceAllByUserName(userName string) {
	ctx := context.Background()
	err := c.kosmoClient.DeleteAllOf(ctx, &cosmov1alpha1.Workspace{}, client.InNamespace(cosmov1alpha1.UserNamespace(userName)))
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(func() ([]cosmov1alpha1.Workspace, error) {
		return c.kosmoClient.ListWorkspaces(ctx, cosmov1alpha1.UserNamespace(userName))
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

func (c *TestUtil) UpsertNetworkRule(userName, workspaceName string, customHostPrefix string, portNumber int32, httpPath string, public bool, index int) {
	ctx := context.Background()

	r := cosmov1alpha1.NetworkRule{
		PortNumber:       portNumber,
		CustomHostPrefix: customHostPrefix,
		HTTPPath:         httpPath,
		Public:           public,
	}
	_, err := c.kosmoClient.AddNetworkRule(ctx, workspaceName, userName, r, index)
	Expect(err).ShouldNot(HaveOccurred())
}

func (c *TestUtil) DeleteNetworkRule(userName, workspaceName string, index int) {
	ctx := context.Background()

	_, err := c.kosmoClient.DeleteNetworkRule(ctx, workspaceName, userName, index)
	Expect(err).ShouldNot(HaveOccurred())
}
