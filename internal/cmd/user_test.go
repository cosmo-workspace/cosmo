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
	})

	AfterEach(func() {
		clientMock.Clear()
		test_DeleteCosmoUserAll()
		test_DeleteTemplateAll()
	})

	//==================================================================================
	desc := func(args ...string) string { return strings.Join(args, " ") }
	errSnap := func(err error) string {
		if err == nil {
			return "success"
		} else {
			return err.Error()
		}
	}
	//==================================================================================
	Describe("[all]", func() {

		DescribeTable("❌ fail with invalid arg: kubeconfig",
			func(args ...string) {
				By("---------------test start----------------")
				options.Client = nil
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).Should(HaveOccurred())
				Ω(consoleOut()).To(MatchSnapShot())
				By("---------------test end---------------")
			},
			Entry(desc, "user", "create", "user1", "--kubeconfig", "XXXX"),
			Entry(desc, "user", "get", "--kubeconfig", "XXXX"),
			Entry(desc, "user", "delete", "user1", "--kubeconfig", "XXXX"),
			Entry(desc, "user", "update", "user1", "--kubeconfig", "XXXX"),
			Entry(desc, "user", "reset-password", "user1", "--kubeconfig", "XXXX"),
		)
	})

	//==================================================================================
	Describe("[create]", func() {

		run_test := func(args ...string) {
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			o := consoleOut()
			o = regexp.MustCompile("Default password: .*").ReplaceAllString(o, "Default password: xxxxxxxx")
			Ω(o).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			if err == nil {
				wsv1User, err := k8sClient.GetUser(context.Background(), args[2])
				Expect(err).NotTo(HaveOccurred()) // created
				userSnap := struct{ Name, Namespace, Spec, Status interface{} }{
					Name:      wsv1User.Name,
					Namespace: wsv1User.Namespace,
					Spec:      wsv1User.Spec,
					Status:    wsv1User.Status,
				}
				Expect(userSnap).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")
				test_CreateTemplate(wsv1alpha1.TemplateTypeUserAddon, "user-temple1")
				run_test(args...)
			},
			Entry(desc, "user", "create", "user-create", "--name", "create 1", "--role", "cosmo-admin", "--addons", "user-temple1", "--addon-vars", "Addon=user-temple1,HOGE=HOGEHOGE"),
			Entry(desc, "user", "create", "user-create", "--name", "create 1", "--admin", "--addons", "user-temple1", "--addon-vars", "Addon=user-temple1,HOGE=HOGEHOGE"),
			// Entry(desc, "user", "create", "user-create", "--name", "create 1", "--role", "cosmo-admin", "--addons", "user-temple1", "--addon-vars", "Addon=user-temple1,HOGE=HOGEHOGE,FUGA=FUGAFUGA"),
			Entry(desc, "user", "create", "user-create"),
			Entry(desc, "user", "create", "user-create", "--addons", "user-temple1", "--addon-vars", "Addon=user-temple1,"),
		)

		DescribeTable("✅ success to create password immediately:",
			func(args ...string) {
				test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")
				run_test(args...)
			},
			Entry(desc, "user", "create", "user-create"),
		)

		DescribeTable("✅ success to create password later:",
			func(args ...string) {
				timer := time.AfterFunc(100*time.Millisecond, func() {
					test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create-later")
				})
				defer timer.Stop()
				run_test(args...)
			},
			Entry(desc, "user", "create", "user-create-later"),
		)

		DescribeTable("❌ fail with invalid args:",
			run_test,
			Entry(desc, "user", "create"),
			Entry(desc, "user", "create", "--admin"),
			Entry(desc, "user", "create", "TESTuser"),
			Entry(desc, "user", "create", "user-create", "--admin", "--role", "cosmo-admin"),
			Entry(desc, "user", "create", "user-create", "--role", "xxx"),
			Entry(desc, "user", "create", "user-create", "user-test", "--addons", "user-temple1", "--addon-vars", "Addon=user-temple1,HOGE=xxx=yyy"),
		)

		DescribeTable("❌ fail to create password timeout",
			func(args ...string) {
				timer := time.AfterFunc(30*time.Second, func() {
					test_CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create-timeout")
				})
				defer timer.Stop()
				run_test(args...)
			},
			Entry(desc, "user", "create", "user-create-timeout"),
		)
	})

	//==================================================================================
	Describe("[get]", func() {

		run_test := func(args ...string) {
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			Ω(consoleOut()).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				test_CreateLoginUser("user1", "name1", "", "password")
				test_CreateLoginUser("user2", "name2", wsv1alpha1.UserAdminRole, "password")
				run_test(args...)
			},
			Entry(desc, "user", "get"),
		)

		DescribeTable("✅ success with empty user:",
			run_test,
			Entry(desc, "user", "get"),
		)

		DescribeTable("❌ fail with an unexpected error at list:",
			func(args ...string) {
				clientMock.SetListError("\\.ListUsers$", errors.New("mock user list error"))
				run_test(args...)
			},
			Entry(desc, "user", "get"),
		)
	})

	//==================================================================================
	Describe("[delete]", func() {

		run_test := func(args ...string) {
			test_CreateCosmoUser("user-delete1", "delete", "")
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			Ω(consoleOut()).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			user, _ := k8sClient.GetUser(context.Background(), "user-delete1")
			if err == nil {
				Expect(user).Should(BeNil()) // deleted
			} else {
				Expect(user).ShouldNot(BeNil()) // undeleted
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(desc, "user", "delete", "user-delete1"),
		)

		DescribeTable("❌ fail with invalid args:",
			run_test,
			Entry(desc, "user", "delete"),
			Entry(desc, "user", "delete", "XXXXX"),
		)

		DescribeTable("❌ fail with an unexpected error at delete:",
			func(args ...string) {
				clientMock.SetDeleteError("\\.RunE$", errors.New("mock delete user error"))
				run_test(args...)
			},

			Entry(desc, "user", "delete", "user-delete1"),
		)
	})

	//==================================================================================
	Describe("[update]", func() {

		run_test := func(args ...string) {
			test_CreateLoginUser("user1", "name1", "", "password")
			test_CreateLoginUser("user2", "name2", wsv1alpha1.UserAdminRole, "password")
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			Ω(consoleOut()).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			if err == nil {
				wsv1User, err := k8sClient.GetUser(context.Background(), args[2])
				Expect(err).NotTo(HaveOccurred())
				userSnap := struct{ Name, Namespace, Spec, Status interface{} }{
					Name:      wsv1User.Name,
					Namespace: wsv1User.Namespace,
					Spec:      wsv1User.Spec,
					Status:    wsv1User.Status,
				}
				Expect(userSnap).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(desc, "user", "update", "user1", "--name", "namechanged"),
			Entry(desc, "user", "update", "user1", "--role", "cosmo-admin"),
			Entry(desc, "user", "update", "user2", "--role", ""),
		)

		DescribeTable("❌ fail with invalid args:",
			run_test,
			Entry(desc, "user", "update"),
			Entry(desc, "user", "update", "user1"),
			Entry(desc, "user", "update", "XXXXXX", "--name", "namechanged", "--role", "cosmo-admin"),
			Entry(desc, "user", "update", "user1", "--name", ""),
			Entry(desc, "user", "update", "user1", "--name", "name1", "--role", ""),
			Entry(desc, "user", "update", "user1", "--role", "xxxxx"),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(args ...string) {
				clientMock.SetUpdateError("\\.RunE$", errors.New("mock update error"))
				run_test(args...)
			},
			Entry(desc, "user", "update", "user1", "--name", "namechanged"),
		)
	})

	//==================================================================================
	Describe("[reset-password]", func() {

		run_test := func(args ...string) {
			test_CreateLoginUser("user1", "name1", "", "password")
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			o := consoleOut()
			o = regexp.MustCompile("New password: .*").ReplaceAllString(o, "New password: xxxxxxxx")
			Ω(o).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(desc, "user", "reset-password", "user1"),
		)

		DescribeTable("❌ fail with invalid args:",
			run_test,
			Entry(desc, "user", "reset-password", "XXXXXX"),
			Entry(desc, "user", "reset-password"),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(args ...string) {
				clientMock.SetGetError("\\.GetDefaultPassword$", errors.New("mock get error"))
				run_test(args...)
			},
			Entry(desc, "user", "reset-password", "user1"),
		)
	})
})
