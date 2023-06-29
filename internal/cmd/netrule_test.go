package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
)

var _ = Describe("cosmoctl [netrule]", func() {

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

		testUtil.CreateLoginUser("user2", "お名前", nil, cosmov1alpha1.UserAuthTypePasswordSecert, "password")
		testUtil.CreateLoginUser("user1", "アドミン", []cosmov1alpha1.UserRole{cosmov1alpha1.PrivilegedRole}, cosmov1alpha1.UserAuthTypePasswordSecert, "password")
		testUtil.CreateTemplate(cosmov1alpha1.TemplateLabelEnumTypeWorkspace, "template1")
		By("---------------BeforeEach end----------------")
	})

	AfterEach(func() {
		By("---------------AfterEach start---------------")
		clientMock.Clear()
		testUtil.DeleteWorkspaceAll()
		testUtil.DeleteCosmoUserAll()
		testUtil.DeleteTemplateAll()
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
	Describe("[create]", func() {

		run_test := func(args ...string) {
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			Expect(consoleOut()).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			if err == nil {
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserName(context.Background(), args[5], "user1")
				Expect(err).NotTo(HaveOccurred())
				Ω(ObjectSnapshot(wsv1Workspace)).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				testUtil.CreateWorkspace("user1", "ws1", "template1", nil)
				testUtil.UpsertNetworkRule("user1", "ws1", "nw1", 1111, "/", false, -1)
				testUtil.UpsertNetworkRule("user1", "ws1", "nw3", 2222, "/", false, -1)
				run_test(args...)
			},
			Entry(desc, "netrule", "create", "--user", "user1", "--workspace", "ws1", "--port", "3000", "--host-prefix", "nw11", "--path", "/abc"),
			Entry(desc, "netrule", "create", "--namespace", "cosmo-user-user1", "--workspace", "ws1", "--port", "4000", "--host-prefix", "nw12", "--path", "/def"),
			Entry(desc, "netrule", "create", "--user", "user1", "--workspace", "ws1", "--port", "4000", "--host-prefix", "nw13", "--path", "/def"),
			Entry(desc, "netrule", "create", "--user", "user1", "--workspace", "ws1", "--port", "4000"),
			Entry(desc, "netrule", "create", "--user", "user1", "--workspace", "ws1", "--port", "4000", "--path", "/def"),
		)

		DescribeTable("❌ fail with invalid args:",
			func(args ...string) {
				testUtil.CreateWorkspace("user1", "ws1", "template1", nil)
				testUtil.UpsertNetworkRule("user1", "ws1", "nw1", 1111, "/", false, -1)
				run_test(args...)
			},
			Entry(desc, "netrule", "create", "--user", "xxx", "--workspace", "ws1", "--port", "4000"),
			Entry(desc, "netrule", "create", "--user", "user1", "--workspace", "xxx", "--port", "4000"),
			Entry(desc, "netrule", "create", "--user", "user1", "--workspace", "ws1", "--port", "0"),
			Entry(desc, "netrule", "create", "--user", "user1", "--workspace", "ws1", "--port", "124000"),
			Entry(desc, "netrule", "create", "--user", "user1", "--workspace", "ws1", "--port", "4000", "--host-prefix", "main"),
			Entry(desc, "netrule", "create", "--user", "user1", "--workspace", "ws1", "--port", "4000"),
			Entry(desc, "netrule", "create"),
			Entry(desc, "netrule", "create", "--namespace", "xxxxx", "--workspace", "ws1", "--port", "4000"),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(args ...string) {
				testUtil.CreateWorkspace("user1", "ws1", "template1", nil)
				clientMock.UpdateMock = func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) (mocked bool, err error) {
					if clientMock.IsCallingFrom("\\.RunE$") {
						return true, errors.New("mock update error")
					}
					return false, nil
				}
				run_test(args...)
			},
			Entry(desc, "netrule", "create", "--workspace", "ws1", "--user", "user1", "--host-prefix", "nw99", "--port", "4000", "--path", "/def"),
		)
	})

	//==================================================================================
	Describe("[delete]", func() {

		run_test := func(args ...string) {
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			Expect(consoleOut()).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			if err == nil {
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserName(context.Background(), args[5], args[3])
				Expect(err).NotTo(HaveOccurred())
				Ω(ObjectSnapshot(wsv1Workspace)).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				testUtil.CreateWorkspace("user1", "ws1", "template1", nil)
				testUtil.UpsertNetworkRule("user1", "ws1", "nw1", 1111, "/", false, -1)
				testUtil.UpsertNetworkRule("user1", "ws1", "nw2", 2222, "/", false, -1)
				run_test(args...)
			},
			Entry(desc, "netrule", "delete", "--user", "user1", "--workspace", "ws1", "--index", "0"),
			Entry(desc, "netrule", "delete", "--user", "user1", "--workspace", "ws1", "--index", "1"),
		)

		DescribeTable("❌ fail with invalid args:",
			func(args ...string) {
				testUtil.CreateWorkspace("user1", "ws1", "template1", nil)
				testUtil.UpsertNetworkRule("user1", "ws1", "nw1", 1111, "/", false, -1)
				run_test(args...)
			},
			Entry(desc, "netrule", "delete", "--user", "xxx", "--workspace", "ws1", "--index", "1"),
			Entry(desc, "netrule", "delete", "--user", "user1", "--workspace", "xxx", "--index", "1"),
			Entry(desc, "netrule", "delete", "--user", "user1", "--workspace", "ws1", "--index", "-1"),
			Entry(desc, "netrule", "delete", "--user", "user1", "--workspace", "ws1", "--index", "3"),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(args ...string) {
				testUtil.CreateWorkspace("user1", "ws1", "template1", nil)
				testUtil.UpsertNetworkRule("user1", "ws1", "nw1", 1111, "/", false, -1)
				clientMock.SetUpdateError("\\.RunE$", errors.New("mock update error"))
				run_test(args...)
			},
			Entry(desc, "netrule", "delete", "--user", "user1", "--workspace", "ws1", "--index", "1"),
		)
	})

	//==================================================================================
})
