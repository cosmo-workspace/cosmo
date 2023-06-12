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

	const (
		normalUser     string = "normal-user"
		adminUser      string = "admin-user"
		privilegedUser string = "priv-user"
	)

	var (
		userSession       string
		adminSession      string
		privilegedSession string
		client            dashboardv1alpha1connect.UserServiceClient
	)

	BeforeEach(func() {
		userSession = test_CreateLoginUserSession(normalUser, "お名前", []cosmov1alpha1.UserRole{{Name: "team-developer"}}, "password")
		adminSession = test_CreateLoginUserSession(adminUser, "アドミン", []cosmov1alpha1.UserRole{{Name: "team-admin"}}, "password")
		privilegedSession = test_CreateLoginUserSession(privilegedUser, "特権", []cosmov1alpha1.UserRole{cosmov1alpha1.PrivilegedRole}, "password")
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
		switch loginUser {
		case adminUser:
			return adminSession
		case privilegedUser:
			return privilegedSession
		default:
			return userSession
		}
	}

	//==================================================================================
	Describe("[CreateUser]", func() {

		withUser := func(name string) func(req *dashv1alpha1.CreateUserRequest) *time.Timer {
			return func(req *dashv1alpha1.CreateUserRequest) *time.Timer {
				testUtil.CreateUserNameSpaceandDefaultPasswordIfAbsent(name)
				_, err := client.CreateUser(ctx, NewRequestWithSession(req, privilegedSession))
				Expect(err).NotTo(HaveOccurred())
				return nil
			}
		}

		withNamespace := func(delay time.Duration) func(req *dashv1alpha1.CreateUserRequest) *time.Timer {
			return func(req *dashv1alpha1.CreateUserRequest) *time.Timer {
				createNamespace := func() {
					testUtil.CreateUserNameSpaceandDefaultPasswordIfAbsent(req.UserName)
				}
				if delay > 0 {
					return time.AfterFunc(delay, createNamespace)
				} else {
					createNamespace()
					return nil
				}
			}
		}
		withUserAddon := func(name string) func(req *dashv1alpha1.CreateUserRequest) *time.Timer {
			return func(req *dashv1alpha1.CreateUserRequest) *time.Timer {
				testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeUserAddon, name)
				return nil
			}
		}

		run_test := func(loginUser string, req *dashv1alpha1.CreateUserRequest, beforeFuncs ...func(req *dashv1alpha1.CreateUserRequest) *time.Timer) {
			for _, f := range beforeFuncs {
				if f != nil {
					if t := f(req); t != nil {
						defer t.Stop()
					}
				}
			}

			By("---------------test start----------------")
			ctx := context.Background()
			res, err := client.CreateUser(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				if res.Msg.User.AuthType == cosmov1alpha1.UserAuthTypePasswordSecert.String() {
					Ω(res.Msg.User.DefaultPassword).ShouldNot(BeEmpty())
					res.Msg.User.DefaultPassword = "xxxxxxxx"
				}
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

		DescribeTable("✅ success create user by privileged role:",
			run_test,
			Entry("create normal user", privilegedUser, &dashv1alpha1.CreateUserRequest{
				UserName:    "create-user-by-priv",
				DisplayName: "create 1",
				Roles:       []string{"team-a", "team-b"},
				AuthType:    "password-secret",
				Addons: []*dashv1alpha1.UserAddon{{
					Template: "user-tmpl1",
					Vars:     map[string]string{"HOGE": "FUGA"},
				}},
			}, withNamespace(0), withUserAddon("user-tmpl1")),
			Entry("create normal user with ldap auth", privilegedUser, &dashv1alpha1.CreateUserRequest{
				UserName: "create-user-ldap", AuthType: "ldap"}, withNamespace(0)),
			Entry("create user with only name", privilegedUser, &dashv1alpha1.CreateUserRequest{UserName: "create-user-only-name"}, withNamespace(0)),
			Entry("create privileged user", privilegedUser, &dashv1alpha1.CreateUserRequest{UserName: "create-user-priv-by-priv", Roles: []string{"cosmo-admin"}}, withNamespace(0)),
			Entry("create user with namespace creation short delay before timeout", privilegedUser, &dashv1alpha1.CreateUserRequest{UserName: "create-user-with-delay"}, withNamespace(100*time.Millisecond)),
		)

		DescribeTable("✅ success create user by admin:",
			run_test,
			Entry("create group-developer user", adminUser, &dashv1alpha1.CreateUserRequest{
				UserName: "create-user-team-by-admin",
				Roles:    []string{"team-developer"},
			}, withNamespace(0)),
			Entry("create group-admin user", privilegedUser, &dashv1alpha1.CreateUserRequest{
				UserName: "create-user-admin-by-admin",
				Roles:    []string{"team-admin"},
			}, withNamespace(0)),
			Entry("create group-admin user", privilegedUser, &dashv1alpha1.CreateUserRequest{
				UserName: "create-user-admin-by-admin",
				Roles:    []string{"team-developer", "team-etc"},
			}, withNamespace(0)),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("invalid auth type", privilegedUser, &dashv1alpha1.CreateUserRequest{
				UserName: "create-user-invalid-auth",
				AuthType: "INVALID"}),
			Entry("no name", privilegedUser, &dashv1alpha1.CreateUserRequest{
				UserName: ""}),
			Entry("user already exist", privilegedUser, &dashv1alpha1.CreateUserRequest{
				UserName: "create-user-existing",
			}, withUser("create-user-existing")),
			Entry("including invalid charactor in username", privilegedUser, &dashv1alpha1.CreateUserRequest{
				UserName: "create-user-INVALID"}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry("normal user cannot create user", normalUser, &dashv1alpha1.CreateUserRequest{
				UserName: "create-user-by-normal"}),
			Entry("admin user cannot create user including other roles", adminUser, &dashv1alpha1.CreateUserRequest{
				UserName: "create-user-other-role-by-admin",
				Roles:    []string{"team-developer", "cosmo-admin"},
			}),
		)

		DescribeTable("❌ fail to create password timeout",
			run_test,
			Entry(nil, privilegedUser, &dashv1alpha1.CreateUserRequest{UserName: "create-user-no-namespace"}),
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
			Entry(nil, privilegedUser),
			Entry(nil, adminUser),
		)

		DescribeTable("✅ success with empty user:",
			func(loginUser string) {
				clientMock.SetListError((*Server).GetUsers, nil)
				run_test(loginUser)
			},
			Entry(nil, privilegedUser),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry(nil, normalUser),
		)

		DescribeTable("❌ fail with an unexpected error at list:",
			func(loginUser string) {
				clientMock.SetListError((*Server).GetUsers, errors.New("mock user list error"))
				run_test(loginUser)
			},
			Entry(nil, privilegedUser),
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
			Entry("get myself", adminUser, adminUser),
			Entry("get myself", normalUser, normalUser),
			Entry("privileged user can get other", privilegedUser, normalUser),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("user not found", privilegedUser, "notfound"),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry("cannot get the other", normalUser, privilegedUser),
			Entry("cannot get the other", adminUser, normalUser),
		)

		DescribeTable("❌ fail with an unexpected error to get:",
			func(loginUser string, username string) {
				clientMock.SetGetError((*Server).GetUser, errors.New("mock test error"))
				run_test(loginUser, username)
			},
			Entry("unexpected err", normalUser, normalUser),
		)
	})

	//==================================================================================
	Describe("[DeleteUser]", func() {

		var (
			noRoleUser                     string = "del-norole"
			teamDevRoleUser                string = "del-team-dev"
			otherteamDevRoleUser           string = "del-otherteam-dev"
			teamDevAndOtherteamDevRoleUser string = "del-team-otherteam-dev"
		)

		run_test := func(loginUser string, req *dashv1alpha1.DeleteUserRequest) {
			testUtil.CreateCosmoUser(noRoleUser, "", nil, cosmov1alpha1.UserAuthTypePasswordSecert)
			testUtil.CreateCosmoUser(teamDevRoleUser, "",
				[]cosmov1alpha1.UserRole{{Name: "team-developer"}}, cosmov1alpha1.UserAuthTypePasswordSecert)
			testUtil.CreateCosmoUser(otherteamDevRoleUser, "",
				[]cosmov1alpha1.UserRole{{Name: "otherteam-developer"}}, cosmov1alpha1.UserAuthTypePasswordSecert)
			testUtil.CreateCosmoUser(teamDevAndOtherteamDevRoleUser, "",
				[]cosmov1alpha1.UserRole{{Name: "team-developer"}, {Name: "otherteam-developer"}}, cosmov1alpha1.UserAuthTypePasswordSecert)
			By("---------------test start----------------")
			ctx := context.Background()
			_, beferr := k8sClient.GetUser(context.Background(), req.UserName)

			res, err := client.DeleteUser(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				_, err = k8sClient.GetUser(context.Background(), req.UserName)
				Expect(err).To(HaveOccurred())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())
				_, err = k8sClient.GetUser(context.Background(), req.UserName)
				if beferr != nil {
					Expect(err).Should(Equal(beferr))
				}
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, privilegedUser, &dashv1alpha1.DeleteUserRequest{UserName: noRoleUser}),
			Entry(nil, privilegedUser, &dashv1alpha1.DeleteUserRequest{UserName: teamDevRoleUser}),
			Entry(nil, privilegedUser, &dashv1alpha1.DeleteUserRequest{UserName: teamDevAndOtherteamDevRoleUser}),
			Entry(nil, privilegedUser, &dashv1alpha1.DeleteUserRequest{UserName: adminUser}),
			Entry(nil, privilegedUser, &dashv1alpha1.DeleteUserRequest{UserName: normalUser}),
			Entry(nil, adminUser, &dashv1alpha1.DeleteUserRequest{UserName: teamDevRoleUser}),
			Entry(nil, adminUser, &dashv1alpha1.DeleteUserRequest{UserName: noRoleUser}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("user not found", privilegedUser, &dashv1alpha1.DeleteUserRequest{UserName: "xxxxxx"}),
			Entry("deleting myself", privilegedUser, &dashv1alpha1.DeleteUserRequest{UserName: privilegedUser}),
			Entry("deleting myself", adminUser, &dashv1alpha1.DeleteUserRequest{UserName: adminUser}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry("normal user cannot delete anyone", normalUser, &dashv1alpha1.DeleteUserRequest{UserName: noRoleUser}),
			Entry("admin user cannot delete the user which has an other roles", adminUser, &dashv1alpha1.DeleteUserRequest{UserName: otherteamDevRoleUser}),
			Entry("admin user cannot delete the user which has an other roles", adminUser, &dashv1alpha1.DeleteUserRequest{UserName: teamDevAndOtherteamDevRoleUser}),
		)

		DescribeTable("❌ fail with an unexpected error to get:",
			func(loginUser string, req *dashv1alpha1.DeleteUserRequest) {
				clientMock.SetGetError(`\.preFetchUserMiddleware\.|\.DeleteUser$`, errors.New("mock get user error"))
				run_test(loginUser, req)
			},
			Entry("unexpected err on get", privilegedUser, &dashv1alpha1.DeleteUserRequest{UserName: noRoleUser}),
		)

		DescribeTable("❌ fail with an unexpected error to delete:",
			func(loginUser string, req *dashv1alpha1.DeleteUserRequest) {
				clientMock.SetDeleteError((*Server).DeleteUser, errors.New("mock delete user error"))
				run_test(loginUser, req)
			},
			Entry("unexpected err on delete", privilegedUser, &dashv1alpha1.DeleteUserRequest{UserName: noRoleUser}),
		)
	})

	//==================================================================================
	Describe("[UpdateUserDisplayName]", func() {

		run_test := func(loginUser string, req *dashv1alpha1.UpdateUserDisplayNameRequest) {
			By("---------------test start----------------")
			ctx := context.Background()
			befUser, beferr := k8sClient.GetUser(ctx, req.UserName)

			res, err := client.UpdateUserDisplayName(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				wsv1User, err := k8sClient.GetUser(ctx, req.UserName)
				Expect(err).NotTo(HaveOccurred())
				Ω(userSnap(befUser)).To(MatchSnapShot())
				Ω(userSnap(wsv1User)).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())

				_, afterr := k8sClient.GetUser(ctx, req.UserName)
				if beferr != nil {
					Expect(afterr).Should(Equal(beferr))
				}
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(nil, privilegedUser, &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: normalUser, DisplayName: "namechanged"}),
			Entry(nil, normalUser, &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: normalUser, DisplayName: "namechanged"}),
			Entry("empty display name", privilegedUser, &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: normalUser, DisplayName: ""}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("user not found", privilegedUser, &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: "XXXXXX", DisplayName: "namechanged"}),
			Entry("user not found: empty user name", privilegedUser, &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: "", DisplayName: ""}),
			Entry("no change", privilegedUser, &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: normalUser, DisplayName: "お名前"}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry(nil, normalUser, &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: privilegedUser, DisplayName: "namechanged"}),
			Entry(nil, adminUser, &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: normalUser, DisplayName: "namechanged"}),
		)

		DescribeTable("❌ fail with an unexpected error to update:",
			func(loginUser string, req *dashv1alpha1.UpdateUserDisplayNameRequest) {
				clientMock.SetUpdateError((*Server).UpdateUserDisplayName, errors.New("mock update user error"))
				run_test(loginUser, req)
			},
			Entry(nil, privilegedUser, &dashv1alpha1.UpdateUserDisplayNameRequest{UserName: normalUser, DisplayName: "namechanged"}),
		)
	})

	//==================================================================================
	Describe("[UpdateUserRole]", func() {
		var (
			noRoleUser                     string = "upd-norole"
			teamDevRoleUser                string = "upd-team-dev"
			otherteamDevRoleUser           string = "upd-otherteam-dev"
			teamDevAndOtherteamDevRoleUser string = "upd-team-otherteam-dev"
		)

		run_test := func(loginUser string, req *dashv1alpha1.UpdateUserRoleRequest) {
			testUtil.CreateCosmoUser(noRoleUser, "", nil, cosmov1alpha1.UserAuthTypePasswordSecert)
			testUtil.CreateCosmoUser(teamDevRoleUser, "",
				[]cosmov1alpha1.UserRole{{Name: "team-developer"}}, cosmov1alpha1.UserAuthTypePasswordSecert)
			testUtil.CreateCosmoUser(otherteamDevRoleUser, "",
				[]cosmov1alpha1.UserRole{{Name: "otherteam-developer"}}, cosmov1alpha1.UserAuthTypePasswordSecert)
			testUtil.CreateCosmoUser(teamDevAndOtherteamDevRoleUser, "",
				[]cosmov1alpha1.UserRole{{Name: "team-developer"}, {Name: "otherteam-developer"}}, cosmov1alpha1.UserAuthTypePasswordSecert)

			By("---------------test start----------------")
			ctx := context.Background()
			befUser, beferr := k8sClient.GetUser(ctx, req.UserName)

			res, err := client.UpdateUserRole(ctx, NewRequestWithSession(req, getSession(loginUser)))
			if err == nil {
				Ω(res.Msg).To(MatchSnapShot())
				wsv1User, err := k8sClient.GetUser(ctx, req.UserName)
				Expect(err).NotTo(HaveOccurred())
				Ω(userSnap(befUser)).To(MatchSnapShot())
				Ω(userSnap(wsv1User)).To(MatchSnapShot())
			} else {
				Ω(err.Error()).To(MatchSnapShot())
				Expect(res).Should(BeNil())

				_, afterr := k8sClient.GetUser(ctx, req.UserName)
				if beferr != nil {
					Expect(afterr).Should(Equal(beferr))
				}
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry("priv attach cosmo-admin to normal-user", privilegedUser, &dashv1alpha1.UpdateUserRoleRequest{
				UserName: normalUser, Roles: []string{"cosmo-admin"}}),
			Entry("admin attach custom-role to normal-user", adminUser, &dashv1alpha1.UpdateUserRoleRequest{
				UserName: noRoleUser, Roles: []string{"team-developer"}}),
			Entry("admin attach custom-role to other team user", adminUser, &dashv1alpha1.UpdateUserRoleRequest{
				UserName: otherteamDevRoleUser, Roles: []string{"team-developer", "otherteam-developer"}}),
			Entry("priv detach role from priv", privilegedUser, &dashv1alpha1.UpdateUserRoleRequest{
				UserName: privilegedUser, Roles: []string{""}}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry("user not found", privilegedUser, &dashv1alpha1.UpdateUserRoleRequest{
				UserName: "XXXXXX", Roles: []string{"cosmo-admin"}}),
			Entry("no change", privilegedUser, &dashv1alpha1.UpdateUserRoleRequest{
				UserName: privilegedUser, Roles: []string{"cosmo-admin"}}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry("normal user cannot update roles", normalUser, &dashv1alpha1.UpdateUserRoleRequest{
				UserName: noRoleUser, Roles: []string{"cosmo-admin"}}),
			Entry("admin user cannot attach other team role", adminUser, &dashv1alpha1.UpdateUserRoleRequest{
				UserName: noRoleUser, Roles: []string{"cosmo-admin"}}),
			Entry("admin user cannot attach other team role", adminUser, &dashv1alpha1.UpdateUserRoleRequest{
				UserName: teamDevRoleUser, Roles: []string{"team-developer", "otherteam-developer"}}),
			Entry("admin user cannot detach other team role", adminUser, &dashv1alpha1.UpdateUserRoleRequest{
				UserName: teamDevAndOtherteamDevRoleUser, Roles: []string{"team-developer"}}),
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
			Entry(nil, privilegedUser, &dashv1alpha1.UpdateUserPasswordRequest{UserName: privilegedUser, CurrentPassword: "password", NewPassword: "newPassword"}),
			Entry(nil, normalUser, &dashv1alpha1.UpdateUserPasswordRequest{UserName: normalUser, CurrentPassword: "password", NewPassword: "newPassword"}),
		)

		DescribeTable("❌ fail with authorization by role:",
			run_test,
			Entry(nil, normalUser, &dashv1alpha1.UpdateUserPasswordRequest{UserName: privilegedUser, CurrentPassword: "password", NewPassword: "newPassword"}),
			Entry(nil, adminUser, &dashv1alpha1.UpdateUserPasswordRequest{UserName: normalUser, CurrentPassword: "password", NewPassword: "newPassword"}),
		)

		DescribeTable("❌ fail with invalid request:",
			run_test,
			Entry(nil, privilegedUser, &dashv1alpha1.UpdateUserPasswordRequest{UserName: "XXXXXX", CurrentPassword: "password", NewPassword: "newPassword"}),
			Entry(nil, privilegedUser, &dashv1alpha1.UpdateUserPasswordRequest{UserName: privilegedUser, CurrentPassword: "", NewPassword: "newPassword"}),
			Entry(nil, privilegedUser, &dashv1alpha1.UpdateUserPasswordRequest{UserName: privilegedUser, CurrentPassword: "xxxxxx", NewPassword: "newPassword"}),
			Entry(nil, privilegedUser, &dashv1alpha1.UpdateUserPasswordRequest{UserName: privilegedUser, CurrentPassword: "password", NewPassword: ""}),
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
			Entry(nil, privilegedUser, &dashv1alpha1.UpdateUserPasswordRequest{UserName: privilegedUser, CurrentPassword: "password", NewPassword: "newPassword"}),
		)

		DescribeTable("❌ fail with an unexpected error :",
			func(loginUser string, req *dashv1alpha1.UpdateUserPasswordRequest) {
				clientMock.SetUpdateError((*Server).UpdateUserPassword, errors.New("mock update error"))
				run_test(loginUser, req)
			},
			Entry(nil, privilegedUser, &dashv1alpha1.UpdateUserPasswordRequest{UserName: privilegedUser, CurrentPassword: "password", NewPassword: "newPassword"}),
		)
	})
})
