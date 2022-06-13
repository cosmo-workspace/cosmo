package cmd

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
)

var _ = Describe("cosmoctl [user]", func() {

	var (
		clientMock kosmo.ClientMock
		rootCmd    *cobra.Command
		options    *cmdutil.CliOptions
		outBuf     *bytes.Buffer
	)
	consoleOut := func() string {
		out, _ := ioutil.ReadAll(outBuf)
		return string(out)
	}

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		_ = clientgoscheme.AddToScheme(scheme)
		_ = cosmov1alpha1.AddToScheme(scheme)
		_ = wsv1alpha1.AddToScheme(scheme)

		baseclient, err := kosmo.NewClientByRestConfig(cfg, scheme)
		Expect(err).NotTo(HaveOccurred())
		clientMock = kosmo.NewClientMock(baseclient)
		klient := kosmo.NewClient(&clientMock)

		options = cmdutil.NewCliOptions()
		options.Client = &klient
		outBuf = bytes.NewBufferString("")
		options.Out = outBuf
		options.ErrOut = outBuf
		options.Scheme = scheme
		rootCmd = NewRootCmd(options)
		By("---------------BeforeEach end----------------")
	})

	AfterEach(func() {
		By("---------------AfterEach start---------------")
		clientMock.Clear()
		test_DeleteCosmoUserAll()
		test_DeleteTemplateAll()
	})

	//==================================================================================
	Describe("[all]", func() {

		DescribeTable("fail with invalid arg: kubeconfig",
			func(args ...string) {
				options.Client = nil
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).Should(HaveOccurred())
				Ω(consoleOut()).To(MatchSnapShot())
			},
			func(args ...string) string { return strings.Join(args, " ") },
			Entry(nil, "user", "create", "user1", "--kubeconfig", "XXXX"),
			Entry(nil, "user", "get", "--kubeconfig", "XXXX"),
			Entry(nil, "user", "delete", "user1", "--kubeconfig", "XXXX"),
			Entry(nil, "user", "update", "user1", "--kubeconfig", "XXXX"),
			Entry(nil, "user", "reset-password", "user1", "--kubeconfig", "XXXX"),
		)

	})

	//==================================================================================
	Describe("[create]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")
					test_CreateTemplate(wsv1alpha1.TemplateTypeUserAddon, "user-temple1")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					o := consoleOut()
					o = regexp.MustCompile("Default password: .*").ReplaceAllString(o, "Default password: xxxxxxxx")
					Ω(o).To(MatchSnapShot())

					wsv1User, err := k8sClient.GetUser(context.Background(), args[2])
					Expect(err).NotTo(HaveOccurred()) // created

					userSnap := struct{ Name, Namespace, Spec, Status interface{} }{
						Name:      wsv1User.Name,
						Namespace: wsv1User.Namespace,
						Spec:      wsv1User.Spec,
						Status:    wsv1User.Status,
					}
					Expect(userSnap).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "create", "user-create", "--name", "create 1", "--role", "cosmo-admin", "--addons", "user-temple1", "--addon-vars", "Addon=user-temple1,HOGE=HOGEHOGE"),
				Entry(nil, "user", "create", "user-create", "--name", "create 1", "--admin", "--addons", "user-temple1", "--addon-vars", "Addon=user-temple1,HOGE=HOGEHOGE"),
				// Entry(nil, "user", "create", "user-create", "--name", "create 1", "--role", "cosmo-admin", "--addons", "user-temple1", "--addon-vars", "Addon=user-temple1,HOGE=HOGEHOGE,FUGA=FUGAFUGA"),
				Entry(nil, "user", "create", "user-create"),
				Entry(nil, "user", "create", "user-create", "--addons", "user-temple1", "--addon-vars", "Addon=user-temple1,"),
			)

			DescribeTable("to create password immediately:",
				func(args ...string) {
					test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					o := consoleOut()
					o = regexp.MustCompile("Default password: .*").ReplaceAllString(o, "Default password: xxxxxxxx")
					Ω(o).To(MatchSnapShot())

					user, _ := k8sClient.GetUser(context.Background(), "user-create")
					Expect(user).ShouldNot(BeNil()) // created
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "create", "user-create"),
			)

			DescribeTable("to create password later:",
				func(args ...string) {
					timer := time.AfterFunc(100*time.Millisecond, func() {
						test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create-later")
					})
					defer timer.Stop()

					rootCmd.SetArgs([]string{"user", "create", "user-create-later"})
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					o := consoleOut()
					o = regexp.MustCompile("Default password: .*").ReplaceAllString(o, "Default password: xxxxxxxx")
					Ω(o).To(MatchSnapShot())

					user, _ := k8sClient.GetUser(context.Background(), "user-create-later")
					Expect(user).ShouldNot(BeNil()) // created
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "create", "user-create-later"),
			)
		})

		Describe("fail", func() {

			DescribeTable("with invalid args:",
				func(args ...string) {
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "create"),
				Entry(nil, "user", "create", "--admin"),
				Entry(nil, "user", "create", "TESTuser"),
				Entry(nil, "user", "create", "user-create", "--admin", "--role", "cosmo-admin"),
				Entry(nil, "user", "create", "user-create", "--role", "xxx"),
				Entry(nil, "user", "create", "user-create", "user-test", "--addons", "user-temple1", "--addon-vars", "Addon=user-temple1,HOGE=xxx=yyy"),
			)

			DescribeTable("to create password timeout",
				func(args ...string) {
					timer := time.AfterFunc(30*time.Second, func() {
						test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create-timeout")
					})
					defer timer.Stop()

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Ω(consoleOut()).To(MatchSnapShot())

					user, _ := k8sClient.GetUser(context.Background(), "user-create-timeout")
					Expect(user).ShouldNot(BeNil()) // created
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "create", "user-create-timeout"),
			)
		})
	})

	//==================================================================================
	Describe("[get]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					test_CreateLoginUser("user1", "name1", "", "password")
					test_CreateLoginUser("user2", "name2", wsv1alpha1.UserAdminRole, "password")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "get"),
			)

			DescribeTable("with empty user:",
				func(args ...string) {
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "get"),
			)
		})

		Describe("fail", func() {

			DescribeTable("with an unexpected error at list:",
				func(args ...string) {
					clientMock.ListMock = func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.ListUsers$") {
							return true, errors.New("user list error")
						}
						return false, nil
					}
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "get"),
			)
		})
	})

	//==================================================================================
	Describe("[delete]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					test_CreateCosmoUser("user-delete1", "delete", "")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(consoleOut()).To(MatchSnapShot())
					user, _ := k8sClient.GetUser(context.Background(), "user-delete1")
					Expect(user).Should(BeNil()) // deleted
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "delete", "user-delete1"),
			)
		})

		Describe("fail", func() {

			DescribeTable("with invalid args:",
				func(args ...string) {
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Ω(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "delete"),
				Entry(nil, "user", "delete", "XXXXX"),
			)

			DescribeTable("with an unexpected error at delete:",
				func(args ...string) {
					clientMock.DeleteMock = func(ctx context.Context, obj client.Object, opts ...client.DeleteOption) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.RunE$") {
							return true, errors.New("delete user error")
						}
						return false, nil
					}
					test_CreateCosmoUser("user-delete1", "delete", "")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Ω(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "delete", "user-delete1"),
			)
		})
	})

	//==================================================================================
	Describe("[update]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					test_CreateLoginUser("user1", "name1", "", "password")
					test_CreateLoginUser("user2", "name2", wsv1alpha1.UserAdminRole, "password")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					Ω(consoleOut()).To(MatchSnapShot())

					wsv1User, err := k8sClient.GetUser(context.Background(), args[2])
					Expect(err).NotTo(HaveOccurred())

					userSnap := struct{ Name, Namespace, Spec, Status interface{} }{
						Name:      wsv1User.Name,
						Namespace: wsv1User.Namespace,
						Spec:      wsv1User.Spec,
						Status:    wsv1User.Status,
					}
					Expect(userSnap).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "update", "user1", "--name", "namechanged"),
				Entry(nil, "user", "update", "user1", "--role", "cosmo-admin"),
				Entry(nil, "user", "update", "user2", "--role", ""),
			)
		})

		Describe("fail", func() {

			DescribeTable("with invalid args:",
				func(args ...string) {
					test_CreateLoginUser("user1", "name1", "", "password")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Ω(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "update"),
				Entry(nil, "user", "update", "user1"),
				Entry(nil, "user", "update", "XXXXXX", "--name", "namechanged", "--role", "cosmo-admin"),
				Entry(nil, "user", "update", "user1", "--name", ""),
				Entry(nil, "user", "update", "user1", "--name", "name1", "--role", ""),
				Entry(nil, "user", "update", "user1", "--role", "xxxxx"),
			)

			DescribeTable("with an unexpected error at update:",
				func(args ...string) {
					clientMock.UpdateMock = func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.RunE$") {
							return true, errors.New("mock update error")
						}
						return false, nil
					}
					test_CreateLoginUser("user1", "name1", "", "password")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Ω(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "update", "user1", "--name", "namechanged"),
			)
		})
	})

	//==================================================================================
	Describe("[reset-password]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					test_CreateLoginUser("user1", "name1", "", "password")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					o := consoleOut()
					o = regexp.MustCompile("New password: .*").ReplaceAllString(o, "New password: xxxxxxxx")
					Ω(o).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "reset-password", "user1"),
			)
		})

		Describe("fail", func() {

			DescribeTable("with invalid args:",
				func(args ...string) {
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Ω(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "reset-password", "XXXXXX"),
				Entry(nil, "user", "reset-password"),
			)

			DescribeTable("with an unexpected error at update:",
				func(args ...string) {
					clientMock.GetMock = func(ctx context.Context, key client.ObjectKey, obj client.Object) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.GetDefaultPassword$") {
							return true, errors.New("mock get error")
						}
						return false, nil
					}
					test_CreateLoginUser("user1", "name1", "", "password")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Ω(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "user", "reset-password", "user1"),
			)
		})
	})
})
