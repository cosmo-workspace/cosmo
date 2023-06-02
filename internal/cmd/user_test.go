package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
)

var _ = Describe("cosmoctl [user]", func() {

	var (
		clientMock kubeutil.ClientMock
		rootCmd    *cobra.Command
		options    *cmdutil.CliOptions
		outBuf     *bytes.Buffer
	)
	consoleOut := func() string {
		out, _ := io.ReadAll(outBuf)
		return string(out)
	}

	userSnap := func(us *cosmov1alpha1.User) struct{ Name, Namespace, Spec, Status interface{} } {
		return struct{ Name, Namespace, Spec, Status interface{} }{
			Name:      us.Name,
			Namespace: us.Namespace,
			Spec:      us.Spec,
			Status:    us.Status,
		}
	}

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		utilruntime.Must(clientgoscheme.AddToScheme(scheme))
		utilruntime.Must(cosmov1alpha1.AddToScheme(scheme))
		// +kubebuilder:scaffold:scheme

		baseclient, err := kosmo.NewClientByRestConfig(cfg, scheme)
		Expect(err).NotTo(HaveOccurred())
		clientMock = kubeutil.NewClientMock(baseclient)
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
		testUtil.DeleteCosmoUserAll()
		testUtil.DeleteTemplateAll()
		testUtil.DeleteClusterTemplateAll()
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
			Entry(desc, "user", "reset-password", "user1", "--password", "XXXXXXXX", "--kubeconfig", "XXXX"),
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
				Expect(userSnap(wsv1User)).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				testUtil.CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")
				testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeUserAddon, "user-template1")
				testUtil.CreateClusterTemplate(cosmov1alpha1.TemplateLabelEnumTypeUserAddon, "user-clustertemplate1")
				run_test(args...)
			},
			Entry(desc, "user", "create", "user-create", "--name", "create 1", "--role", "cosmo-admin", "--auth-type", "password-secret", "user-template1,HOGE:HOGEHOGE"),
			Entry(desc, "user", "create", "user-create", "--name", "create 1", "--role", "cosmo-admin", "--auth-type", "ldap", "user-template1,HOGE:HOGEHOGE"),
			Entry(desc, "user", "create", "user-create", "--name", "create 1", "--role", "cosmo-admin", "--addon", "user-template1,HOGE:HOGEHOGE"),
			Entry(desc, "user", "create", "user-create", "--name", "create 1", "--admin", "--addon", "user-template1,HOGE:HOGEHOGE"),
			Entry(desc, "user", "create", "user-create"),
			Entry(desc, "user", "create", "user-create", "--addon", "user-template1"),
			Entry(desc, "user", "create", "user-create", "--addon", "user-template1,HOGE: HOGE HOGE ,FUGA:FUGAF:UGA"),
			Entry(desc, "user", "create", "user-create", "--addon", "user-template1", "--cluster-addon", "user-clustertemplate1"),
			Entry(desc, "user", "create", "user-create", "--admin", "--role", "cosmo-admin"),
			Entry(desc, "user", "create", "user-create", "--role", "xxx"),
		)

		DescribeTable("✅ success to create password immediately:",
			func(args ...string) {
				testUtil.CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create")
				run_test(args...)
			},
			Entry(desc, "user", "create", "user-create"),
		)

		DescribeTable("✅ success to create password later:",
			func(args ...string) {
				timer := time.AfterFunc(100*time.Millisecond, func() {
					testUtil.CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create-later")
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
			Entry(desc, "user", "create", "user-create", "--addon", "XXXXXXXXX,HOGE:yyy"),
			Entry(desc, "user", "create", "user-create", "--addon", "user-template1 ,HOGE:yyy"),
			Entry(desc, "user", "create", "user-create", "--addon", "user-template1,HOGE :yyy"),
			Entry(desc, "user", "create", "user-create", "--cluster-addon", "user-clustertemplate1,HOGE :"),
			Entry(desc, "user", "create", "user-create", "--auth-type", "xxxx"),
		)

		DescribeTable("❌ fail to create password timeout",
			func(args ...string) {
				timer := time.AfterFunc(30*time.Second, func() {
					testUtil.CreateUserNameSpaceandDefaultPasswordIfAbsent("user-create-timeout")
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
				testUtil.CreateLoginUser("user1", "name1", nil, cosmov1alpha1.UserAuthTypePasswordSecert, "password")
				testUtil.CreateLoginUser("user2", "name2", []cosmov1alpha1.UserRole{cosmov1alpha1.PrivilegedRole}, cosmov1alpha1.UserAuthTypePasswordSecert, "password")
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
			testUtil.CreateCosmoUser("user-delete1", "delete", nil, cosmov1alpha1.UserAuthTypePasswordSecert)
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

		var (
			noRoleUser string = "upd-norole"
			privUser   string = "upd-priv"
		)

		run_test := func(args ...string) {
			testUtil.CreateCosmoUser(noRoleUser, "ロールなし", nil, cosmov1alpha1.UserAuthTypePasswordSecert)
			testUtil.CreateCosmoUser(privUser, "特権",
				[]cosmov1alpha1.UserRole{{Name: "cosmo-admin"}}, cosmov1alpha1.UserAuthTypePasswordSecert)

			By("---------------test start----------------")
			var befUser *cosmov1alpha1.User
			if len(args) > 2 {
				befUser, _ = k8sClient.GetUser(ctx, args[2])
			}

			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			Ω(consoleOut()).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			if err == nil {
				wsv1User, err := k8sClient.GetUser(context.Background(), args[2])
				Expect(err).NotTo(HaveOccurred())
				Expect(userSnap(befUser)).To(MatchSnapShot())
				Expect(userSnap(wsv1User)).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			run_test,
			Entry(desc, "user", "update", noRoleUser, "--name", "namechanged"),
			Entry(desc, "user", "update", noRoleUser, "--role", "cosmo-admin"),
			Entry(desc, "user", "update", noRoleUser, "--role", "team-developer,otherteam-developer"),
			Entry(desc, "user", "update", noRoleUser, "--name", "name1", "--role", "team-dev"),
			Entry(desc, "user", "update", privUser, "--role", ""),
			Entry(desc, "user", "update", noRoleUser, "--name", ""),
		)

		DescribeTable("❌ fail with invalid args:",
			run_test,
			Entry(desc, "user", "update"),
			Entry(desc, "user", "update", privUser),
			Entry(desc, "user", "update", "notfound", "--name", "namechanged", "--role", "cosmo-admin"),
			Entry(desc, "user", "update", privUser, "--role", "cosmo-admin"),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(args ...string) {
				clientMock.SetUpdateError("\\.RunE$", errors.New("mock update error"))
				run_test(args...)
			},
			Entry(desc, "user", "update", noRoleUser, "--name", "namechanged"),
		)
	})

	//==================================================================================
	Describe("[reset-password]", func() {

		run_test := func(args ...string) {
			testUtil.CreateLoginUser("user1", "name1", nil, cosmov1alpha1.UserAuthTypePasswordSecert, "password")
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
			Entry(desc, "user", "reset-password", "user1", "--password", "XXXXXXXX"),
		)

		DescribeTable("❌ fail with invalid args:",
			run_test,
			Entry(desc, "user", "reset-password", "XXXXXX"),
			Entry(desc, "user", "reset-password"),
			Entry(desc, "user", "reset-password", "user1", "--password", ""),
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
