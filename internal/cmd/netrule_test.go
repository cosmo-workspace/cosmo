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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
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
		_ = clientgoscheme.AddToScheme(scheme)
		_ = cosmov1alpha1.AddToScheme(scheme)
		_ = wsv1alpha1.AddToScheme(scheme)

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

		test_CreateLoginUser("user2", "お名前", "", "password")
		test_CreateLoginUser("user1", "アドミン", wsv1alpha1.UserAdminRole, "password")
		test_CreateTemplate(wsv1alpha1.TemplateTypeWorkspace, "template1")
		By("---------------BeforeEach end----------------")
	})

	AfterEach(func() {
		By("---------------AfterEach start---------------")
		clientMock.Clear()
		test_DeleteWorkspaceAll()
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

	workspaceSnap := func(ws *wsv1alpha1.Workspace) struct{ Name, Namespace, Spec, Status interface{} } {
		return struct{ Name, Namespace, Spec, Status interface{} }{
			Name:      ws.Name,
			Namespace: ws.Namespace,
			Spec:      ws.Spec,
			Status:    ws.Status,
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
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserName(context.Background(), args[6], "user1")
				Expect(err).NotTo(HaveOccurred())
				Ω(workspaceSnap(wsv1Workspace)).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")
				test_createNetworkRule("user1", "ws1", "nw3", 2222, "gp2", "/")
				run_test(args...)
			},
			Entry(desc, "netrule", "create", "nw11", "--user", "user1", "--workspace", "ws1", "--port", "3000", "--path", "/abc", "--group", "gp11"),
			Entry(desc, "netrule", "create", "nw12", "--user", "user1", "--workspace", "ws1", "--port", "4000", "--path", "/def"),
			Entry(desc, "netrule", "create", "nw12", "--user", "user1", "--workspace", "ws1", "--port", "4000", "--path", "/def"),
		)

		DescribeTable("❌ fail with invalid args:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")
				run_test(args...)
			},
			Entry(desc, "netrule", "create", "nw11", "--user", "user1", "--workspace", "ws1", "--port", "3000", "--path", "/", "-A"),
			Entry(desc, "netrule", "create", "nw11", "--user", "user1", "--namespace", "cosmo-user-user1", "--workspace", "ws1", "--port", "3000", "--path", "/"),
			Entry(desc, "netrule", "create", "nw11", "--namespace", "xxxxx", "--workspace", "ws1", "--port", "3000", "--path", "/"),
			Entry(desc, "netrule", "create"),
			Entry(desc, "netrule", "create", "nw11", "--user", "user1", "--port", "3000", "--path", "/"),
			Entry(desc, "netrule", "create", "nw11", "--user", "user1", "--workspace", "ws1", "--path", "/"),
			Entry(desc, "netrule", "create", "nw11", "--user", "xxxxx", "--workspace", "ws1", "--port", "3000", "--path", "/"),
			Entry(desc, "netrule", "create", "nw11", "--user", "user1", "--workspace", "xxx", "--port", "3000", "--path", "/"),
			Entry(desc, "netrule", "create", "nw11", "--user", "user1", "--workspace", "ws1", "--port", "1111", "--path", "/", "--group", "gp1"),
			Entry(desc, "netrule", "create", "nw1", "--user", "user1", "--workspace", "ws1", "--port", "1111", "--path", "/", "--group", "gp1"),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				clientMock.UpdateMock = func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) (mocked bool, err error) {
					if clientMock.IsCallingFrom("\\.RunE$") {
						return true, errors.New("mock update error")
					}
					return false, nil
				}
				run_test(args...)
			},
			Entry(desc, "netrule", "create", "ws1", "--user", "user1", "--workspace", "nw12", "--port", "4000", "--path", "/def"),
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
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserName(context.Background(), args[2], "user1")
				Expect(err).NotTo(HaveOccurred())
				Ω(workspaceSnap(wsv1Workspace)).To(MatchSnapShot())
			}
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")
				test_createNetworkRule("user1", "ws1", "nw2", 2222, "gp2", "/")
				run_test(args...)
			},
			Entry(desc, "netrule", "delete", "ws1", "--user", "user1", "--workspace", "nw1"),
			Entry(desc, "netrule", "rm-net", "ws1", "--user", "user1", "--workspace", "nw1"),
		)

		DescribeTable("❌ fail with invalid args:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")
				run_test(args...)
			},
			Entry(desc, "netrule", "delete", "nw11", "--user", "user1", "--workspace", "ws1", "-A"),
			Entry(desc, "netrule", "delete", "nw11", "--user", "user1", "--namespace", "cosmo-user-user1", "--workspace", "ws1"),
			Entry(desc, "netrule", "delete", "nw11", "--namespace", "xxxxx", "--workspace", "ws1"),
			Entry(desc, "netrule", "delete"),
			Entry(desc, "netrule", "delete", "nw11", "--user", "user1"),
			Entry(desc, "netrule", "delete", "nw11", "--user", "xxxxx", "--workspace", "ws1"),
			Entry(desc, "netrule", "delete", "nw11", "--user", "user1", "--workspace", "xxx"),
			Entry(desc, "netrule", "delete", "main", "--user", "user1", "--workspace", "ws1"),
			Entry(desc, "netrule", "delete", "xxxx", "--user", "user1", "--workspace", "ws1"),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")
				clientMock.SetUpdateError("\\.RunE$", errors.New("mock update error"))
				run_test(args...)
			},
			Entry(desc, "netrule", "delete", "nw1", "--user", "user1", "--workspace", "ws1"),
		)
	})

	//==================================================================================
})
