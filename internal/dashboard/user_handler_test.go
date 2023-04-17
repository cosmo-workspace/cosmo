package dashboard

import (
	"context"
	"errors"
	"net/http"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"google.golang.org/protobuf/types/known/emptypb"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl_client "sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

var _ = Describe("Dashboard server [User]", func() {

	var (
		userSession  string
		adminSession string
		client       dashboardv1alpha1connect.UserServiceClient
	)

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession("normal-user", "お名前", nil, "password")
		adminSession = test_CreateLoginUserSession("admin-user", "アドミン", []cosmov1alpha1.UserRole{{Name: cosmov1alpha1.UserAdminRole}}, "password")
		client = dashboardv1alpha1connect.NewUserServiceClient(http.DefaultClient, "http://localhost:8888")
	})

	AfterEach(func() {
		clientMock.Clear()
		testUtil.DeleteCosmoUserAll()
		testUtil.DeleteTemplateAll()
	})

	//==================================================================================
	userSnap := func(us *cosmov1alpha1.User) struct{ Name, Namespace, Spec, Status interface{} } {
		return struct{ Name, Namespace, Spec, Status interface{} }{
			Name:      us.Name,
			Namespace: us.Namespace,
			Spec:      us.Spec,
			Status:    us.Status,
		}
	}

	getSession := func(loginUser string) string {
		if loginUser == "admin-user" {
			return adminSession
		} else {
			return userSession
		}
	}

	//==================================================================================
	Describe("[CreateUser]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.CreateUserRequest) {
			if req.UserName == "user-create" {
				testUtil.CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")
				testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeUserAddon, "user-temple1")
			} else if req.UserName == "user-create-later" {
				timer := time.AfterFunc(100*time.Millisecond, func() {
					testUtil.CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create-later")
				})
				defer timer.Stop()
			}
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.CreateUser(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg.User.DefaultPassword).ShouldNot(BeEmpty())
				res.Msg.User.DefaultPassword = "xxxxxxxx"
				Ω(res.Msg).To(MatchSnapShot())
				wsv1User, err := k8sClient.GetUser(context.Background(), req.UserName)
				Expect(err).NotTo(HaveOccurred())
				Ω(userSnap(wsv1User)).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success succeed in normal context:",
			run_test,
			Entry(nil, "admin-user", &dashv1alpha1.CreateUserRequest{
				UserName:    "user-create",
				DisplayName: "create 1",
				Roles:       []string{"cosmo-admin"},
				AuthType:    "kosmo-secret",
				Addons: []*dashv1alpha1.UserAddons{{
					Template: "user-temple1",
					Vars:     map[string]string{"HOGE": "FUGA"},
				}},
			}),
			Entry(nil, "admin-user", &dashv1alpha1.CreateUserRequest{UserName: "user-create"}),
			Entry(nil, "admin-user", &dashv1alpha1.CreateUserRequest{UserName: "user-create-later"}),
			Entry(nil, "admin-user", &dashv1alpha1.CreateUserRequest{UserName: "user-create-custom-role", Roles: []string{"team-a", "team-b"}}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "admin-user", &dashv1alpha1.CreateUserRequest{UserName: "user-create", AuthType: "xxxxxx"}),
			Entry(nil, "admin-user", &dashv1alpha1.CreateUserRequest{UserName: ""}),
			Entry(nil, "admin-user", &dashv1alpha1.CreateUserRequest{UserName: "user-createX"}),
			Entry(nil, "admin-user", &dashv1alpha1.CreateUserRequest{UserName: "normal-user"}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry(nil, "normal-user", &dashv1alpha1.CreateUserRequest{UserName: "user-create"}),
		)

		DescribeTable("❌ fail to create password timeout",
			run_test,
			Entry(nil, "admin-user", &dashv1alpha1.CreateUserRequest{UserName: "user-create-timeout"}),
		)
	})

	//==================================================================================
	Describe("[GetUsers]", func() {

		run_test := func(loginUser string) {
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.GetUsers(ctx, NewRequestWithSession(&emptypb.Empty{}, getSession(loginUser)))
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
			Entry(nil, "admin-user"),
		)

		DescribeTable("✅ success with empty user:",
			func(loginUser string) {
				clientMock.SetListError((*Server).GetUsers, nil)
				run_test(loginUser)
			},
			Entry(nil, "admin-user"),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry(nil, "normal-user"),
		)

		DescribeTable("❌ fail with an unexpected error at list:",
			func(loginUser string) {
				clientMock.SetListError((*Server).GetUsers, errors.New("mock user list error"))
				run_test(loginUser)
			},
			Entry(nil, "admin-user"),
		)
	})

	//==================================================================================
	Describe("[GetUser]", func() {

		run_test := func(loginUser string, username string) {
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.GetUser(ctx, NewRequestWithSession(&dashv1alpha1.GetUserRequest{UserName: username}, getSession(loginUser)))
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
			Entry(nil, "admin-user", "normal-user"),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "admin-user", "XXXXX"),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry(nil, "normal-user", "admin-user"),
		)

		DescribeTable("❌ fail with an unexpected error to get:",
			func(loginUser string, username string) {
				clientMock.SetGetError((*Server).GetUser, errors.New("get user error"))
				run_test(loginUser, username)
			},
			Entry(nil, "admin-user", "normal-user"),
		)
	})

	//==================================================================================
	Describe("[DeleteUser]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.DeleteUserRequest) {
			testUtil.CreateCosmoUser("user-delete1", "delete", nil)
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.DeleteUser(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				_, err = k8sClient.GetUser(context.Background(), req.UserName)
				Expect(err).To(HaveOccurred())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
				_, err = k8sClient.GetUser(context.Background(), "user-delete1")
				Expect(err).NotTo(HaveOccurred())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "admin-user", &dashv1alpha1.DeleteUserRequest{UserName: "user-delete1"}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "admin-user", &dashv1alpha1.DeleteUserRequest{UserName: "xxxxxx"}),
			Entry(nil, "admin-user", &dashv1alpha1.DeleteUserRequest{UserName: "admin-user"}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry(nil, "normal-user", &dashv1alpha1.DeleteUserRequest{UserName: "user-delete1"}),
		)

		DescribeTable("❌ fail with an unexpected error to get:",
			func(loginUser string, req *dashv1alpha1.DeleteUserRequest) {
				clientMock.SetGetError(`\.preFetchUserMiddleware\.|\.DeleteUser$`, errors.New("mock get user error")) ///
				//clientMock.SetGetError((*Server).DeleteUser, errors.New("mock get user error"))
				run_test(loginUser, req)
			},
			Entry(nil, "admin-user", &dashv1alpha1.DeleteUserRequest{UserName: "user-delete1"}),
		)

		DescribeTable("❌ fail with an unexpected error to delete:",
			func(loginUser string, req *dashv1alpha1.DeleteUserRequest) {
				clientMock.SetDeleteError((*Server).DeleteUser, errors.New("mock delete user error"))
				run_test(loginUser, req)
			},
			Entry(nil, "admin-user", &dashv1alpha1.DeleteUserRequest{UserName: "user-delete1"}),
		)
	})

	//==================================================================================
	Describe("[UpdateUserDisplayName]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.UpdateUserDisplayNameRequest) {
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.UpdateUserDisplayName(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				wsv1User, err := k8sClient.GetUser(context.Background(), req.UserName)
				Expect(err).NotTo(HaveOccurred())
				Ω(userSnap(wsv1User)).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: "normal-user", DisplayName: "namechanged"}),
			Entry(nil, "normal-user", &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: "normal-user", DisplayName: "namechanged"}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: "XXXXXX", DisplayName: "namechanged"}),
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: "normal-user", DisplayName: ""}),
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: "", DisplayName: ""}),
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: "normal-user", DisplayName: "お名前"}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry(nil, "normal-user", &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: "admin-user", DisplayName: "namechanged"}),
		)

		DescribeTable("❌ fail with an unexpected error to update:",
			func(loginUser string, req *dashv1alpha1.UpdateUserDisplayNameRequest) {
				clientMock.SetUpdateError((*Server).UpdateUserDisplayName, errors.New("mock update user error"))
				run_test(loginUser, req)
			},
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: "normal-user", DisplayName: "namechanged"}),
		)
	})

	//==================================================================================
	Describe("[UpdateUserRole]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.UpdateUserRoleRequest) {
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.UpdateUserRole(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				wsv1User, err := k8sClient.GetUser(context.Background(), req.UserName)
				Expect(err).NotTo(HaveOccurred())
				Ω(userSnap(wsv1User)).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
				wsv1User, err := k8sClient.GetUser(context.Background(), req.UserName)
				if err != nil {
					Ω(err.Error()).To(MatchSnapShot())
				}
				if wsv1User != nil {
					Ω(userSnap(wsv1User)).To(MatchSnapShot())
				}
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("attach cosmo-admin to normal-user", "admin-user", &dashv1alpha1.UpdateUserRoleRequest{UserName: "normal-user", Roles: []string{"cosmo-admin"}}),
			Entry("attach custom-role to normal-user", "admin-user", &dashv1alpha1.UpdateUserRoleRequest{UserName: "normal-user", Roles: []string{"xxxxx"}}),
			Entry("detach role from admin-user", "admin-user", &dashv1alpha1.UpdateUserRoleRequest{UserName: "admin-user", Roles: []string{""}}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("user not found", "admin-user", &dashv1alpha1.UpdateUserRoleRequest{UserName: "XXXXXX", Roles: []string{"cosmo-admin"}}),
			Entry("no change", "admin-user", &dashv1alpha1.UpdateUserRoleRequest{UserName: "admin-user", Roles: []string{"cosmo-admin"}}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry("permission denied", "normal-user", &dashv1alpha1.UpdateUserRoleRequest{UserName: "normal-user", Roles: []string{"cosmo-admin"}}),
		)
	})

	//==================================================================================
	Describe("[UpdateUserPassword]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.UpdateUserPasswordRequest) {
			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.UpdateUserPassword(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				verified, _, _ := k8sClient.VerifyPassword(context.Background(), req.UserName, []byte("newPassword"))
				Expect(verified).Should(BeTrue())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success with invalid request:",
			run_test,
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserPasswordRequest{UserName: "admin-user", CurrentPassword: "password", NewPassword: "newPassword"}),
			Entry(nil, "normal-user", &dashv1alpha1.UpdateUserPasswordRequest{UserName: "normal-user", CurrentPassword: "password", NewPassword: "newPassword"}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry(nil, "normal-user", &dashv1alpha1.UpdateUserPasswordRequest{UserName: "admin-user", CurrentPassword: "password", NewPassword: "newPassword"}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserPasswordRequest{UserName: "XXXXXX", CurrentPassword: "password", NewPassword: "newPassword"}),
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserPasswordRequest{UserName: "admin-user", CurrentPassword: "", NewPassword: "newPassword"}),
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserPasswordRequest{UserName: "admin-user", CurrentPassword: "xxxxxx", NewPassword: "newPassword"}),
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserPasswordRequest{UserName: "admin-user", CurrentPassword: "password", NewPassword: ""}),
		)

		DescribeTable("❌ fail to verify password:",
			func(loginUser string, req *dashv1alpha1.UpdateUserPasswordRequest) {
				clientMock.GetMock = func(ctx context.Context, key ctrl_client.ObjectKey, obj ctrl_client.Object, opts ...ctrl_client.GetOption) (mocked bool, err error) {
					if key.Name == cosmov1alpha1.UserPasswordSecretName {
						return true, apierrs.NewNotFound(schema.GroupResource{}, "secret")
					}
					return false, nil
				}
				//clientMock.SetGetError((*Server).PutUserPassword, apierrs.NewNotFound(schema.GroupResource{}, "secret"))
				run_test(loginUser, req)
			},
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserPasswordRequest{UserName: "admin-user", CurrentPassword: "password", NewPassword: "newPassword"}),
		)

		DescribeTable("❌ fail with an unexpected error :",
			func(loginUser string, req *dashv1alpha1.UpdateUserPasswordRequest) {
				clientMock.SetUpdateError((*Server).UpdateUserPassword, errors.New("mock update error"))
				run_test(loginUser, req)
			},
			Entry(nil, "admin-user", &dashv1alpha1.UpdateUserPasswordRequest{UserName: "admin-user", CurrentPassword: "password", NewPassword: "newPassword"}),
		)
	})
})
