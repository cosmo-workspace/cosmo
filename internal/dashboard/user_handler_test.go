package dashboard

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	dashv1alpha1 "github.com/cosmo-workspace/cosmo/api/openapi/dashboard/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
)

var _ = Describe("Dashboard server [User]", func() {

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession("usertest", "お名前", "", "password")
		adminSession = test_CreateLoginUserSession("usertest-admin", "アドミン", wsv1alpha1.UserAdminRole, "password")
	})

	AfterEach(func() {
		test_DeleteCosmoUserAll()
		test_DeleteTemplateAll()
	})

	Describe("authorization by role", func() {

		var session []*http.Cookie

		deny403 := func(whenText string, request request) {
			When(whenText, func() {
				It("should deny with 403 Forbidden", func() {
					test_HttpSendAndVerify(session, request, response{statusCode: http.StatusForbidden, body: `{"message": "not authorized"}`})
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
			})

			When("update own resource", func() {
				deny403("Create a new User", request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create" }`})
				deny403("Get all users", request{method: http.MethodGet, path: "/api/v1alpha1/user"})
				ok200("Get user by ID", request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest"})
				ok200("Delete user by ID", request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest"})
				ok200("Update user name", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/name", body: `{"displayName": "newname"}`})
				deny403("Update user role", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/role", body: `{"role": "cosmo-admin"}`})
				ok200("Update user password", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/password", body: `{ "currentPassword": "password", "newPassword": "newPassword"}`})
			})

			When("update resource of others", func() {
				deny403("Create a new User", request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create" }`})
				deny403("Get all users", request{method: http.MethodGet, path: "/api/v1alpha1/user"})
				deny403("Get user by ID", request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin"})
				deny403("Delete user by ID", request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin"})
				deny403("Update user name", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/name", body: `{"displayName": "newname"}`})
				deny403("Update user role", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/role", body: `{"role": "cosmo-admin"}`})
				deny403("Update user password", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/password", body: `{ "currentPassword": "password", "newPassword": "newPassword"}`})
			})

		})

		When("access API with admin user session", func() {

			BeforeEach(func() {
				session = adminSession
				test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")
			})

			When("update own resource", func() {
				ok201("Create a new User", request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create" }`})
				ok200("Get all users", request{method: http.MethodGet, path: "/api/v1alpha1/user"})
				ok200("Get user by ID", request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin"})
				ok200("Delete user by ID", request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin"})
				ok200("Update user name", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/name", body: `{"displayName": "newname"}`})
				ok200("Update user role", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/role", body: `{"role": "cosmo-admin"}`})
				ok200("Update user password", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/password", body: `{ "currentPassword": "password", "newPassword": "newPassword"}`})
			})

			When("update resource of others", func() {
				ok201("Create a new User", request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create" }`})
				ok200("Get all users", request{method: http.MethodGet, path: "/api/v1alpha1/user"})
				ok200("Get user by ID", request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest"})
				ok200("Delete user by ID", request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest"})
				ok200("Update user name", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/name", body: `{"displayName": "newname"}`})
				ok200("Update user role", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/role", body: `{"role": "cosmo-admin"}`})
				ok200("Update user password", request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/password", body: `{ "currentPassword": "password", "newPassword": "newPassword"}`})
			})

		})

	})

	When("create a new User", func() {

		Describe("invalid request error", func() {

			When("user role is invalid", func() {
				It("should deny with 400 BadRequest", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create", "role": "xxxxxx"}`},
						response{statusCode: http.StatusBadRequest, body: `{"message": "'userrole' is invalid"}`},
					)
				})
			})

			When("authtype is invalid", func() {
				It("should deny with 400 BadRequest", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create", "authType": "xxxxx"}`},
						response{statusCode: http.StatusBadRequest, body: `{"message": "'authtype' is invalid"}`},
					)
				})
			})

			When("user id empty", func() {
				It("should deny with 400 BadRequest", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": ""}`},
						response{statusCode: http.StatusBadRequest, body: `{"message":"required field 'id' is zero value."}`},
					)
				})
			})

			When("user id is invalid (include upper case)", func() {
				It("should deny with 503 ServiceUnavailable", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-createX"}`},
						response{statusCode: http.StatusServiceUnavailable, body: `{"message":"failed to create user"}`},
					)
				})
			})

			When("user is already exists", func() {
				It("should deny witn 429 TooManyRequests", func() {
					test_CreateCosmoUser("user-create-alreadyexist", "already", "")

					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create-alreadyexist"}`},
						response{statusCode: http.StatusTooManyRequests, body: `{"message": "user already exists"}`},
					)
				})
			})
		})

		Describe("valid request", func() {

			When("all valid arguments are specified", func() {
				It("should be registered as specified", func() {
					test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")
					test_CreateTemplate(wsv1alpha1.TemplateTypeUserAddon, "user-temple1")

					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPost,
							path:   "/api/v1alpha1/user",
							body: `{ "id": "user-create", "displayName": "create 1", "role":"cosmo-admin", "authType": "kosmo-secret",` +
								`"addons": [{"template": "user-temple1","vars": {"HOGE": "FUGA"}}]}`,
						},
						response{
							statusCode: http.StatusCreated,
							body: `{"message": "Successfully created",` +
								`"user": { "id": "user-create", "displayName": "create 1", "role":"cosmo-admin", "authType": "kosmo-secret", ` +
								`"addons": [{"template": "user-temple1", "vars": {"HOGE": "FUGA"}}],` +
								`"defaultPassword": "%s"}}`,
							bodyValues: func() []interface{} {
								defaultPass, _ := k8sClient.GetDefaultPassword(context.Background(), "user-create")
								return []interface{}{*defaultPass}
							},
						},
					)
					userObj, err := k8sClient.GetUser(context.Background(), "user-create")
					Expect(err).NotTo(HaveOccurred()) // created

					user := convertUserToDashv1alpha1User(*userObj)
					Expect(&dashv1alpha1.User{
						Id:          "user-create",
						DisplayName: "create 1",
						Role:        "cosmo-admin",
						AuthType:    "kosmo-secret",
						Addons: []dashv1alpha1.ApiV1alpha1UserAddons{
							{
								Template: "user-temple1",
								Vars:     map[string]string{"HOGE": "FUGA"},
							},
						},
					}).Should(Equal(user))
				})
			})

			When("optional arguments are omitted", func() {
				It("should be registered default value", func() {
					test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")

					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPost,
							path:   "/api/v1alpha1/user",
							body:   `{ "id": "user-create"}`,
						},
						response{
							statusCode: http.StatusCreated,
							body: `{"message": "Successfully created",` +
								`"user": { "id": "user-create", "displayName": "user-create", "authType": "kosmo-secret", ` +
								`"defaultPassword": "%s"}}`,
							bodyValues: func() []interface{} {
								defaultPass, _ := k8sClient.GetDefaultPassword(context.Background(), "user-create")
								return []interface{}{*defaultPass}
							},
						},
					)
					userObj, err := k8sClient.GetUser(context.Background(), "user-create")
					Expect(err).NotTo(HaveOccurred()) // created

					user := convertUserToDashv1alpha1User(*userObj)
					Expect(&dashv1alpha1.User{
						Id:          "user-create",
						DisplayName: "user-create",
						Role:        "",
						AuthType:    "kosmo-secret",
						Addons:      []dashv1alpha1.ApiV1alpha1UserAddons{},
					}).Should(Equal(user))
				})
			})
		})

		Describe("default password creation timing", func() {

			When("password create immediately", func() {
				It("should succeed", func() {
					test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")

					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create"}`},
						response{statusCode: http.StatusCreated, body: "@ignore"},
					)
					user, _ := k8sClient.GetUser(context.Background(), "user-create")
					Expect(user).ShouldNot(BeNil()) // created
				})
			})

			When("password create later", func() {
				It("should succeed", func() {
					timer := time.AfterFunc(100*time.Millisecond, func() {
						test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create-later")
					})
					defer timer.Stop()

					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create-later"}`},
						response{statusCode: http.StatusCreated, body: "@ignore"},
					)
					user, _ := k8sClient.GetUser(context.Background(), "user-create-later")
					Expect(user).ShouldNot(BeNil()) // created
				})
			})

			When("password create timeout", func() {
				It("should create user and fail with 503 ServiceUnavailable", func() {
					timer := time.AfterFunc(30*time.Second, func() {
						test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create-timeout")
					})
					defer timer.Stop()

					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create-timeout"}`},
						response{
							statusCode: http.StatusServiceUnavailable,
							body:       `{"message": "Request timeout"}`,
						},
					)
					user, _ := k8sClient.GetUser(context.Background(), "user-create-timeout")
					Expect(user).ShouldNot(BeNil()) // created
				})
			})
		})
	})

	When("get all users", func() {

		Describe("invalid request error", func() {
			// nothing
		})

		When("user is empty", func() {
			// Can't test
		})

		When("user is not empty", func() {
			It("should return items", func() {
				test_HttpSendAndVerify(adminSession,
					request{method: http.MethodGet, path: "/api/v1alpha1/user"},
					response{
						statusCode: http.StatusOK,
						body: `{"items":[` +
							`{"id":"usertest","displayName":"お名前","authType":"kosmo-secret"},` +
							`{"id":"usertest-admin","displayName":"アドミン","role":"cosmo-admin","authType":"kosmo-secret"}` +
							`]}`,
					},
				)
			})
		})
	})

	When("get user by ID", func() {

		Describe("invalid request error", func() {

			When("user is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodGet, path: "/api/v1alpha1/user/XXXXX"},
						response{statusCode: http.StatusNotFound, body: `{"message": "user is not found"}`},
					)
				})
			})
		})

		Describe("valid request", func() {

			It("should return item", func() {
				test_HttpSendAndVerify(userSession,
					request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest"},
					response{
						statusCode: http.StatusOK,
						body:       `{"user": { "id": "usertest", "displayName": "お名前",  "authType": "kosmo-secret"}}`,
					},
				)
			})
		})

	})

	When("delete user by ID", func() {

		Describe("invalid request error", func() {

			When("user is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodDelete, path: "/api/v1alpha1/user/XXXXX"},
						response{statusCode: http.StatusNotFound, body: `{"message": "user is not found"}`},
					)
				})
			})

			When("user is empty", func() {
				// Can't test
			})
		})

		Describe("valid request", func() {

			When("user is exist", func() {
				It("should successfully deleted", func() {
					test_CreateCosmoUser("user-delete1", "delete", "")

					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodDelete, path: "/api/v1alpha1/user/user-delete1"},
						response{
							statusCode: http.StatusOK,
							body: `{ "message": "Successfully deleted",` +
								`"user": { "id": "user-delete1", "displayName": "delete", "authType": "kosmo-secret"}}`,
						},
					)
					user, _ := k8sClient.GetUser(context.Background(), "user-delete1")
					Expect(user).Should(BeNil()) // deleted
				})
			})

		})
	})

	When("update user name", func() {

		Describe("invalid request error", func() {

			When("user is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPut, path: "/api/v1alpha1/user/XXXXXX/name", body: `{"displayName": "namechanged"}`},
						response{statusCode: http.StatusNotFound, body: `{"message": "user is not found"}`},
					)
				})
			})

			When("diplayName is empty", func() {
				It("should deny with 400 BadRequest", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/name", body: `{"displayName": ""}`},
						response{statusCode: http.StatusBadRequest, body: `{"message":"required field 'displayName' is zero value."}`},
					)
				})
			})

			When("diplayName is not specified", func() {
				It("should deny with 400 BadRequest", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/name", body: `{}`},
						response{statusCode: http.StatusBadRequest, body: `{"message":"required field 'displayName' is zero value."}`},
					)
				})
			})

		})

		Describe("valid request", func() {

			When("update own name", func() {
				It("should successfully updated", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/name", body: `{"displayName": "namechanged"}`},
						response{
							statusCode: http.StatusOK,
							body: `{"message": "Successfully updated",` +
								`"user": { "id": "usertest", "displayName": "namechanged", "authType": "kosmo-secret"}}`,
						},
					)
					userObj, err := k8sClient.GetUser(context.Background(), "usertest")
					Expect(err).NotTo(HaveOccurred())

					user := convertUserToDashv1alpha1User(*userObj)
					Expect(&dashv1alpha1.User{
						Id:          "usertest",
						DisplayName: "namechanged",
						Role:        "",
						AuthType:    "kosmo-secret",
						Addons:      []dashv1alpha1.ApiV1alpha1UserAddons{},
					}).Should(Equal(user))
				})
			})

		})

	})

	When("update user role", func() {

		Describe("invalid request error", func() {

			When("user is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPut, path: "/api/v1alpha1/user/XXXXXX/role", body: `{"role": "cosmo-admin"}`},
						response{statusCode: http.StatusNotFound, body: `{"message": "user is not found"}`},
					)
				})
			})

			When("user role is invalid", func() {
				It("should deny with 400 BadRequest", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/role", body: `{"role": "xxxxx"}`},
						response{statusCode: http.StatusBadRequest, body: `{"message": "'userrole' is invalid"}`},
					)
				})
			})
		})

		Describe("valid request", func() {

			When("access API with admin user session and update own name", func() {
				It("should successfully updated", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/role", body: `{"role": "cosmo-admin"}`},
						response{
							statusCode: http.StatusOK,
							body: `{"message": "Successfully updated",` +
								`"user": { "id": "usertest", "displayName": "お名前", "role": "cosmo-admin", "authType": "kosmo-secret"}}`,
						},
					)
					userObj, err := k8sClient.GetUser(context.Background(), "usertest")
					Expect(err).NotTo(HaveOccurred())

					user := convertUserToDashv1alpha1User(*userObj)
					Expect(&dashv1alpha1.User{
						Id:          "usertest",
						DisplayName: "お名前",
						Role:        "cosmo-admin",
						AuthType:    "kosmo-secret",
						Addons:      []dashv1alpha1.ApiV1alpha1UserAddons{},
					}).Should(Equal(user))
				})
			})

			When("role is empty", func() {
				It("should successfully updated", func() {
					test_HttpSendAndVerify(adminSession,
						request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/role", body: `{"role": ""}`},
						response{
							statusCode: http.StatusOK,
							body: `{"message": "Successfully updated",` +
								`"user": { "id": "usertest", "displayName": "お名前", "authType": "kosmo-secret"}}`,
						},
					)
					userObj, err := k8sClient.GetUser(context.Background(), "usertest")
					Expect(err).NotTo(HaveOccurred())

					user := convertUserToDashv1alpha1User(*userObj)
					Expect(&dashv1alpha1.User{
						Id:          "usertest",
						DisplayName: "お名前",
						Role:        "",
						AuthType:    "kosmo-secret",
						Addons:      []dashv1alpha1.ApiV1alpha1UserAddons{},
					}).Should(Equal(user))
				})
			})

		})
	})

	When("update user password", func() {

		Describe("invalid request error", func() {

			When("user is not found", func() {
				It("should deny with 404 NotFound", func() {
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPut, path: "/api/v1alpha1/user/xxxxxxxx/password",
							body: `{ "currentPassword": "password", "newPassword": "newPassword"}`},
						response{statusCode: http.StatusNotFound, body: `{"message": "user is not found"}`},
					)
				})
			})

			When("currentPassword is empty", func() {
				It("should deny with 400 BadRequest", func() {
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/password",
							body: `{ "currentPassword": "", "newPassword": "newPassword"}`},
						response{statusCode: http.StatusBadRequest, body: `{"message":"required field 'currentPassword' is zero value."}`},
					)
				})
			})

			When("currentPassword is invarid", func() {
				It("should deny with 403 Forbidden", func() {
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/password",
							body: `{ "currentPassword": "xxxxxx", "newPassword": "newPassword"}`},
						response{statusCode: http.StatusForbidden, body: `{"message":"current password is invalid"}`},
					)
				})
			})

			When("newPassword is empty", func() {
				It("should deny with 400 BadRequest", func() {
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/password",
							body: `{ "currentPassword": "password", "newPassword": ""}`},
						response{statusCode: http.StatusBadRequest, body: `{"message":"required field 'newPassword' is zero value."}`},
					)
				})
			})

		})

		Describe("valid request", func() {

			When("update own password", func() {
				It("should successfully updated", func() {
					test_HttpSendAndVerify(adminSession,
						request{
							method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/password",
							body: `{ "currentPassword": "password", "newPassword": "newPassword"}`},
						response{
							statusCode: http.StatusOK,
							body:       `{"message": "Successfully updated"}`,
						},
					)
					verified, _, _ := k8sClient.VerifyPassword(context.Background(), "usertest-admin", []byte("newPassword"))
					Expect(verified).Should(BeTrue())
				})
			})
		})
	})
})
