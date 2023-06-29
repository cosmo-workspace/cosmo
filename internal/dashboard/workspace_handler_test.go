package dashboard

import (
	"context"
	"errors"
	"net/http"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/utils/pointer"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

var _ = Describe("Dashboard server [Workspace]", func() {

	var (
		userSession  string
		adminSession string
		client       dashboardv1alpha1connect.WorkspaceServiceClient
	)

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession("normal-user", "user", nil, "password")
		adminSession = test_CreateLoginUserSession("admin-user", "admin", []cosmov1alpha1.UserRole{cosmov1alpha1.PrivilegedRole}, "password")
		testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeWorkspace, "template1")
		client = dashboardv1alpha1connect.NewWorkspaceServiceClient(http.DefaultClient, "http://localhost:8888")
	})

	AfterEach(func() {
		clientMock.Clear()
		testUtil.DeleteWorkspaceAll()
		testUtil.DeleteCosmoUserAll()
		testUtil.DeleteTemplateAll()
	})

	//==================================================================================
	getSession := func(loginUser string) string {
		if loginUser == "admin-user" {
			return adminSession
		} else {
			return userSession
		}
	}

	//==================================================================================
	Describe("[CreateWorkspace]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.CreateWorkspaceRequest) {
			testUtil.CreateWorkspace("admin-user", "existing-ws", "template1", nil)
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.CreateWorkspace(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserName(context.Background(), req.WsName, req.UserName)
				Expect(err).NotTo(HaveOccurred())
				Ω(ObjectSnapshot(wsv1Workspace)).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("admin user with vars", "admin-user", &dashv1alpha1.CreateWorkspaceRequest{UserName: "admin-user", WsName: "ws1", Template: "template1", Vars: map[string]string{"HOGE": "HOGEHOGE"}}),
			Entry("admin user without vars", "admin-user", &dashv1alpha1.CreateWorkspaceRequest{UserName: "admin-user", WsName: "ws1", Template: "template1"}),
			Entry("normal user without vars", "normal-user", &dashv1alpha1.CreateWorkspaceRequest{UserName: "normal-user", WsName: "ws1", Template: "template1"}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("invalid username", "admin-user", &dashv1alpha1.CreateWorkspaceRequest{UserName: "xxxxx", WsName: "ws1", Template: "template1"}),
			Entry("empty workspace name", "admin-user", &dashv1alpha1.CreateWorkspaceRequest{UserName: "admin-user", WsName: "", Template: "template1"}),
			Entry("empty template", "admin-user", &dashv1alpha1.CreateWorkspaceRequest{UserName: "admin-user", WsName: "ws1", Template: ""}),
			Entry("invalid workspace name", "admin-user", &dashv1alpha1.CreateWorkspaceRequest{UserName: "admin-user", WsName: "XXXX", Template: "template1"}),
			Entry("invalid template", "admin-user", &dashv1alpha1.CreateWorkspaceRequest{UserName: "admin-user", WsName: "ws1", Template: "XXX"}),
			Entry("workspace already exists", "admin-user", &dashv1alpha1.CreateWorkspaceRequest{UserName: "admin-user", WsName: "existing-ws", Template: "template1"}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry("normal user cannot create admin user's workspace", "normal-user", &dashv1alpha1.CreateWorkspaceRequest{UserName: "admin-user", WsName: "ws1", Template: "template1"}),
		)
	})

	//==================================================================================
	Describe("[GetWorkspaces]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.GetWorkspacesRequest) {
			testUtil.CreateWorkspace("admin-user", "ws1", "template1", nil)
			testUtil.CreateWorkspace("admin-user", "ws2", "template1", nil)
			testUtil.UpsertNetworkRule("admin-user", "ws2", "nw1", 1111, "/", false, -1)
			testUtil.UpsertNetworkRule("admin-user", "ws2", "nw3", 2222, "/", false, -1)
			testUtil.UpsertNetworkRule("admin-user", "ws2", "nw2", 3333, "/", false, -1)
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.GetWorkspaces(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("admin user can get own workspaces", "admin-user", &dashv1alpha1.GetWorkspacesRequest{UserName: "admin-user"}),
			Entry("admin user can get normal user's workspaces", "admin-user", &dashv1alpha1.GetWorkspacesRequest{UserName: "normal-user"}),
			Entry("normal user can get own workspaces", "normal-user", &dashv1alpha1.GetWorkspacesRequest{UserName: "normal-user"}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("invalid user name", "admin-user", &dashv1alpha1.GetWorkspacesRequest{UserName: "xxxxx"}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry("normal user cannot get admin's workspaces", "normal-user", &dashv1alpha1.GetWorkspacesRequest{UserName: "admin-user"}),
		)

		DescribeTable("❌ fail with unexpected error:",
			func(loginUser string, req *dashv1alpha1.GetWorkspacesRequest) {
				clientMock.SetListError((*Server).GetWorkspaces, errors.New("mock get list error"))
				run_test(loginUser, req)
			},
			Entry("unexpected err", "admin-user", &dashv1alpha1.GetWorkspacesRequest{UserName: "admin-user"}),
		)
	})

	//==================================================================================
	Describe("[GetWorkspace]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.GetWorkspaceRequest) {
			testUtil.CreateWorkspace("admin-user", "ws1", "template1", nil)
			testUtil.CreateWorkspace("normal-user", "ws1", "template1", map[string]string{"HOGE": "HOGEHOGE"})
			testUtil.UpsertNetworkRule("normal-user", "ws1", "add", 18080, "/", false, -1)
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.GetWorkspace(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("admin user can get normal user's workspace", "admin-user", &dashv1alpha1.GetWorkspaceRequest{UserName: "normal-user", WsName: "ws1"}),
			Entry("normal user can get own workspace", "normal-user", &dashv1alpha1.GetWorkspaceRequest{UserName: "normal-user", WsName: "ws1"}),
			Entry("admin user can get own workspace", "admin-user", &dashv1alpha1.GetWorkspaceRequest{UserName: "admin-user", WsName: "ws1"}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("invalid user name", "admin-user", &dashv1alpha1.GetWorkspaceRequest{UserName: "xxxxx", WsName: "ws1"}),
			Entry("invalid workspace name", "admin-user", &dashv1alpha1.GetWorkspaceRequest{UserName: "admin-user", WsName: "xxx"}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry("normal user cannot get admin's workspace", "normal-user", &dashv1alpha1.GetWorkspaceRequest{UserName: "admin-user", WsName: "ws1"}),
		)

		DescribeTable("❌ fail with an unexpected error at list:",
			func(loginUser string, req *dashv1alpha1.GetWorkspaceRequest) {
				clientMock.SetGetError((*Server).GetWorkspace, errors.New("mock get workspace error"))
				run_test(loginUser, req)
			},
			Entry("unexpected err", "admin-user", &dashv1alpha1.GetWorkspaceRequest{UserName: "normal-user", WsName: "ws1"}),
		)
	})

	//==================================================================================
	Describe("[DeleteWorkspace]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.DeleteWorkspaceRequest) {
			testUtil.CreateWorkspace("normal-user", "ws1", "template1", map[string]string{"HOGE": "HOGEHOGE"})
			testUtil.CreateWorkspace("admin-user", "ws1", "template1", map[string]string{"HOGE": "HOGEHOGE"})
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.DeleteWorkspace(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				_, err := k8sClient.GetWorkspaceByUserName(context.Background(), req.WsName, req.UserName)
				Expect(err).To(HaveOccurred())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("admin user can delete normal user's workspace", "admin-user", &dashv1alpha1.DeleteWorkspaceRequest{UserName: "normal-user", WsName: "ws1"}),
			Entry("normal user can delete own workspace", "normal-user", &dashv1alpha1.DeleteWorkspaceRequest{UserName: "normal-user", WsName: "ws1"}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("invalid user name", "admin-user", &dashv1alpha1.DeleteWorkspaceRequest{UserName: "xxxxx", WsName: "ws1"}),
			Entry("invalid workspace name", "admin-user", &dashv1alpha1.DeleteWorkspaceRequest{UserName: "admin-user", WsName: "xxx"}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry("normal user cannot delete admin's workspace", "normal-user", &dashv1alpha1.DeleteWorkspaceRequest{UserName: "admin-user", WsName: "ws1"}),
		)

		DescribeTable("❌ fail with an unexpected error at delete:",
			func(loginUser string, req *dashv1alpha1.DeleteWorkspaceRequest) {
				clientMock.SetDeleteError((*Server).DeleteWorkspace, errors.New("mock delete workspace error"))
				run_test(loginUser, req)
			},
			Entry("unexpected err", "admin-user", &dashv1alpha1.DeleteWorkspaceRequest{UserName: "normal-user", WsName: "ws1"}),
		)
	})

	//==================================================================================
	Describe("[UpdateWorkspace]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.UpdateWorkspaceRequest) {
			testUtil.CreateWorkspace("admin-user", "ws1", "template1", map[string]string{})
			testUtil.CreateWorkspace("normal-user", "ws1", "template1", map[string]string{})
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.UpdateWorkspace(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserName(context.Background(), req.WsName, req.UserName)
				Expect(err).NotTo(HaveOccurred())
				Ω(ObjectSnapshot(wsv1Workspace)).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("admin user can update own workspace's replica", "admin-user", &dashv1alpha1.UpdateWorkspaceRequest{UserName: "admin-user", WsName: "ws1", Replicas: pointer.Int64(0)}),
			Entry("admin user can update own workspace with no change", "admin-user", &dashv1alpha1.UpdateWorkspaceRequest{UserName: "admin-user", WsName: "ws1"}),
			Entry("normal user can update own workspace's replica", "normal-user", &dashv1alpha1.UpdateWorkspaceRequest{UserName: "normal-user", WsName: "ws1", Replicas: pointer.Int64(5)}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("invalid user name", "admin-user", &dashv1alpha1.UpdateWorkspaceRequest{UserName: "xxxxx", WsName: "ws1", Replicas: pointer.Int64(0)}),
			Entry("invalid workspace name", "admin-user", &dashv1alpha1.UpdateWorkspaceRequest{UserName: "normal-user", WsName: "xxx", Replicas: pointer.Int64(1)}),
			Entry("no change", "admin-user", &dashv1alpha1.UpdateWorkspaceRequest{UserName: "admin-user", WsName: "ws1", Replicas: pointer.Int64(1)}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry("normal user cannot update admin's workspace", "normal-user", &dashv1alpha1.UpdateWorkspaceRequest{UserName: "admin-user", WsName: "ws1", Replicas: pointer.Int64(0)}),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(loginUser string, req *dashv1alpha1.UpdateWorkspaceRequest) {
				clientMock.SetUpdateError((*Server).UpdateWorkspace, errors.New("mock update workspace error"))
				run_test(loginUser, req)
			},
			Entry("unexpected err", "admin-user", &dashv1alpha1.UpdateWorkspaceRequest{UserName: "admin-user", WsName: "ws1", Replicas: pointer.Int64(0)}),
		)
	})

	//==================================================================================
	Describe("[UpsertNetworkRule]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.UpsertNetworkRuleRequest) {
			testUtil.CreateWorkspace("admin-user", "ws1", "template1", map[string]string{})
			testUtil.UpsertNetworkRule("admin-user", "ws1", "nw1", 9999, "/", false, -1)
			testUtil.CreateWorkspace("normal-user", "ws1", "template1", map[string]string{})
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.UpsertNetworkRule(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserName(context.Background(), req.WsName, req.UserName)
				Expect(err).NotTo(HaveOccurred())
				Ω(ObjectSnapshot(wsv1Workspace)).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("insert with only port", "admin-user", &dashv1alpha1.UpsertNetworkRuleRequest{UserName: "admin-user", WsName: "ws1", Index: -1,
				NetworkRule: &dashv1alpha1.NetworkRule{PortNumber: 3000}}),
			Entry("update existing record: public", "admin-user", &dashv1alpha1.UpsertNetworkRuleRequest{UserName: "admin-user", WsName: "ws1", Index: 1,
				NetworkRule: &dashv1alpha1.NetworkRule{PortNumber: 9999, CustomHostPrefix: "nw1", HttpPath: "/", Public: true}}),
			Entry("update existing record: port number", "admin-user", &dashv1alpha1.UpsertNetworkRuleRequest{UserName: "admin-user", WsName: "ws1", Index: 1,
				NetworkRule: &dashv1alpha1.NetworkRule{PortNumber: 3000, CustomHostPrefix: "nw1", HttpPath: "/", Public: true}}),
			Entry("insert the same excluding path", "admin-user", &dashv1alpha1.UpsertNetworkRuleRequest{UserName: "admin-user", WsName: "ws1", Index: -1,
				NetworkRule: &dashv1alpha1.NetworkRule{PortNumber: 3000, CustomHostPrefix: "nw1", HttpPath: "/dev", Public: true}}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("invalid user name", "admin-user", &dashv1alpha1.UpsertNetworkRuleRequest{UserName: "xxxxx", WsName: "ws1", Index: -1,
				NetworkRule: &dashv1alpha1.NetworkRule{PortNumber: 3001}}),
			Entry("invalid workspace name", "admin-user", &dashv1alpha1.UpsertNetworkRuleRequest{UserName: "admin-user", WsName: "xxx", Index: -1,
				NetworkRule: &dashv1alpha1.NetworkRule{PortNumber: 3002}}),
			Entry("duplicate ports: insert", "admin-user", &dashv1alpha1.UpsertNetworkRuleRequest{UserName: "admin-user", WsName: "ws1", Index: -1,
				NetworkRule: &dashv1alpha1.NetworkRule{PortNumber: 3003, CustomHostPrefix: "main", HttpPath: "/", Public: false}}),
			Entry("duplicate ports: update", "admin-user", &dashv1alpha1.UpsertNetworkRuleRequest{UserName: "admin-user", WsName: "ws1", Index: 0,
				NetworkRule: &dashv1alpha1.NetworkRule{PortNumber: 3004, CustomHostPrefix: "nw1", HttpPath: "/", Public: false}}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry(nil, "normal-user", &dashv1alpha1.UpsertNetworkRuleRequest{UserName: "admin-user", WsName: "ws1",
				NetworkRule: &dashv1alpha1.NetworkRule{PortNumber: 2000}}),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(loginUser string, req *dashv1alpha1.UpsertNetworkRuleRequest) {
				clientMock.SetUpdateError((*Server).UpsertNetworkRule, errors.New("mock update networkrule error"))
				run_test(loginUser, req)
			},
			Entry(nil, "admin-user", &dashv1alpha1.UpsertNetworkRuleRequest{UserName: "admin-user", WsName: "ws1",
				NetworkRule: &dashv1alpha1.NetworkRule{PortNumber: 2001}}),
		)
	})

	//==================================================================================
	Describe("[DeleteNetworkRule]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.DeleteNetworkRuleRequest) {
			testUtil.CreateWorkspace("normal-user", "ws1", "template1", map[string]string{})
			testUtil.UpsertNetworkRule("normal-user", "ws1", "nw1", 9999, "/", false, -1)
			testUtil.CreateWorkspace("admin-user", "ws1", "template1", map[string]string{})
			testUtil.UpsertNetworkRule("admin-user", "ws1", "nw1", 9999, "/", false, -1)
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.DeleteNetworkRule(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserName(context.Background(), req.WsName, req.UserName)
				Expect(err).NotTo(HaveOccurred())
				Ω(ObjectSnapshot(wsv1Workspace)).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "admin-user", &dashv1alpha1.DeleteNetworkRuleRequest{UserName: "admin-user", WsName: "ws1", Index: 1}),
			Entry(nil, "normal-user", &dashv1alpha1.DeleteNetworkRuleRequest{UserName: "normal-user", WsName: "ws1", Index: 1}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "admin-user", &dashv1alpha1.DeleteNetworkRuleRequest{UserName: "xxxxx", WsName: "ws1", Index: 1}),
			Entry(nil, "admin-user", &dashv1alpha1.DeleteNetworkRuleRequest{UserName: "admin-user", WsName: "xxx", Index: 1}),
			Entry(nil, "admin-user", &dashv1alpha1.DeleteNetworkRuleRequest{UserName: "admin-user", WsName: "ws1", Index: -1}),
			Entry(nil, "admin-user", &dashv1alpha1.DeleteNetworkRuleRequest{UserName: "admin-user", WsName: "ws1", Index: 2}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry(nil, "normal-user", &dashv1alpha1.DeleteNetworkRuleRequest{UserName: "admin-user", WsName: "ws1", Index: 1}),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(loginUser string, req *dashv1alpha1.DeleteNetworkRuleRequest) {
				clientMock.SetUpdateError((*Server).DeleteNetworkRule, errors.New("mock delete network rule error"))
				run_test(loginUser, req)
			},
			Entry(nil, "admin-user", &dashv1alpha1.DeleteNetworkRuleRequest{UserName: "admin-user", WsName: "ws1", Index: 1}),
		)
	})

})
