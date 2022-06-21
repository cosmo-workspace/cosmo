package dashboard

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
)

var _ = Describe("Dashboard server [Template]", func() {

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession("usertest", "お名前", "", "password")
		adminSession = test_CreateLoginUserSession("usertest-admin", "アドミン", wsv1alpha1.UserAdminRole, "password")
	})

	AfterEach(func() {
		clientMock.Clear()
		test_DeleteCosmoUserAll()
		test_DeleteTemplateAll()
	})

	Describe("[GetWorkspaceTemplates]", func() {

		run_test := func(context string) {
			if context == "not empty" {
				test_CreateTemplate(wsv1alpha1.TemplateTypeWorkspace, "template1")
				test_CreateTemplate(wsv1alpha1.TemplateTypeWorkspace, "template2")
			}
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodGet, path: "/api/v1alpha1/template/workspace"})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "empty"),
			Entry(nil, "not empty"),
		)

		DescribeTable("❌ fail with an unexpected error at list:",
			func(context string) {
				clientMock.SetListError((*Server).GetWorkspaceTemplates, errors.New("template list error"))
				run_test(context)
			},
			Entry(nil, "not empty"),
		)
	})

	Describe("[GetUserAddonTemplates]", func() {

		run_test := func(context string) {
			if context == "not empty" {
				test_CreateTemplate(wsv1alpha1.TemplateTypeUserAddon, "useraddon1")
				test_CreateTemplate(wsv1alpha1.TemplateTypeUserAddon, "useraddon2")
			}
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodGet, path: "/api/v1alpha1/template/useraddon"})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "empty"),
			Entry(nil, "not empty"),
		)

		DescribeTable("❌ fail with an unexpected error at list:",
			func(context string) {
				clientMock.SetListError((*Server).GetUserAddonTemplates, errors.New("template list error"))
				run_test(context)
			},
			Entry(nil, "not empty"),
		)
	})
})
