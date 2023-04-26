package dashboard

import (
	"context"
	"errors"
	"net/http"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"google.golang.org/protobuf/types/known/emptypb"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

var _ = Describe("Dashboard server [Template]", func() {

	var (
		userSession  string
		adminSession string
		client       dashboardv1alpha1connect.TemplateServiceClient
	)

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession("normal-user", "お名前", nil, "password")
		adminSession = test_CreateLoginUserSession("admin-user", "アドミン", []cosmov1alpha1.UserRole{cosmov1alpha1.PrivilegedRole}, "password")
		client = dashboardv1alpha1connect.NewTemplateServiceClient(http.DefaultClient, "http://localhost:8888")
	})

	AfterEach(func() {
		clientMock.Clear()
		testUtil.DeleteCosmoUserAll()
		testUtil.DeleteTemplateAll()
	})

	Describe("[GetWorkspaceTemplates]", func() {

		run_test := func(loginUser string, testCase string) {
			if testCase == "not empty" {
				testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeWorkspace, "template1")
				testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeWorkspace, "template2")
			}
			session := userSession
			if loginUser == "admin-user" {
				session = adminSession
			}
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.GetWorkspaceTemplates(ctx, NewRequestWithSession(&emptypb.Empty{}, session))
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
			Entry(nil, "admin-user", "empty"),
			Entry(nil, "admin-user", "not empty"),
			Entry(nil, "normal-user", "empty"),
			Entry(nil, "normal-user", "not empty"),
		)

		DescribeTable("❌ fail with an unexpected error at list:",
			func(user string, testCase string) {
				clientMock.SetListError((*Server).GetWorkspaceTemplates, errors.New("template list error"))
				run_test(user, testCase)
			},
			Entry(nil, "admin-user", "not empty"),
		)
	})

	Describe("[GetUserAddonTemplates]", func() {

		run_test := func(loginUser string, testCase string) {
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
			res, err := client.GetUserAddonTemplates(ctx, NewRequestWithSession(&emptypb.Empty{}, session))
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
			Entry(nil, "admin-user", "empty"),
			Entry(nil, "admin-user", "not empty"),
			Entry(nil, "normal-user", "empty"),
			Entry(nil, "normal-user", "not empty"),
		)

		DescribeTable("❌ fail with an unexpected error at list:",
			func(user string, testCase string) {
				clientMock.SetListError((*Server).GetUserAddonTemplates, errors.New("template list error"))
				run_test(user, testCase)
			},
			Entry(nil, "admin-user", "not empty"),
		)
	})
})
