package dashboard

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
)

var _ = Describe("Dashboard server [Workspace]", func() {

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession("usertest", "user", "", "password")
		adminSession = test_CreateLoginUserSession("usertest-admin", "admin", wsv1alpha1.UserAdminRole, "password")
		test_CreateTemplate(wsv1alpha1.TemplateTypeWorkspace, "template1")
	})

	AfterEach(func() {
		test_DeleteWorkspaceAll()
		test_DeleteCosmoUserAll()
		test_DeleteTemplateAll()
	})

	Describe("authorization by role", func() {

		var session []*http.Cookie

		deny403 := func(whenText string, request request) {
			When(whenText, func() {
				It("should deny with 403 Forbidden", func() {
					test_HttpSendAndVerify(session, request, response{statusCode: http.StatusForbidden, body: ""})
				})
			})
		}
		ok200 := func(whenText string, request request) {
			When(whenText, func() {
				It("should succeed with 200 ok", func() {
					test_HttpSendAndVerify(session, request, response{statusCode: http.StatusOK, body: "@ignore"})
				})
			})
		}
		ok201 := func(whenText string, request request) {
			When(whenText, func() {
				It("should succeed with 201 created", func() {
					test_HttpSendAndVerify(session, request, response{statusCode: http.StatusCreated, body: "@ignore"})
				})
			})
		}

		When("access API with normal user session", func() {

			BeforeEach(func() {
				session = userSession
				test_CreateWorkspace("usertest", "ws1", "template1", map[string]string{})
				test_createNetworkRule("usertest", "ws1", "nw1", 9999, "gp1", "/")
				test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})
				test_createNetworkRule("usertest-admin", "ws1", "nw1", 9999, "gp1", "/")
			})

			When("update own resource", func() {

				ok201("create a new workspace", request{method: http.MethodPost, path: "/api/v1alpha1/user/usertest/workspace", body: `{"name": "ws2","template": "template1"}`})
				ok200("Get all workspace of user", request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest/workspace"})
				ok200("get workspace by name", request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest/workspace/ws1"})
				ok200("Delete workspace", request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest/workspace/ws1"})
				ok200("Update workspace", request{method: http.MethodPatch, path: "/api/v1alpha1/user/usertest/workspace/ws1", body: `{"replicas": 1}`})
				ok200("Upsert workspace network rule", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/workspace/ws1/network/nw2", body: `{"portNumber": 3000,"group": "gp2","httpPath": "/"}`})
				ok200("Remove workspace network rule", request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest/workspace/ws1/network/nw1"})
			})

			When("update resource of others", func() {
				deny403("create a new workspace", request{method: http.MethodPost, path: "/api/v1alpha1/user/usertest-admin/workspace", body: `{"name": "ws2","template": "template1"}`})
				deny403("Get all workspace of user", request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin/workspace"})
				deny403("get workspace by name", request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1"})
				deny403("Delete workspace", request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1"})
				deny403("Update workspace", request{method: http.MethodPatch, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1", body: `{"replicas": 1}`})
				deny403("Upsert workspace network rule", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw2", body: `{"portNumber": 3000,"group": "gp2","httpPath": "/"}`})
				deny403("Remove workspace network rule", request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw1"})
			})

		})

		When("access API with admin user session", func() {

			BeforeEach(func() {
				session = adminSession
				test_CreateWorkspace("usertest", "ws1", "template1", map[string]string{})
				test_createNetworkRule("usertest", "ws1", "nw1", 9999, "gp1", "/")
				test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})
				test_createNetworkRule("usertest-admin", "ws1", "nw1", 9999, "gp1", "/")
			})

			When("update own resource", func() {
				ok201("create a new workspace", request{method: http.MethodPost, path: "/api/v1alpha1/user/usertest-admin/workspace", body: `{"name": "ws2","template": "template1"}`})
				ok200("Get all workspace of user", request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin/workspace"})
				ok200("get workspace by name", request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1"})
				ok200("Delete workspace", request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1"})
				ok200("Update workspace", request{method: http.MethodPatch, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1", body: `{"replicas": 1}`})
				ok200("Upsert workspace network rule", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw2", body: `{"portNumber": 3000,"group": "gp2","httpPath": "/"}`})
				ok200("Remove workspace network rule", request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw1"})
			})

			When("update resource of others", func() {
				ok201("create a new workspace", request{method: http.MethodPost, path: "/api/v1alpha1/user/usertest/workspace", body: `{"name": "ws2","template": "template1"}`})
				ok200("Get all workspace of user", request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest/workspace"})
				ok200("get workspace by name", request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest/workspace/ws1"})
				ok200("Delete workspace", request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest/workspace/ws1"})
				ok200("Update workspace", request{method: http.MethodPatch, path: "/api/v1alpha1/user/usertest/workspace/ws1", body: `{"replicas": 1}`})
				ok200("Upsert workspace network rule", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/workspace/ws1/network/nw2", body: `{"portNumber": 3000,"group": "gp2","httpPath": "/"}`})
				ok200("Remove workspace network rule", request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest/workspace/ws1/network/nw1"})
			})

		})

	})

	When("create a new Workspace", func() {

		Describe("invalid request error", func() {

			When("user is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPost, path: "/api/v1alpha1/user/xxxxx/workspace", body: `{"name": "","template": "template1"}`,
						},
						response{statusCode: http.StatusNotFound, body: `{"message": "user is not found"}`},
					)
				})
			})

			When("name is empty", func() {
				It("should deny with 400 BadRequest", func() {
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPost, path: "/api/v1alpha1/user/usertest-admin/workspace", body: `{"name": "","template": "template1"}`,
						},
						response{statusCode: http.StatusBadRequest, body: `{"message":"required field 'name' is zero value."}`},
					)
				})
			})

			When("template is empty", func() {
				It("should deny with 400 BadRequest", func() {
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPost, path: "/api/v1alpha1/user/usertest-admin/workspace", body: `{"name": "ws1","template": ""}`,
						},
						response{statusCode: http.StatusBadRequest, body: `{"message":"required field 'template' is zero value."}`},
					)
				})
			})

			When("name is invalid (include upper case)", func() {
				It("should deny with 500 InternalServerError", func() {
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPost, path: "/api/v1alpha1/user/usertest-admin/workspace", body: `{"name": "XXXX","template": "template1"}`,
						},
						response{statusCode: http.StatusInternalServerError, body: `{"message": "failed to create workspace"}`},
					)
				})
			})

			When("failed to get workspace config in template", func() {
				It("should deny with 400 BadRequest", func() {
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPost, path: "/api/v1alpha1/user/usertest-admin/workspace", body: `{"name": "ws1","template": "XXX"}`,
						},
						response{statusCode: http.StatusBadRequest, body: `{"message": "failed to get workspace config in template"}`},
					)
				})
			})

			When("workspace is already exists", func() {
				It("should deny witn 429 TooManyRequests", func() {
					test_CreateWorkspace("usertest-admin", "ws1", "template1", nil)

					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPost, path: "/api/v1alpha1/user/usertest-admin/workspace", body: `{"name": "ws1","template": "template1"}`,
						},
						response{statusCode: http.StatusTooManyRequests, body: `{"message": "Workspace already exists"}`},
					)
				})
			})
		})

		Describe("valid request", func() {

			When("all valid arguments are specified", func() {
				It("should be registered as specified", func() {

					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPost, path: "/api/v1alpha1/user/usertest-admin/workspace",
							body: `{"name": "ws1","template": "template1","vars": { "HOGE": "HOGEHOGE"}}`,
						},
						response{
							statusCode: http.StatusCreated,
							body: `{
								"message": "Successfully created",
								"workspace": {
								  "name": "ws1",
								  "ownerID": "usertest-admin",
								  "spec": {
									"template": "template1",
									"replicas": 0,
									"vars": {
									  "HOGE": "HOGEHOGE"
									}
								  },
								  "status": {
									"phase": "Pending"
								  }
								}
							  }`,
						},
					)
					wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), "ws1", "usertest-admin")
					Expect(err).NotTo(HaveOccurred()) // created

					workspace := convertWorkspaceTodashv1alpha1Workspace(*wsv1Workspace)
					Expect(&dashv1alpha1.Workspace{
						Name:    "ws1",
						OwnerID: "usertest-admin",
						Spec: dashv1alpha1.WorkspaceSpec{
							Template:          "template1",
							Replicas:          0,
							Vars:              map[string]string{"HOGE": "HOGEHOGE"},
							AdditionalNetwork: []dashv1alpha1.NetworkRule{},
						},
						Status: dashv1alpha1.WorkspaceStatus{},
					}).Should(Equal(workspace))
				})
			})

			When("optional arguments are omitted", func() {
				It("should be registered default value", func() {

					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPost, path: "/api/v1alpha1/user/usertest-admin/workspace",
							body: `{"name": "ws1","template": "template1"}`,
						},
						response{
							statusCode: http.StatusCreated,
							body: `{
								"message": "Successfully created",
								"workspace": {
								  "name": "ws1",
								  "ownerID": "usertest-admin",
								  "spec": {
									"template": "template1",
									"replicas": 0
								  },
								  "status": {
									"phase": "Pending"
								  }
								}
							  }`,
						},
					)
					wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), "ws1", "usertest-admin")
					Expect(err).NotTo(HaveOccurred()) // created

					workspace := convertWorkspaceTodashv1alpha1Workspace(*wsv1Workspace)
					Expect(&dashv1alpha1.Workspace{
						Name:    "ws1",
						OwnerID: "usertest-admin",
						Spec: dashv1alpha1.WorkspaceSpec{
							Template:          "template1",
							Replicas:          0,
							AdditionalNetwork: []dashv1alpha1.NetworkRule{},
						},
						Status: dashv1alpha1.WorkspaceStatus{},
					}).Should(Equal(workspace))
				})
			})

		})
	})

	When("Get all workspace of user", func() {

		Describe("invalid request error", func() {

			When("user is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodGet, path: "/api/v1alpha1/user/xxxx/workspace"},
						response{statusCode: http.StatusNotFound, body: `{"message": "user is not found"}`},
					)
				})
			})

		})

		Describe("valid request", func() {

			When("workspace is empty", func() {
				It("should return no item", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin/workspace"},
						response{statusCode: http.StatusOK, body: `{"message":"No items found","items":[]}`})
				})
			})

			When("workspace is not empty", func() {
				It("should return items", func() {
					test_CreateWorkspace("usertest-admin", "ws1", "template1", nil)
					test_CreateWorkspace("usertest-admin", "ws2", "template1", nil)
					test_createNetworkRule("usertest-admin", "ws2", "nw1", 1111, "gp1", "/")
					test_createNetworkRule("usertest-admin", "ws2", "nw3", 2222, "gp1", "/")
					test_createNetworkRule("usertest-admin", "ws2", "nw2", 3333, "gp1", "/")

					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin/workspace"},
						response{statusCode: http.StatusOK, body: `{"items":[` +
							`{"name": "ws1","ownerID": "usertest-admin","spec": {"template": "template1","replicas": 0},"status": {"phase": ""}},` +
							`{"name": "ws2","ownerID": "usertest-admin","spec": {"template": "template1","replicas": 0,` +
							`"additionalNetwork": [` +
							`{"portName": "nw1","portNumber": 1111,"group": "gp1","httpPath": "/","public": false},` +
							`{"portName": "nw2","portNumber": 3333,"group": "gp1","httpPath": "/","public": false},` +
							`{"portName": "nw3","portNumber": 2222,"group": "gp1","httpPath": "/","public": false}` +
							`]},` +
							`"status": {"phase": ""}}` +
							`]}`,
						},
					)
				})
			})
		})
	})

	When("get workspace by name", func() {

		Describe("invalid request error", func() {

			When("user is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodGet, path: "/api/v1alpha1/user/xxxxxx/workspace/ws1"},
						response{statusCode: http.StatusNotFound, body: `{"message": "user is not found"}`},
					)
				})
			})

			When("workspace is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin/workspace/xxx"},
						response{statusCode: http.StatusNotFound, body: `{"message": "workspace is not found"}`},
					)
				})
			})

		})

		Describe("valid request", func() {

			When("access API with normal user session and get a own workspace", func() {
				It("should return a workspace", func() {
					test_CreateWorkspace("usertest", "ws1", "template1", map[string]string{"HOGE": "HOGEHOGE"})
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest/workspace/ws1"},
						response{statusCode: http.StatusOK, body: `{
							"workspace": { "name": "ws1","ownerID": "usertest","spec": {"template": "template1","replicas": 0,"vars": {"HOGE": "HOGEHOGE"}},"status": { "phase": ""}}
						  }`,
						})
				})
			})
		})

	})

	When("Delete workspace", func() {

		Describe("invalid request error", func() {

			When("user is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodDelete, path: "/api/v1alpha1/user/xxxx/workspace/ws1"},
						response{statusCode: http.StatusNotFound, body: `{"message": "user is not found"}`},
					)
				})
			})

			When("workspace is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin/workspace/xxx"},
						response{statusCode: http.StatusNotFound, body: `{"message": "workspace is not found"}`},
					)
				})
			})
		})

		Describe("valid request", func() {

			When("workspace is exist", func() {
				It("should successfully deleted", func() {
					test_CreateWorkspace("usertest", "ws1", "template1", map[string]string{"HOGE": "HOGEHOGE"})

					test_HttpSendAndVerify(userSession,
						request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest/workspace/ws1"},
						response{
							statusCode: http.StatusOK,
							body: `{"message": "Successfully deleted",` +
								`"workspace": {"name": "ws1","ownerID": "usertest",` +
								`"spec": {"template": "template1","replicas": 0,"vars": {"HOGE": "HOGEHOGE"}},"status": {"phase": ""}}}`,
						},
					)
					workspace, _ := k8sClient.GetWorkspaceByUserID(context.Background(), "ws1", "user-delete1")
					Expect(workspace).Should(BeNil()) // deleted
				})
			})

		})
	})

	When("Update workspace", func() {

		Describe("invalid request error", func() {

			When("user is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPatch, path: "/api/v1alpha1/user/xxxxx/workspace/ws1", body: `{"replicas": 0}`},
						response{statusCode: http.StatusNotFound, body: `{"message": "user is not found"}`},
					)
				})
			})

			When("workspace is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPatch, path: "/api/v1alpha1/user/usertest-admin/workspace/xxx", body: `{"replicas": 0}`},
						response{statusCode: http.StatusNotFound, body: `{"message": "workspace is not found"}`},
					)
				})
			})

		})

		Describe("valid request", func() {

			When("all valid arguments are specified", func() {
				It("should be registered as specified", func() {

					test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})

					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPatch, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1", body: `{"replicas": 5}`},
						// TODO: message
						response{statusCode: http.StatusOK,
							body: `{"message":"Successfully updated",` +
								`"workspace":{"name":"ws1","ownerID":"usertest-admin","spec":{"template":"template1","replicas":5},"status":{"phase":""}}}`},
					)

					wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), "ws1", "usertest-admin")
					Expect(err).NotTo(HaveOccurred())

					workspace := convertWorkspaceTodashv1alpha1Workspace(*wsv1Workspace)
					Expect(&dashv1alpha1.Workspace{
						Name:    "ws1",
						OwnerID: "usertest-admin",
						Spec: dashv1alpha1.WorkspaceSpec{
							Template:          "template1",
							Replicas:          5,
							AdditionalNetwork: []dashv1alpha1.NetworkRule{},
						},
						Status: dashv1alpha1.WorkspaceStatus{},
					}).Should(Equal(workspace))
				})
			})

			When("optional arguments are omitted", func() {
				It("should be no changed", func() {

					test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})

					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPatch, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1", body: `{}`},
						// TODO: message
						response{statusCode: http.StatusOK,
							body: `{"message":"No change",` +
								`"workspace":{"name":"ws1","ownerID":"usertest-admin","spec":{"template":"template1","replicas":0},"status":{"phase":""}}}`},
					)

					wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), "ws1", "usertest-admin")
					Expect(err).NotTo(HaveOccurred())

					workspace := convertWorkspaceTodashv1alpha1Workspace(*wsv1Workspace)
					Expect(&dashv1alpha1.Workspace{
						Name:    "ws1",
						OwnerID: "usertest-admin",
						Spec: dashv1alpha1.WorkspaceSpec{
							Template:          "template1",
							Replicas:          0,
							AdditionalNetwork: []dashv1alpha1.NetworkRule{},
						},
						Status: dashv1alpha1.WorkspaceStatus{},
					}).Should(Equal(workspace))
				})
			})

		})

	})

	When("Upsert workspace network rule", func() {

		Describe("invalid request error", func() {

			When("user is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPut, path: "/api/v1alpha1/user/xxxxx/workspace/ws1/network/nw2",
							body: `{"portNumber": 3000,"group": "gp2","httpPath": "/","public":false}`,
						},
						response{statusCode: http.StatusNotFound, body: `{"message": "user is not found"}`},
					)
				})
			})

			When("workspace is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/workspace/xxx/network/nw2",
							body: `{"portNumber": 3000,"group": "gp2","httpPath": "/","public":false}`,
						},
						response{statusCode: http.StatusNotFound, body: `{"message": "workspace is not found"}`},
					)
				})
			})

			When("no change in network rules", func() {
				It("should deny with 400 BadRequest", func() {
					test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})

					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw2",
							body: `{"portNumber": 3000,"group": "gp2","httpPath": "/","public":false}`,
						},
						response{statusCode: http.StatusOK, body: "@ignore"},
					)
					By("Update with the same rules")
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw2",
							body: `{"portNumber": 3000,"group": "gp2","httpPath": "/","public":false}`,
						},
						response{statusCode: http.StatusBadRequest, body: `{ "message": "no change in network rules"}`},
					)
				})
			})

		})

		Describe("valid request", func() {

			When("all valid arguments are specified", func() {
				It("should be registered as specified", func() {

					test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})

					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw2",
							body: `{"portNumber": 3000,"group": "gp2","httpPath": "/","public":false}`,
						},
						response{
							statusCode: http.StatusOK,
							body: `{"message":"Successfully upserted network rule",` +
								`"networkRule":{"portName":"nw2","portNumber":3000,"group":"gp2","httpPath":"/","public":false}}`},
					)

					wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), "ws1", "usertest-admin")
					Expect(err).NotTo(HaveOccurred())

					workspace := convertWorkspaceTodashv1alpha1Workspace(*wsv1Workspace)
					Expect(&dashv1alpha1.Workspace{
						Name:    "ws1",
						OwnerID: "usertest-admin",
						Spec: dashv1alpha1.WorkspaceSpec{
							Template: "template1",
							Replicas: 0,
							AdditionalNetwork: []dashv1alpha1.NetworkRule{
								{PortName: "nw2", PortNumber: 3000, Group: "gp2", HttpPath: "/", Url: ""},
							},
						},
						Status: dashv1alpha1.WorkspaceStatus{Phase: "", MainUrl: "", UrlBase: ""},
					}).Should(Equal(workspace))
				})
			})

			When("optional arguments are omitted", func() {
				It("should be registered as specified", func() {

					test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})

					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw2",
							body: `{"portNumber": 3000,"public":true}`,
						},
						response{
							statusCode: http.StatusOK,
							body: `{"message":"Successfully upserted network rule",` +
								`"networkRule":{"portName":"nw2","portNumber":3000,"public":true}}`},
					)

					wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), "ws1", "usertest-admin")
					Expect(err).NotTo(HaveOccurred())

					workspace := convertWorkspaceTodashv1alpha1Workspace(*wsv1Workspace)
					Expect(&dashv1alpha1.Workspace{
						Name:    "ws1",
						OwnerID: "usertest-admin",
						Spec: dashv1alpha1.WorkspaceSpec{
							Template: "template1",
							Replicas: 0,
							AdditionalNetwork: []dashv1alpha1.NetworkRule{
								{PortName: "nw2", PortNumber: 3000, Group: "", HttpPath: "", Url: "", Public: true},
							},
						},
						Status: dashv1alpha1.WorkspaceStatus{Phase: "", MainUrl: "", UrlBase: ""},
					}).Should(Equal(workspace))
				})
			})

		})

	})

	When("Remove workspace network rule", func() {

		Describe("invalid request error", func() {

			When("user is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodDelete, path: "/api/v1alpha1/user/xxxxx/workspace/ws1/network/nw2"},
						response{statusCode: http.StatusNotFound, body: `{"message": "user is not found"}`},
					)
				})
			})

			When("workspace is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin/workspace/xxx/network/nw2"},
						response{statusCode: http.StatusNotFound, body: `{"message": "workspace is not found"}`},
					)
				})
			})

			When("network rule is not found", func() {
				It("should deny with 404 NotFound", func() {

					test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})

					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw2"},
						response{statusCode: http.StatusBadRequest, body: `{"message":"port name nw2 is not found"}`},
					)
				})
			})

		})

		Describe("valid request", func() {

			When("all valid arguments are specified", func() {
				It("should be registered as specified", func() {

					test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})
					test_createNetworkRule("usertest-admin", "ws1", "nw1", 9999, "gp1", "/")

					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw1",
						},
						response{
							statusCode: http.StatusOK,
							body: `{"message":"Successfully removed network rule",` +
								`"networkRule":{"portName":"nw1","portNumber":9999,"group":"gp1","httpPath":"/","public":false}}`},
					)

					wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), "ws1", "usertest-admin")
					Expect(err).NotTo(HaveOccurred())

					workspace := convertWorkspaceTodashv1alpha1Workspace(*wsv1Workspace)
					Expect(&dashv1alpha1.Workspace{
						Name:    "ws1",
						OwnerID: "usertest-admin",
						Spec: dashv1alpha1.WorkspaceSpec{
							Template:          "template1",
							Replicas:          0,
							AdditionalNetwork: []dashv1alpha1.NetworkRule{},
						},
						Status: dashv1alpha1.WorkspaceStatus{Phase: "", MainUrl: "", UrlBase: ""},
					}).Should(Equal(workspace))
				})
			})
		})

	})

})
