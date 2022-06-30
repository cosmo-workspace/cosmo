package dashboard

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
)

var _ = Describe("Dashboard server [Workspace]", func() {

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession("usertest", "user", "", "password")
		adminSession = test_CreateLoginUserSession("usertest-admin", "admin", wsv1alpha1.UserAdminRole, "password")
		test_CreateTemplate(wsv1alpha1.TemplateTypeWorkspace, "template1")
	})

	AfterEach(func() {
		clientMock.Clear()
		test_DeleteWorkspaceAll()
		test_DeleteCosmoUserAll()
		test_DeleteTemplateAll()
	})

	//==================================================================================
	workspaceSnap := func(ws *wsv1alpha1.Workspace) struct{ Name, Namespace, Spec, Status interface{} } {
		return struct{ Name, Namespace, Spec, Status interface{} }{
			Name:      ws.Name,
			Namespace: ws.Namespace,
			Spec:      ws.Spec,
			Status:    ws.Status,
		}
	}
	//==================================================================================
	Describe("authorization by role", func() {

		DescribeTable("access API with admin user session:",
			func(stat int, req request) {
				test_CreateWorkspace("usertest", "ws1", "template1", map[string]string{})
				test_createNetworkRule("usertest", "ws1", "nw1", 9999, "gp1", "/")
				test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})
				test_createNetworkRule("usertest-admin", "ws1", "nw1", 9999, "gp1", "/")
				By("---------------test start----------------")
				res, _ := test_HttpSend(adminSession, req)
				Ω(res.StatusCode).Should(Equal(stat))
				By("---------------test end---------------")
			},
			func(stat int, req request) string { return fmt.Sprintf("%d %+v", stat, req) },
			// update own resource
			Entry(nil, 201, request{method: http.MethodPost, path: "/api/v1alpha1/user/usertest-admin/workspace", body: `{"name": "ws2","template": "template1"}`}),
			Entry(nil, 200, request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin/workspace"}),
			Entry(nil, 200, request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1"}),
			Entry(nil, 200, request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1"}),
			Entry(nil, 200, request{method: http.MethodPatch, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1", body: `{"replicas": 0}`}),
			Entry(nil, 200, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw2", body: `{"portNumber": 3000,"group": "gp2","httpPath": "/"}`}),
			Entry(nil, 200, request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw1"}),
			// update resource of others
			Entry(nil, 201, request{method: http.MethodPost, path: "/api/v1alpha1/user/usertest/workspace", body: `{"name": "ws2","template": "template1"}`}),
			Entry(nil, 200, request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest/workspace"}),
			Entry(nil, 200, request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest/workspace/ws1"}),
			Entry(nil, 200, request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest/workspace/ws1"}),
			Entry(nil, 200, request{method: http.MethodPatch, path: "/api/v1alpha1/user/usertest/workspace/ws1", body: `{"replicas": 2}`}),
			Entry(nil, 200, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/workspace/ws1/network/nw2", body: `{"portNumber": 3000,"group": "gp2","httpPath": "/"}`}),
			Entry(nil, 200, request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest/workspace/ws1/network/nw1"}),
		)

		DescribeTable("access API with normal user session:",
			func(stat int, req request) {
				test_CreateWorkspace("usertest", "ws1", "template1", map[string]string{})
				test_createNetworkRule("usertest", "ws1", "nw1", 9999, "gp1", "/")
				test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})
				test_createNetworkRule("usertest-admin", "ws1", "nw1", 9999, "gp1", "/")
				By("---------------test start----------------")
				res, _ := test_HttpSend(userSession, req)
				Ω(res.StatusCode).Should(Equal(stat))
				By("---------------test end---------------")
			},
			func(stat int, req request) string { return fmt.Sprintf("%d %+v", stat, req) },
			// update own resource
			Entry(nil, 201, request{method: http.MethodPost, path: "/api/v1alpha1/user/usertest/workspace", body: `{"name": "ws2","template": "template1"}`}),
			Entry(nil, 200, request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest/workspace"}),
			Entry(nil, 200, request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest/workspace/ws1"}),
			Entry(nil, 200, request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest/workspace/ws1"}),
			Entry(nil, 200, request{method: http.MethodPatch, path: "/api/v1alpha1/user/usertest/workspace/ws1", body: `{"replicas": 0}`}),
			Entry(nil, 200, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest/workspace/ws1/network/nw2", body: `{"portNumber": 3000,"group": "gp2","httpPath": "/"}`}),
			Entry(nil, 200, request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest/workspace/ws1/network/nw1"}),
			// update resource of others
			Entry(nil, 403, request{method: http.MethodPost, path: "/api/v1alpha1/user/usertest-admin/workspace", body: `{"name": "ws2","template": "template1"}`}),
			Entry(nil, 403, request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin/workspace"}),
			Entry(nil, 403, request{method: http.MethodGet, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1"}),
			Entry(nil, 403, request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1"}),
			Entry(nil, 403, request{method: http.MethodPatch, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1", body: `{"replicas": 1}`}),
			Entry(nil, 403, request{method: http.MethodPut, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw2", body: `{"portNumber": 3000,"group": "gp2","httpPath": "/"}`}),
			Entry(nil, 403, request{method: http.MethodDelete, path: "/api/v1alpha1/user/usertest-admin/workspace/ws1/network/nw1"}),
		)
	})

	//==================================================================================
	Describe("[PostWorkspace]", func() {

		run_test := func(userId, wsName, requestBody string) {
			test_CreateWorkspace("usertest-admin", "existing-ws", "template1", nil)
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodPost, path: fmt.Sprintf("/api/v1alpha1/user/%s/workspace", userId), body: requestBody})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())

			if wsName != "" {
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), wsName, userId)
				if res.StatusCode == http.StatusCreated {
					Expect(err).NotTo(HaveOccurred())
					Ω(workspaceSnap(wsv1Workspace)).To(MatchSnapShot())
				} else {
					Expect(err).To(HaveOccurred())
				}
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "usertest-admin", "ws1", `{"name": "ws1","template": "template1","vars": { "HOGE": "HOGEHOGE"}}`),
			Entry(nil, "usertest-admin", "ws1", `{"name": "ws1","template": "template1"}`),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "xxxxx", "ws1", `{"name": "ws1","template": "template1"}`),
			Entry(nil, "usertest-admin", "", `{"name": "","template": "template1"}`),
			Entry(nil, "usertest-admin", "ws1", `{"name": "ws1","template": ""}`),
			Entry(nil, "usertest-admin", "XXXX", `{"name": "XXXX","template": "template1"}`),
			Entry(nil, "usertest-admin", "ws1", `{"name": "ws1","template": "XXX"}`),
			Entry(nil, "usertest-admin", "", `{"name": "existing-ws","template": "template1"}`),
		)
	})

	//==================================================================================
	Describe("[GetWorkspaces]", func() {

		run_test := func(userId string) {
			test_CreateWorkspace("usertest-admin", "ws1", "template1", nil)
			test_CreateWorkspace("usertest-admin", "ws2", "template1", nil)
			test_createNetworkRule("usertest-admin", "ws2", "nw1", 1111, "gp1", "/")
			test_createNetworkRule("usertest-admin", "ws2", "nw3", 2222, "gp1", "/")
			test_createNetworkRule("usertest-admin", "ws2", "nw2", 3333, "gp1", "/")
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodGet, path: fmt.Sprintf("/api/v1alpha1/user/%s/workspace", userId)})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "usertest-admin"),
			Entry(nil, "usertest"),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "xxxxx"),
		)

		DescribeTable("❌ fail with unexpected error:",
			func(userId string) {
				clientMock.SetListError((*Server).GetWorkspaces, errors.New("mock get list error"))
				run_test(userId)
			},
			Entry(nil, "usertest-admin"),
		)
	})

	//==================================================================================
	Describe("[GetWorkspace]", func() {

		run_test := func(userId, wsName string) {
			test_CreateWorkspace("usertest", "ws1", "template1", map[string]string{"HOGE": "HOGEHOGE"})
			test_createNetworkRule("usertest", "ws1", "main", 18080, "mainnw", "/")
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodGet, path: fmt.Sprintf("/api/v1alpha1/user/%s/workspace/%s", userId, wsName)})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "usertest", "ws1"),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "xxxxx", "ws1"),
			Entry(nil, "usertest-admin", "xxx"),
		)

		DescribeTable("❌ fail with an unexpected error at list:",
			func(userId, wsName string) {
				clientMock.GetMock = func(ctx context.Context, key client.ObjectKey, obj client.Object) (mocked bool, err error) {
					if key.Name == wsName {
						return true, errors.New("mock get workspace error")
					}
					return false, nil
				}
				//clientMock.SetGetError(`\.GetWorkspace$`, errors.New("mock get workspace error"))
				//clientMock.SetGetError((*Server).GetWorkspace, errors.New("mock get workspace error"))
				run_test(userId, wsName)
			},
			Entry(nil, "usertest", "ws1"),
		)
	})

	//==================================================================================
	Describe("[DeleteWorkspace]", func() {

		run_test := func(userId, wsName string) {
			test_CreateWorkspace("usertest", "ws1", "template1", map[string]string{"HOGE": "HOGEHOGE"})
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodDelete, path: fmt.Sprintf("/api/v1alpha1/user/%s/workspace/%s", userId, wsName)})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())

			_, err := k8sClient.GetWorkspaceByUserID(context.Background(), "ws1", "usertest")
			if res.StatusCode == http.StatusOK {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "usertest", "ws1"),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "xxxxx", "ws1"),
			Entry(nil, "usertest-admin", "xxx"),
		)

		DescribeTable("❌ fail with an unexpected error at delete:",
			func(userId, wsName string) {
				clientMock.SetDeleteError((*Server).DeleteWorkspace, errors.New("mock delete workspace error"))
				run_test(userId, wsName)
			},
			Entry(nil, "usertest", "ws1"),
		)
	})

	//==================================================================================
	Describe("[PatchWorkspace]", func() {

		run_test := func(userId, wsName, requestBody string) {
			test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodPatch, path: fmt.Sprintf("/api/v1alpha1/user/%s/workspace/%s", userId, wsName), body: requestBody})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())

			if res.StatusCode == http.StatusOK {
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), wsName, userId)
				Expect(err).NotTo(HaveOccurred())
				Ω(workspaceSnap(wsv1Workspace)).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "usertest-admin", "ws1", `{"replicas": 0}`),
			Entry(nil, "usertest-admin", "ws1", `{"replicas": 5}`),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "xxxxx", "ws1", `{"replicas": 0}`),
			Entry(nil, "usertest", "xxx", `{"replicas": 1}`),
			Entry(nil, "usertest-admin", "ws1", `{"replicas": 1}`),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(userId, wsName, requestBody string) {
				clientMock.SetUpdateError((*Server).PatchWorkspace, errors.New("mock update workspace error"))
				run_test(userId, wsName, requestBody)
			},
			Entry(nil, "usertest-admin", "ws1", `{"replicas": 0}`),
		)
	})

	//==================================================================================
	Describe("[PutNetworkRule]", func() {

		run_test := func(userId, wsName, nw, requestBody string) {
			test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})
			test_createNetworkRule("usertest-admin", "ws1", "nw1", 9999, "gp1", "/")
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodPut, path: fmt.Sprintf("/api/v1alpha1/user/%s/workspace/%s/network/%s", userId, wsName, nw), body: requestBody})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())

			if res.StatusCode == http.StatusOK {
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), wsName, userId)
				Expect(err).NotTo(HaveOccurred())
				Ω(workspaceSnap(wsv1Workspace)).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "usertest-admin", "ws1", "nw2", `{"portNumber": 3000,"group": "gp2","httpPath": "/","public":false}`),
			Entry(nil, "usertest-admin", "ws1", "nw2", `{"portNumber": 3000,"public":true}`),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "xxxxx", "ws1", "nw2", `{"portNumber": 3000,"group": "gp2","httpPath": "/","public":false}`),
			Entry(nil, "usertest-admin", "xxx", "nw2", `{"portNumber": 3000,"group": "gp2","httpPath": "/","public":false}`),
			Entry(nil, "usertest-admin", "ws1", "nw1", `{"portNumber": 9999,"group": "gp1","httpPath": "/","public":false}`),
			Entry(nil, "usertest-admin", "ws1", "nw9", `{"portNumber": 9999,"group": "gp1","httpPath": "/","public":false}`),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(userId, wsName, nw, requestBody string) {
				clientMock.SetUpdateError((*Server).PutNetworkRule, errors.New("mock update networkrule error"))
				run_test(userId, wsName, nw, requestBody)
			},
			Entry(nil, "usertest-admin", "ws1", "nw2", `{"portNumber": 3000,"group": "gp2","httpPath": "/","public":false}`),
		)
	})

	//==================================================================================
	Describe("[DeleteNetworkRule]", func() {

		run_test := func(userId, wsName, nw string) {
			test_CreateWorkspace("usertest-admin", "ws1", "template1", map[string]string{})
			test_createNetworkRule("usertest-admin", "ws1", "nw1", 9999, "gp1", "/")
			test_createNetworkRule("usertest-admin", "ws1", "main", 18080, "main", "/")
			By("---------------test start----------------")
			res, body := test_HttpSend(adminSession, request{method: http.MethodDelete, path: fmt.Sprintf("/api/v1alpha1/user/%s/workspace/%s/network/%s", userId, wsName, nw)})
			Ω(res.StatusCode).To(MatchSnapShot())
			Ω(string(body)).To(MatchSnapShot())

			if res.StatusCode == http.StatusOK {
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), wsName, userId)
				Expect(err).NotTo(HaveOccurred())
				Ω(workspaceSnap(wsv1Workspace)).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "usertest-admin", "ws1", "nw1"),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "xxxxx", "ws1", "nw2"),
			Entry(nil, "usertest-admin", "xxx", "nw2"),
			Entry(nil, "usertest-admin", "ws1", "xxx"),
			Entry(nil, "usertest-admin", "ws1", "main"),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(userId, wsName, nw string) {
				clientMock.SetUpdateError((*Server).DeleteNetworkRule, errors.New("mock delete network rule error"))
				run_test(userId, wsName, nw)
			},
			Entry(nil, "usertest-admin", "ws1", "nw1"),
		)
	})

})
