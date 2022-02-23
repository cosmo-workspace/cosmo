package dashboard

import (
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
)

var _ = Describe("Dashboard server [auth]", func() {

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession("usertest", "お名前", "", "password")
		adminSession = test_CreateLoginUserSession("usertest-admin", "アドミン", wsv1alpha1.UserAdminRole, "password")
	})

	AfterEach(func() {
		test_DeleteCosmoUserAll()
	})

	When("Login", func() {

		When("user is empty", func() {
			It("should deny with 400 BadRequest", func() {
				test_HttpSendAndVerify(nil,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/login", body: `{"password": "password"}`},
					response{statusCode: http.StatusBadRequest, body: `{"message": "required field 'id' is zero value."}`},
				)
			})
		})

		When("password is empty", func() {
			It("should deny 400 BadRequest", func() {
				test_HttpSendAndVerify(nil,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/login", body: `{"id": "usertest"}`},
					response{statusCode: http.StatusBadRequest, body: `{"message": "required field 'password' is zero value."}`},
				)
			})
		})

		When("invalid password", func() {
			It("should deny with 403 Forbidden", func() {
				test_HttpSendAndVerify(nil,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/login", body: `{"id": "usertest", "password": "invalid"}`},
					response{statusCode: http.StatusForbidden, body: `{"message": "incorrect user or password"}`},
				)
			})
		})

		When("user is not found", func() {
			It("should deny with 403 Forbidden", func() {
				test_HttpSendAndVerify(nil,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/login", body: `{"id": "xxxxxxx", "password": "password"}`},
					response{statusCode: http.StatusForbidden, body: `{"message": "incorrect user or password"}`},
				)
			})
		})

		When("valid user authentication", func() {
			It("should success and response with session cookie", func() {
				got, gotBody := test_HttpSend(nil,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/login", body: `{"id": "usertest", "password": "password"}`},
				)

				Expect(got).Should(HaveHTTPStatus(http.StatusOK))

				m := make(map[string]interface{})
				json.Unmarshal(gotBody, &m)

				Expect(m["id"]).Should(Equal("usertest"))

				_, err := time.Parse(time.RFC3339Nano, m["expireAt"].(string))
				Expect(err).NotTo(HaveOccurred())

				Expect(got.Cookies()).ShouldNot(BeNil())
			})
		})

		When("valid admin authentication", func() {
			It("should success and response with session cookie", func() {
				got, gotBody := test_HttpSend(nil,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/login", body: `{"id": "usertest-admin", "password": "password"}`},
				)

				Expect(got).Should(HaveHTTPStatus(http.StatusOK))

				m := make(map[string]interface{})
				json.Unmarshal(gotBody, &m)

				Expect(m["id"]).Should(Equal("usertest-admin"))

				_, err := time.Parse(time.RFC3339Nano, m["expireAt"].(string))
				Expect(err).NotTo(HaveOccurred())

				Expect(got.Cookies()).ShouldNot(BeNil())
			})
		})
	})

	When("Verify", func() {

		When("invalid user session", func() {
			It("should deny with 403 Forbidden", func() {
				session := test_Login("usertest", "password")

				By("logout")
				got, _ := test_HttpSend(session,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/logout"},
				)
				Expect(got).Should(HaveHTTPStatus(http.StatusOK))
				invalidSession := got.Cookies()

				By("verify")
				test_HttpSendAndVerify(invalidSession,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/verify"},
					response{statusCode: http.StatusUnauthorized, body: ""},
				)
			})
		})

		When("valid user session", func() {
			It("should success and response with session cookie", func() {

				session := test_Login("usertest", "password")

				got, gotBody := test_HttpSend(session,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/verify"},
				)

				Expect(got).Should(HaveHTTPStatus(http.StatusOK))

				m := make(map[string]interface{})
				json.Unmarshal(gotBody, &m)

				Expect(m["id"]).Should(Equal("usertest"))

				_, err := time.Parse(time.RFC3339Nano, m["expireAt"].(string))
				Expect(err).NotTo(HaveOccurred())

				Expect(got.Cookies()).ShouldNot(BeNil())
			})
		})
	})

	When("Logout", func() {

		When("invalid user session", func() {
			It("should deny with 403 Forbidden", func() {
				session := test_Login("usertest", "password")

				By("logout")
				got, _ := test_HttpSend(session,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/logout"},
				)
				Expect(got).Should(HaveHTTPStatus(http.StatusOK))
				invalidSession := got.Cookies()

				By("invalid logout")
				test_HttpSendAndVerify(invalidSession,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/logout"},
					response{statusCode: http.StatusUnauthorized, body: ""},
				)
			})
		})

		When("valid user session", func() {
			It("should success and response with session cookie", func() {

				session := test_Login("usertest", "password")

				got, gotBody := test_HttpSend(session,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/verify"},
				)

				Expect(got).Should(HaveHTTPStatus(http.StatusOK))

				m := make(map[string]interface{})
				json.Unmarshal(gotBody, &m)

				Expect(m["id"]).Should(Equal("usertest"))

				_, err := time.Parse(time.RFC3339Nano, m["expireAt"].(string))
				Expect(err).NotTo(HaveOccurred())

				Expect(got.Cookies()).ShouldNot(BeNil())
			})
		})
	})
})
