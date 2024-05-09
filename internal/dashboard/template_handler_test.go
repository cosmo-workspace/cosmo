package dashboard

import (
	"context"
	"errors"
	"net/http"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	dashboardv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

var _ = Describe("Dashboard server [Template]", func() {

	var (
		userSession     string
		roleUserSession string
		adminSession    string
		client          dashboardv1alpha1connect.TemplateServiceClient
	)

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession("normal-user", "お名前", nil, "password")
		roleUserSession = test_CreateLoginUserSession("role-user", "お名前", []cosmov1alpha1.UserRole{{Name: "my-role"}}, "password")
		adminSession = test_CreateLoginUserSession("admin-user", "アドミン", []cosmov1alpha1.UserRole{cosmov1alpha1.PrivilegedRole}, "password")
		client = dashboardv1alpha1connect.NewTemplateServiceClient(http.DefaultClient, "http://localhost:8888")
	})

	AfterEach(func() {
		clientMock.Clear()
		testUtil.DeleteCosmoUserAll()
		testUtil.DeleteTemplateAll()
	})

	Describe("[GetWorkspaceTemplates]", func() {

		run_test := func(loginUser string, testCase string, req *dashboardv1alpha1.GetWorkspaceTemplatesRequest) {
			if testCase == "not empty" {
				testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeWorkspace, "template1")
				testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeWorkspace, "template2")
				testUtil.CreateTemplateForUserRole(cosmov1alpha1.TemplateLabelEnumTypeWorkspace, "template3", "my-role")
			}
			session := userSession
			if loginUser == "admin-user" {
				session = adminSession
			} else if loginUser == "role-user" {
				session = roleUserSession
			}
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.GetWorkspaceTemplates(ctx, NewRequestWithSession(req, session))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "admin-user", "empty", &dashboardv1alpha1.GetWorkspaceTemplatesRequest{}),
			Entry(nil, "admin-user", "not empty", &dashboardv1alpha1.GetWorkspaceTemplatesRequest{}),
			Entry(nil, "normal-user", "empty", &dashboardv1alpha1.GetWorkspaceTemplatesRequest{}),
			Entry(nil, "normal-user", "not empty", &dashboardv1alpha1.GetWorkspaceTemplatesRequest{}),
			Entry(nil, "normal-user", "not empty", &dashboardv1alpha1.GetWorkspaceTemplatesRequest{UseRoleFilter: ptr.To(true)}),
			Entry(nil, "role-user", "not empty", &dashboardv1alpha1.GetWorkspaceTemplatesRequest{}),
		)

		DescribeTable("❌ fail with an unexpected error at list:",
			func(user string, testCase string, req *dashboardv1alpha1.GetWorkspaceTemplatesRequest) {
				clientMock.SetListError((*Server).GetWorkspaceTemplates, errors.New("template list error"))
				run_test(user, testCase, req)
			},
			Entry(nil, "admin-user", "not empty", &dashboardv1alpha1.GetWorkspaceTemplatesRequest{}),
		)
	})

	Describe("[GetUserAddonTemplates]", func() {

		run_test := func(loginUser string, testCase string, req *dashboardv1alpha1.GetUserAddonTemplatesRequest) {
			if testCase == "not empty" {
				testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeUserAddon, "useraddon1")
				testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeUserAddon, "useraddon2")
			}
			session := userSession
			if loginUser == "admin-user" {
				session = adminSession
			}
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.GetUserAddonTemplates(ctx, NewRequestWithSession(req, session))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "admin-user", "empty", &dashboardv1alpha1.GetUserAddonTemplatesRequest{}),
			Entry(nil, "admin-user", "not empty", &dashboardv1alpha1.GetUserAddonTemplatesRequest{}),
			Entry(nil, "normal-user", "empty", &dashboardv1alpha1.GetUserAddonTemplatesRequest{}),
			Entry(nil, "normal-user", "not empty", &dashboardv1alpha1.GetUserAddonTemplatesRequest{}),
		)

		DescribeTable("❌ fail with an unexpected error at list:",
			func(user string, testCase string, req *dashboardv1alpha1.GetUserAddonTemplatesRequest) {
				clientMock.SetListError((*Server).GetUserAddonTemplates, errors.New("template list error"))
				run_test(user, testCase, req)
			},
			Entry(nil, "admin-user", "not empty", &dashboardv1alpha1.GetUserAddonTemplatesRequest{}),
		)
	})
})
