package dashboard

import (
	"context"
	"net/http"

	"github.com/bufbuild/connect-go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ = Describe("Dashboard server [auth]", func() {

	var client dashboardv1alpha1connect.AuthServiceClient

	BeforeEach(func() {
		testUtil.CreateLoginUser("normal-user", "user", "", "password1")
		testUtil.CreateLoginUser("admin-user", "admin", wsv1alpha1.UserAdminRole, "password2")
		client = dashboardv1alpha1connect.NewAuthServiceClient(http.DefaultClient, "http://localhost:8888")
	})

	AfterEach(func() {
		clientMock.Clear()
		testUtil.DeleteCosmoUserAll()
	})

	//==================================================================================

	//==================================================================================
	Describe("[Login]", func() {

		run_test := func(req *dashv1alpha1.LoginRequest) {
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.Login(ctx, connect.NewRequest(req))
			if err == nil {
				Expect(res.Msg.ExpireAt).ShouldNot(BeNil())
				res.Msg.ExpireAt = &timestamppb.Timestamp{}
				Ω(res.Msg).To(MatchSnapShot())
				Expect(res.Header().Get("Set-Cookie")).ShouldNot(BeNil())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, &dashv1alpha1.LoginRequest{UserName: "normal-user", Password: "password1"}),
			Entry(nil, &dashv1alpha1.LoginRequest{UserName: "admin-user", Password: "password2"}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, &dashv1alpha1.LoginRequest{Password: "password1"}),
			Entry(nil, &dashv1alpha1.LoginRequest{UserName: "normal-user"}),
			Entry(nil, &dashv1alpha1.LoginRequest{UserName: "normal-user", Password: "invalid"}),
			Entry(nil, &dashv1alpha1.LoginRequest{UserName: "xxxxxxx", Password: "password1"}),
		)
	})

	//==================================================================================
	Describe("[Verify]", func() {

		run_test := func(sessionType string) {
			ctx := context.Background()
			var session string
			switch sessionType {
			case "nil session":
			case "logined session":
				session = test_Login("normal-user", "password1")
			case "logouted session":
				session = test_Login("normal-user", "password1")
				logoutResp, _ := client.Logout(ctx, NewRequestWithSession(&emptypb.Empty{}, session))
				session = logoutResp.Header().Get("Set-Cookie")
			}
			By("---------------test start----------------")
			res, err := client.Verify(ctx, NewRequestWithSession(&emptypb.Empty{}, session))
			if err == nil {
				Expect(res.Msg.ExpireAt).ShouldNot(BeNil())
				res.Msg.ExpireAt = &timestamppb.Timestamp{}
				Ω(res.Msg).To(MatchSnapShot())
				Expect(res.Header().Get("Set-Cookie")).Should(BeEmpty())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
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
			ctx := context.Background()
			var session string
			switch sessionType {
			case "nil session":
			case "logined session":
				session = test_Login("normal-user", "password1")
			case "logouted session":
				session = test_Login("normal-user", "password1")
				logoutResp, _ := client.Logout(ctx, NewRequestWithSession(&emptypb.Empty{}, session))
				session = logoutResp.Header().Get("Set-Cookie")
			}
			By("---------------test start----------------")
			res, err := client.Logout(ctx, NewRequestWithSession(&emptypb.Empty{}, session))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				Expect(res.Header().Get("Set-Cookie")).Should(BeEmpty())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
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
