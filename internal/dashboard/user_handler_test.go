package dashboard

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
)

var _ = Describe("Dashboard server [User]", func() {

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession("usertest", "お名前", "", "password")
		adminSession = test_CreateLoginUserSession("usertest-admin", "アドミン", wsv1alpha1.UserAdminRole, "password")
	})

	AfterEach(func() {
		clientMock.Clear()
		test_DeleteCosmoUserAll()
		test_DeleteTemplateAll()
	})

	//==================================================================================
	replace := func(src, reg, repl string) string {
		return regexp.MustCompile(reg).ReplaceAllString(src, repl)
	}

	userSnap := func(us *wsv1alpha1.User) struct{ Name, Namespace, Spec, Status interface{} } {
		return struct{ Name, Namespace, Spec, Status interface{} }{
			Name:      us.Name,
			Namespace: us.Namespace,
			Spec:      us.Spec,
			Status:    us.Status,
		}
	}
	//==================================================================================
	Describe("authorization by role", func() {

		DescribeTable("access API with admin user session:",
			func(stat int, req request) {
				test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")
				By("---------------test start----------------")
				res, _ := test_HttpSend(adminSession, req)
				Ω(res.StatusCode).Should(Equal(stat))
				By("---------------test end---------------")
			},
			func(stat int, req request) string { return fmt.Sprintf("%d %+v", stat, req) },
			// update own resource
			Entry(nil, 201, request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create" }`}),
			Entry(nil, 200, request{method: http.MethodGet, path: "/api/v1alpha1/user"}),
			Entry(nil, 200, request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin"}),
			Entry(nil, 403, request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin"}),
			Entry(nil, 200, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/name", body: `{"displayName": "newname"}`}),
			Entry(nil, 200, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/role", body: `{"role": ""}`}),
			Entry(nil, 200, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/password", body: `{ "currentPassword": "password", "newPassword": "newPassword"}`}),
			// update resource of others
			Entry(nil, 201, request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create" }`}),
			Entry(nil, 200, request{method: http.MethodGet, path: "/api/v1alpha1/user"}),
			Entry(nil, 200, request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest"}),
			Entry(nil, 200, request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest"}),
			Entry(nil, 200, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/name", body: `{"displayName": "newname"}`}),
			Entry(nil, 200, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/role", body: `{"role": "cosmo-admin"}`}),
			Entry(nil, 200, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/password", body: `{ "currentPassword": "password", "newPassword": "newPassword"}`}),
		)

		DescribeTable("access API with normal user session:",
			func(stat int, req request) {
				test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")
				By("---------------test start----------------")
				res, _ := test_HttpSend(userSession, req)
				Ω(res.StatusCode).Should(Equal(stat))
				By("---------------test end---------------")
			},
			func(stat int, req request) string { return fmt.Sprintf("%d %+v", stat, req) },
			// update own resource
			Entry(nil, 403, request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create" }`}),
			Entry(nil, 403, request{method: http.MethodGet, path: "/api/v1alpha1/user"}),
			Entry(nil, 200, request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest"}),
			Entry(nil, 403, request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest"}),
			Entry(nil, 200, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/name", body: `{"displayName": "newname"}`}),
			Entry(nil, 403, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/role", body: `{"role": "cosmo-admin"}`}),
			Entry(nil, 200, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/password", body: `{ "currentPassword": "password", "newPassword": "newPassword"}`}),
			// update resource of others
			Entry(nil, 403, request{method: http.MethodPost, path: "/api/v1alpha1/user", body: `{ "id": "user-create" }`}),
			Entry(nil, 403, request{method: http.MethodGet, path: "/api/v1alpha1/user"}),
			Entry(nil, 403, request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin"}),
			Entry(nil, 403, request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin"}),
			Entry(nil, 403, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/name", body: `{"displayName": "newname"}`}),
			Entry(nil, 403, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/role", body: `{"role": ""}`}),
			Entry(nil, 403, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/password", body: `{ "currentPassword": "password", "newPassword": "newPassword"}`}),
		)
	})

	//==================================================================================
	Describe("[PostUser]", func() {

		run_test := func(userId, requestBody string) {
			if userId == "user-create" {
				test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")
				test_CreateTemplate(wsv1alpha1.TemplateTypeUserAddon, "user-temple1")
			} else if userId == "user-create-later" {
				timer := time.AfterFunc(100*time.Millisecond, func() {
					test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create-later")
				})
				defer timer.Stop()
			} else if userId == "user-create-timeout" {
				timer := time.AfterFunc(30*time.Second, func() {
					test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create-timeout")
				})
				defer timer.Stop()
			}
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodPost, path: "/api/v1alpha1/user", body: requestBody})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(replace(string(body), `"defaultPassword":".*"`, `"defaultPassword":"xxxxxxxx"`)).To(MatchSnapShot())

			if userId != "" {
				wsv1User, err := k8sClient.GetUser(context.Background(), userId)
				if res.StatusCode == http.StatusCreated {
					Expect(err).NotTo(HaveOccurred())
					Ω(userSnap(wsv1User)).To(MatchSnapShot())
				} else {
					Expect(err).To(HaveOccurred())
				}
			}
			By("---------------test end---------------")
		}
		desc := func(args ...string) string { return strings.Join(args, " ") }

		DescribeTable("✅ success succeed in normal context:",
			run_test,
			Entry(desc, "user-create", `{ "id": "user-create", "displayName": "create 1", "role":"cosmo-admin", "authType": "kosmo-secret","addons": [{"template": "user-temple1","vars": {"HOGE": "FUGA"}}]}`),
			Entry(desc, "user-create", `{ "id": "user-create"}`),
			Entry(desc, "user-create-later", `{ "id": "user-create-later"}`),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(desc, "user-create", `{ "id": "user-create", "role": "xxxxxx"}`),
			Entry(desc, "user-create", `{ "id": "user-create", "authType": "xxxxx"}`),
			Entry(desc, "", `{ "id": ""}`),
			Entry(desc, "user-createX", `{ "id": "user-createX"}`),
			Entry(desc, "", `{ "id": "usertest"}`),
		)

		DescribeTable("❌ fail to create password timeout",
			run_test,
			Entry(desc, "", `{ "id": "user-create-timeout"}`),
		)
	})

	//==================================================================================
	Describe("[GetUsers]", func() {

		run_test := func() {
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodGet, path: "/api/v1alpha1/user"})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil),
		)

		DescribeTable("✅ success with empty user:",
			func() {
				clientMock.SetListError((*Server).GetUsers, nil)
				run_test()
			},
			Entry(nil),
		)

		DescribeTable("❌ fail with an unexpected error at list:",
			func() {
				clientMock.SetListError((*Server).GetUsers, errors.New("mock user list error"))
				run_test()
			},
			Entry(nil),
		)
	})

	//==================================================================================
	Describe("[GetUser]", func() {

		run_test := func(userId string) {
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodGet, path: fmt.Sprintf("/api/v1alpha1/user/%s", userId)})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "usertest"),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "XXXXX"),
		)

		DescribeTable("❌ fail with an unexpected error to get:",
			func(userId string) {
				clientMock.SetGetError(`\\.GetUser$`, errors.New("mock get user error"))
				//clientMock.SetGetError((*Server).GetUser, errors.New("get user error"))
				run_test(userId)
			},
			Entry(nil, "usertest"),
		)
	})

	//==================================================================================
	Describe("[DeleteUser]", func() {

		run_test := func(userId string) {
			test_CreateCosmoUser("user-delete1", "delete", "")
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodDelete, path: fmt.Sprintf("/api/v1alpha1/user/%s", userId)})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())

			_, err := k8sClient.GetUser(context.Background(), "user-delete1")
			if res.StatusCode == http.StatusOK {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "user-delete1"),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "xxxxxx"),
			Entry(nil, "usertest-admin"),
		)

		DescribeTable("❌ fail with an unexpected error to get:",
			func(userId string) {
				clientMock.SetGetError(`\.preFetchUserMiddleware\.|\.DeleteUser$`, errors.New("mock get user error")) ///
				//clientMock.SetGetError((*Server).DeleteUser, errors.New("mock get user error"))
				run_test(userId)
			},
			Entry(nil, "user-delete1"),
		)

		DescribeTable("❌ fail with an unexpected error to delete:",
			func(userId string) {
				clientMock.SetDeleteError((*Server).DeleteUser, errors.New("mock delete user error"))
				run_test(userId)
			},
			Entry(nil, "user-delete1"),
		)
	})

	//==================================================================================
	Describe("[PutUserName]", func() {

		run_test := func(userId, requestBody string) {
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodPut, path: fmt.Sprintf("/api/v1alpha1/user/%s/name", userId), body: requestBody})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())

			if res.StatusCode == http.StatusOK {
				wsv1User, err := k8sClient.GetUser(context.Background(), userId)
				Expect(err).NotTo(HaveOccurred())
				Ω(userSnap(wsv1User)).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "usertest", `{"displayName": "namechanged"}`),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "XXXXXX", `{"displayName": "namechanged"}`),
			Entry(nil, "usertest", `{"displayName": ""}`),
			Entry(nil, "usertest", `{}`),
			Entry(nil, "usertest", `{"displayName": "お名前"}`),
		)

		DescribeTable("❌ fail with an unexpected error to update:",
			func(userId, requestBody string) {
				clientMock.SetUpdateError((*Server).PutUserName, errors.New("mock update user error"))
				run_test(userId, requestBody)
			},
			Entry(nil, "usertest", `{"displayName": "namechanged"}`),
		)
	})

	//==================================================================================
	Describe("[PutUserRole]", func() {

		run_test := func(userId, requestBody string) {
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodPut, path: fmt.Sprintf("/api/v1alpha1/user/%s/role", userId), body: requestBody})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())

			if res.StatusCode == http.StatusOK {
				wsv1User, err := k8sClient.GetUser(context.Background(), userId)
				Expect(err).NotTo(HaveOccurred())
				Ω(userSnap(wsv1User)).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "usertest", `{"role": "cosmo-admin"}`),
			Entry(nil, "usertest-admin", `{"role": ""}`),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "XXXXXX", `{"role": "cosmo-admin"}`),
			Entry(nil, "usertest", `{"role": "xxxxx"}`),
			Entry(nil, "usertest-admin", `{"role": "cosmo-admin"}`),
			Entry(nil, "usertest", `{"displayName": "お名前"}`),
		)

		DescribeTable("❌ fail with an unexpected error to update:",
			func(userId, requestBody string) {
				clientMock.SetUpdateError((*Server).PutUserRole, errors.New("mock update user error"))
				run_test(userId, requestBody)
			},
			Entry(nil, "usertest", `{"role": "cosmo-admin"}`),
		)
	})

	//==================================================================================
	Describe("[PutUserPassword]", func() {

		run_test := func(userId, requestBody string) {
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodPut, path: fmt.Sprintf("/api/v1alpha1/user/%s/password", userId), body: requestBody})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())

			if res.StatusCode == http.StatusOK {
				verified, _, _ := k8sClient.VerifyPassword(context.Background(), userId, []byte("newPassword"))
				Expect(verified).Should(BeTrue())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success with invalid request:",
			run_test,
			Entry(nil, "usertest-admin", `{ "currentPassword": "password", "newPassword": "newPassword"}`),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "XXXXXX", `{ "currentPassword": "password", "newPassword": "newPassword"}`),
			Entry(nil, "usertest-admin", `{ "currentPassword": "", "newPassword": "newPassword"}`),
			Entry(nil, "usertest-admin", `{ "currentPassword": "xxxxxx", "newPassword": "newPassword"}`),
			Entry(nil, "usertest-admin", `{ "currentPassword": "password", "newPassword": ""}`),
		)

		DescribeTable("❌ fail to verify password:",
			func(userId, requestBody string) {
				clientMock.GetMock = func(ctx context.Context, key client.ObjectKey, obj client.Object) (mocked bool, err error) {
					if key.Name == wsv1alpha1.UserPasswordSecretName {
						return true, apierrs.NewNotFound(schema.GroupResource{}, "secret")
					}
					return false, nil
				}
				//clientMock.SetGetError((*Server).PutUserPassword, apierrs.NewNotFound(schema.GroupResource{}, "secret"))
				run_test(userId, requestBody)
			},
			Entry(nil, "usertest-admin", `{ "currentPassword": "password", "newPassword": "newPassword"}`),
		)

		DescribeTable("❌ fail with an unexpected error :",
			func(userId, requestBody string) {
				clientMock.SetUpdateError((*Server).PutUserPassword, errors.New("mock update error"))
				run_test(userId, requestBody)
			},
			Entry(nil, "usertest-admin", `{ "currentPassword": "password", "newPassword": "newPassword"}`),
		)
	})
})
