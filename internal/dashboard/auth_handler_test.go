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

	Describe("Login authentication", func() {

		When("access login API with invalid user authentication", func() {
			It("should deny with 403 Forbidden", func() {
				test_HttpSendAndVerify(nil,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/login", body: `{"id": "usertest", "password": "invalid"}`},
					response{statusCode: http.StatusForbidden, body: ""},
				)
			})
		})

		When("access login API with valid user authentication", func() {
			It("should success and response with session cookie", func() {
				got, gotBody := test_HttpSend(nil,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/login", body: `{"id": "usertest", "password": "password"}`},
				)

				m := make(map[string]interface{})
				json.Unmarshal(gotBody, &m)

				Expect(m["id"]).Should(Equal("usertest"))

				_, err := time.Parse(time.RFC3339Nano, m["expireAt"].(string))
				Expect(err).NotTo(HaveOccurred())

				Expect(got.Cookies()).ShouldNot(BeNil())
			})
		})

		When("access login API with valid admin authentication", func() {
			It("should success and response with session cookie", func() {
				got, gotBody := test_HttpSend(nil,
					request{method: http.MethodPost, path: "/api/v1alpha1/auth/login", body: `{"id": "usertest-admin", "password": "password"}`},
				)

				m := make(map[string]interface{})
				json.Unmarshal(gotBody, &m)

				Expect(m["id"]).Should(Equal("usertest-admin"))

				_, err := time.Parse(time.RFC3339Nano, m["expireAt"].(string))
				Expect(err).NotTo(HaveOccurred())

				Expect(got.Cookies()).ShouldNot(BeNil())
			})
		})
	})
})
