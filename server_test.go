package dashboard

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type request struct {
	method string
	path   string
	query  map[string]string
	body   string
}
type response struct {
	statusCode int
	body       string
}

var _ = Describe("Dashboard server", func() {
	Context("when access login API with invalid user authentication", func() {
		It("should deny with 403", func() {
			req, err := http.NewRequest(http.MethodPost, "http://localhost:8888/api/v1alpha1/auth/login", bytes.NewBufferString(`{"id": "usertest", "password": "invalid"}`))
			Expect(err).NotTo(HaveOccurred())

			got, err := http.DefaultClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer got.Body.Close()

			Expect(got.StatusCode).Should(Equal(http.StatusForbidden))
		})
	})

	Context("when access login API with valid user authentication", func() {
		It("should success and response with session cookie", func() {
			req, err := http.NewRequest(http.MethodPost, "http://localhost:8888/api/v1alpha1/auth/login", bytes.NewBufferString(`{"id": "usertest", "password": "password"}`))
			Expect(err).NotTo(HaveOccurred())

			got, err := http.DefaultClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer got.Body.Close()

			Expect(got.StatusCode).Should(Equal(http.StatusOK))

			userSession = got.Cookies()
		})
	})

	Context("when access login API with valid admin authentication", func() {
		It("should success and response with session cookie", func() {
			req, err := http.NewRequest(http.MethodPost, "http://localhost:8888/api/v1alpha1/auth/login", bytes.NewBufferString(`{"id": "usertest-admin", "password": "password"}`))
			Expect(err).NotTo(HaveOccurred())

			got, err := http.DefaultClient.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer got.Body.Close()

			Expect(got.StatusCode).Should(Equal(http.StatusOK))

			adminSession = got.Cookies()
		})
	})

	Context("when access API with valid user session", func() {
		tests := []struct {
			name string
			req  request
			want response
		}{
			{
				name: "get user",
				req: request{
					method: http.MethodGet,
					path:   "/api/v1alpha1/user/usertest",
				},
				want: response{
					statusCode: 200,
					body:       `{"user":{"id":"usertest","displayName":"お名前","authType":"kosmo-secret"}}`,
				},
			},
		}
		for _, tt := range tests {
			It(tt.name, func() {
				req, err := http.NewRequest(tt.req.method, "http://localhost:8888"+tt.req.path, bytes.NewBufferString(tt.req.body))
				Expect(err).NotTo(HaveOccurred())

				for _, v := range userSession {
					req.AddCookie(v)
				}

				got, err := http.DefaultClient.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer got.Body.Close()

				Expect(got.StatusCode).Should(Equal(tt.want.statusCode))

				gotBody, err := ioutil.ReadAll(got.Body)
				Expect(err).NotTo(HaveOccurred())

				if tt.want.body != "" {
					Expect(gotBody).Should(Equal(append([]byte(tt.want.body), '\n')))
				}
			})
		}
	})

	Context("when access API with valid admin session", func() {
		tests := []struct {
			name string
			req  request
			want response
		}{
			{
				name: "get users",
				req: request{
					method: http.MethodGet,
					path:   "/api/v1alpha1/user",
				},
				want: response{
					statusCode: 200,
					body:       `{"items":[{"id":"usertest","displayName":"お名前","authType":"kosmo-secret"},{"id":"usertest-admin","displayName":"アドミン","role":"cosmo-admin","authType":"kosmo-secret"}]}`,
				},
			},
		}
		for _, tt := range tests {
			It(tt.name, func() {
				req, err := http.NewRequest(tt.req.method, "http://localhost:8888"+tt.req.path, bytes.NewBufferString(tt.req.body))
				Expect(err).NotTo(HaveOccurred())

				for _, v := range adminSession {
					req.AddCookie(v)
				}

				got, err := http.DefaultClient.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer got.Body.Close()

				Expect(got.StatusCode).Should(Equal(tt.want.statusCode))

				gotBody, err := ioutil.ReadAll(got.Body)
				Expect(err).NotTo(HaveOccurred())

				fmt.Println(string(gotBody))
				fmt.Println(tt.want.body)

				if tt.want.body != "" {
					Expect(gotBody).Should(Equal(append([]byte(tt.want.body), '\n')))
				}
			})
		}
	})
})
