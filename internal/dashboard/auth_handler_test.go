package dashboard

import (
	"net/http"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
)

var _ = Describe("Dashboard server [auth]", func() {

	BeforeEach(func() {
		test_CreateLoginUserSession("usertest", "user", "", "password1")
		test_CreateLoginUserSession("usertest-admin", "admin", wsv1alpha1.UserAdminRole, "password2")
	})

	AfterEach(func() {
		clientMock.Clear()
		test_DeleteCosmoUserAll()
	})

	//==================================================================================
	replace := func(src, reg, repl string) string {
		return regexp.MustCompile(reg).ReplaceAllString(src, repl)
	}

	//==================================================================================
	Describe("[Login]", func() {

		run_test := func(requestBody string) {
			By("---------------test start----------------")
			res, body := test_HttpSend(nil, request{method: http.MethodPost, path: "/api/v1alpha1/auth/login", body: requestBody})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(replace(string(body), `"expireAt":".*"`, `"expireAt":"9999-99-99T99:99:99.99999999+9:00"`)).To(MatchSnapShot())
			Expect(res.Cookies()).ShouldNot(BeNil())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, `{"id": "usertest", "password": "password1"}`),
			Entry(nil, `{"id": "usertest-admin", "password": "password2"}`),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, `{"password": "password1"}`),
			Entry(nil, `{"id": "usertest"}`),
			Entry(nil, `{"id": "usertest", "password": "invalid"}`),
			Entry(nil, `{"id": "xxxxxxx", "password": "password1"}`),
		)
	})

	//==================================================================================
	Describe("[Verify]", func() {

		run_test := func(sessionType string) {
			var session []*http.Cookie = nil
			switch sessionType {
			case "logined session":
				session = test_Login("usertest", "password1")
			case "logouted session":
				session = test_Login("usertest", "password1")
				res, _ := test_HttpSend(session, request{method: http.MethodPost, path: "/api/v1alpha1/auth/logout"})
				Expect(res).Should(HaveHTTPStatus(http.StatusOK))
				session = res.Cookies()
			}
			By("---------------test start----------------")
			res, body := test_HttpSend(session, request{method: http.MethodPost, path: "/api/v1alpha1/auth/verify"})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(replace(string(body), `"expireAt":".*"`, `"expireAt":"9999-99-99T99:99:99.99999999+9:00"`)).To(MatchSnapShot())
			Expect(res.Cookies()).ShouldNot(BeNil())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "logined session"),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "logouted session"),
			Entry(nil, "nil session"),
		)
	})

	//==================================================================================
	Describe("[Logout]", func() {

		run_test := func(sessionType string) {
			var session []*http.Cookie = nil
			switch sessionType {
			case "logined session":
				session = test_Login("usertest", "password1")
			case "logouted session":
				session = test_Login("usertest", "password1")
				res, _ := test_HttpSend(session, request{method: http.MethodPost, path: "/api/v1alpha1/auth/logout"})
				Expect(res).Should(HaveHTTPStatus(http.StatusOK))
				session = res.Cookies()
			}
			By("---------------test start----------------")
			res, body := test_HttpSend(session, request{method: http.MethodPost, path: "/api/v1alpha1/auth/logout"})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())
			Expect(res.Cookies()).ShouldNot(BeNil())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "logined session"),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "logouted session"),
			Entry(nil, "nil session"),
		)
	})
})
