package dashboard

import (
	"net/http"

	. "github.com/onsi/ginkgo"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
)

var _ = Describe("Dashboard server [Template]", func() {

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession("usertest", "お名前", "", "password")
		adminSession = test_CreateLoginUserSession("usertest-admin", "アドミン", wsv1alpha1.UserAdminRole, "password")
	})

	AfterEach(func() {
		test_DeleteCosmoUserAll()
	})

	When("list workspace templates", func() {

		AfterEach(func() {
			test_DeleteTemplateAll()
		})

		When("template is empty", func() {
			It("should return empty item", func() {
				test_HttpSendAndVerify(userSession,
					request{method: http.MethodGet, path: "/api/v1alpha1/template/workspace"},
					response{statusCode: http.StatusOK, body: `{ "message": "No items found", "items": []}`},
				)
			})
		})

		When("template is not empty", func() {
			It("should return items", func() {
				test_CreateTemplate(wsv1alpha1.TemplateTypeWorkspace, "template1")
				test_CreateTemplate(wsv1alpha1.TemplateTypeWorkspace, "template2")

				test_HttpSendAndVerify(userSession,
					request{method: http.MethodGet, path: "/api/v1alpha1/template/workspace"},
					response{
						statusCode: http.StatusOK,
						body: `{ "items": [` +
							`{"name": "template1", "requiredVars": [ { "varName": "{{HOGE}}", "defaultValue": "FUGA"}]},` +
							`{"name": "template2", "requiredVars": [ { "varName": "{{HOGE}}", "defaultValue": "FUGA"}]}` +
							`]}`,
					},
				)
			})
		})
	})

	When("list useraddon templates", func() {
		When("template is empty", func() {
			It("should return empty item", func() {
				test_HttpSendAndVerify(userSession,
					request{method: http.MethodGet, path: "/api/v1alpha1/template/useraddon"},
					response{statusCode: http.StatusOK, body: `{ "message": "No items found", "items": []}`},
				)
			})
		})

		When("template is not empty", func() {
			It("should return items", func() {
				test_CreateTemplate(wsv1alpha1.TemplateTypeUserAddon, "useraddon1")
				test_CreateTemplate(wsv1alpha1.TemplateTypeUserAddon, "useraddon2")

				test_HttpSendAndVerify(userSession,
					request{method: http.MethodGet, path: "/api/v1alpha1/template/useraddon"},
					response{
						statusCode: 200,
						body: `{ "items": [` +
							`{"name": "useraddon1", "requiredVars": [ { "varName": "{{HOGE}}", "defaultValue": "FUGA"}]},` +
							`{"name": "useraddon2", "requiredVars": [ { "varName": "{{HOGE}}", "defaultValue": "FUGA"}]}` +
							`]}`,
					},
				)
			})
		})
	})
})
