package cmd

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"regexp"
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
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
)

var _ = Describe("cosmoctl [workspace]", func() {

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

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).ShouldNot(HaveOccurred())
				Expect(consoleOut()).To(MatchSnapShot())

				wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), args[2], args[4])
				Expect(err).NotTo(HaveOccurred()) // created

				wsSnap := struct{ Name, Namespace, Spec, Status interface{} }{
					Name:      wsv1Workspace.Name,
					Namespace: wsv1Workspace.Namespace,
					Spec:      wsv1Workspace.Spec,
					Status:    wsv1Workspace.Status,
				}
				Expect(wsSnap).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "create", "ws1", "--user", "user1", "--template", "template1", "--vars", "HOGE:HOGEHOGE"),
			Entry(desc, "workspace", "create", "ws1", "--user", "user1", "--template", "template1"),
		)

		DescribeTable("✅ success with dry-run:",
			func(args ...string) {
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).ShouldNot(HaveOccurred())
				o := consoleOut()
				o = regexp.MustCompile(`creationTimestamp: .+`).ReplaceAllString(o, "creationTimestamp: xxxxxxxx")
				o = regexp.MustCompile(`time: .+`).ReplaceAllString(o, "time: xxxxxxxx")
				o = regexp.MustCompile(`uid: .+`).ReplaceAllString(o, "uid: xxxxxxxx")
				Expect(o).To(MatchSnapShot())

				_, err = k8sClient.GetWorkspaceByUserID(context.Background(), args[2], args[4])
				Expect(err).To(HaveOccurred()) // not created
			},
			Entry(desc, "workspace", "create", "ws1", "--user", "user1", "--template", "template1", "--vars", "HOGE:HOGEHOGE", "--dry-run"),
		)

		DescribeTable("❌ fail with invalid args:",
			func(args ...string) {
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).Should(HaveOccurred())
				Expect(consoleOut()).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "create"),
			Entry(desc, "workspace", "create", "ws1", "--user", "user1", "--template", "template1", "--all-namespaces"),
			Entry(desc, "workspace", "create", "ws1", "--user", "xxxxx", "--template", "template1"),
			Entry(desc, "workspace", "create", "ws1", "--user", "user1", "--namespace", "user1", "--template", "template1"),
			Entry(desc, "workspace", "create", "ws1", "--namespace", "xxxx", "--template", "template1"),
			Entry(desc, "workspace", "create", "--user", "user1", "--template", "template1"),
			Entry(desc, "workspace", "create", "ws1", "--user", "--template", "template1"),
			Entry(desc, "workspace", "create", "ws1", "--user", "user1", "--template"),
			Entry(desc, "workspace", "create", "ws1", "--user", "xxxxx", "--template", "template1", "--dry-run"),
			Entry(desc, "workspace", "create", "ws1", "--user", "user1", "--template", "template1", "--vars", "HOGE"),
		)
	})

	//==================================================================================
	Describe("[get]", func() {

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_CreateWorkspace("user1", "ws2", "template1", nil)
				test_createNetworkRule("user1", "ws2", "nw1", 1111, "gp1", "/")
				test_createNetworkRule("user1", "ws2", "nw3", 2222, "gp1", "/")

				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).ShouldNot(HaveOccurred())
				o := consoleOut()
				o = regexp.MustCompile(`creationTimestamp: .+`).ReplaceAllString(o, "creationTimestamp: xxxxxxxx")
				o = regexp.MustCompile(`time: .+`).ReplaceAllString(o, "time: xxxxxxxx")
				o = regexp.MustCompile(`uid: .+`).ReplaceAllString(o, "uid: xxxxxxxx")
				o = regexp.MustCompile(`resourceVersion: .+`).ReplaceAllString(o, "resourceVersion: xxxxxxxx")
				Expect(o).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "get", "--user", "user1"),
			Entry(desc, "workspace", "get", "--user", "user1", "ws2"),
			Entry(desc, "workspace", "get", "--namespace", "cosmo-user-user1"),
			Entry(desc, "workspace", "get", "--namespace", "cosmo-user-user1", "ws2"),
			Entry(desc, "workspace", "get", "-A"),
			Entry(desc, "workspace", "get", "-A", "-o", "yaml"),
			Entry(desc, "workspace", "get", "-A", "-o", "wide"),
			Entry(desc, "workspace", "get", "-A", "--network"),
			Entry(desc, "workspace", "get", "-A", "--network", "-o", "yaml"),
			Entry(desc, "workspace", "get", "-A", "--network", "-o", "wide"),
		)

		DescribeTable("✅ success when workspace is empty:",
			func(args ...string) {
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).ShouldNot(HaveOccurred())
				o := consoleOut()
				o = regexp.MustCompile(`creationTimestamp: .+`).ReplaceAllString(o, "creationTimestamp: xxxxxxxx")
				o = regexp.MustCompile(`time: .+`).ReplaceAllString(o, "time: xxxxxxxx")
				o = regexp.MustCompile(`uid: .+`).ReplaceAllString(o, "uid: xxxxxxxx")
				o = regexp.MustCompile(`resourceVersion: .+`).ReplaceAllString(o, "resourceVersion: xxxxxxxx")
				Expect(o).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "get", "--user", "user1"),
			Entry(desc, "workspace", "get", "--namespace", "cosmo-user-user1"),
			Entry(desc, "workspace", "get", "--all-namespaces"),
			Entry(desc, "workspace", "get", "-A", "-o", "yaml"),
			Entry(desc, "workspace", "get", "-A", "-o", "wide"),
			Entry(desc, "workspace", "get", "-A", "--network"),
			Entry(desc, "workspace", "get", "-A", "--network", "-o", "yaml"),
			Entry(desc, "workspace", "get", "-A", "--network", "-o", "wide"),
		)

		DescribeTable("❌ fail with invalid args:",
			func(args ...string) {
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).Should(HaveOccurred())
				Expect(consoleOut()).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "get", "-A", "ws1"),
			Entry(desc, "workspace", "get", "--namespace", "cosmo-user-user1", "--user", "user1"),
			Entry(desc, "workspace", "get", "--namespace", "xxx"),
			Entry(desc, "workspace", "get", "-A", "--user", "user1"),
			Entry(desc, "workspace", "get", "-A", "-o", "xxxx"),
			Entry(desc, "workspace", "get", "--user", "user1", "xxx"),
			Entry(desc, "workspace", "get", "--user", "xxxx"),
		)

		DescribeTable("❌ fail with an unexpected error at list users:",
			func(args ...string) {
				clientMock.ListMock = func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (mocked bool, err error) {
					if clientMock.IsCallingFrom("\\.ListUsers$") {
						return true, errors.New("mock listUsers error")
					}
					return false, nil
				}
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).Should(HaveOccurred())
				Expect(consoleOut()).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "get", "-A"),
		)

		DescribeTable("❌ fail with an unexpected error at list workspace:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_CreateWorkspace("user1", "ws2", "template1", nil)
				clientMock.ListMock = func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (mocked bool, err error) {
					if clientMock.IsCallingFrom("\\.ListWorkspacesByUserID$") {
						return true, errors.New("mock listWorkspacesByUserID error")
					}
					return false, nil
				}
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).Should(HaveOccurred())
				Expect(consoleOut()).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "get", "-A"),
			Entry(desc, "workspace", "get", "--user", "user1"),
		)
	})

	//==================================================================================
	Describe("[delete]", func() {

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws2", "template1", nil)
				test_createNetworkRule("user1", "ws2", "nw1", 1111, "gp1", "/")
				test_createNetworkRule("user1", "ws2", "nw3", 2222, "gp1", "/")

				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).ShouldNot(HaveOccurred())
				Expect(consoleOut()).To(MatchSnapShot())

				_, err = k8sClient.GetWorkspaceByUserID(context.Background(), args[2], "user1")
				Expect(err).To(HaveOccurred()) // deleted
			},
			Entry(desc, "workspace", "delete", "ws2", "--user", "user1"),
			Entry(desc, "workspace", "delete", "ws2", "--namespace", "cosmo-user-user1"),
		)

		DescribeTable("✅ success with dry-run:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws2", "template1", nil)
				test_createNetworkRule("user1", "ws2", "nw1", 1111, "gp1", "/")
				test_createNetworkRule("user1", "ws2", "nw3", 2222, "gp1", "/")

				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).ShouldNot(HaveOccurred())
				Expect(consoleOut()).To(MatchSnapShot())

				_, err = k8sClient.GetWorkspaceByUserID(context.Background(), args[2], "user1")
				Expect(err).NotTo(HaveOccurred()) // undeleted
			},
			Entry(desc, "workspace", "delete", "ws2", "--dry-run", "--user", "user1"),
			Entry(desc, "workspace", "delete", "ws2", "--dry-run", "--namespace", "cosmo-user-user1"),
		)

		DescribeTable("❌ fail with invalid args:",
			func(args ...string) {
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).Should(HaveOccurred())
				Expect(consoleOut()).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "delete", "ws1", "--user", "user1", "-A"),
			Entry(desc, "workspace", "delete", "ws1", "--namespace", "cosmo-user-user1", "--user", "user1"),
			Entry(desc, "workspace", "delete", "ws1", "--namespace", "xxxx"),
			Entry(desc, "workspace", "delete"),
			Entry(desc, "workspace", "delete", "xxxx", "--user", "user1", "-A"),
			Entry(desc, "workspace", "delete", "ws1", "--user", "user1", "xxx"),
		)

		DescribeTable("❌ fail with an unexpected error at delete:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				clientMock.DeleteMock = func(ctx context.Context, obj client.Object, opts ...client.DeleteOption) (mocked bool, err error) {
					if clientMock.IsCallingFrom("\\.RunE$") {
						return true, errors.New("mock delete error")
					}
					return false, nil
				}
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).Should(HaveOccurred())
				Expect(consoleOut()).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "delete", "ws1", "--user", "user1"),
			Entry(desc, "workspace", "delete", "ws1", "--dry-run", "--user", "user1"),
		)
	})

	//==================================================================================
	Describe("[open-port]", func() {

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")
				test_createNetworkRule("user1", "ws1", "nw3", 2222, "gp1", "/")

				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).ShouldNot(HaveOccurred())
				Expect(consoleOut()).To(MatchSnapShot())

				wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), args[2], "user1")
				Expect(err).NotTo(HaveOccurred())
				Ω(wsv1Workspace.Spec).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw11", "--port", "3000", "--path", "/abc", "--group", "gp11"),
			Entry(desc, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw12", "--port", "4000", "--path", "/def"),
		)

		DescribeTable("❌ fail with invalid args:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")

				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).Should(HaveOccurred())
				Expect(consoleOut()).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw11", "--port", "3000", "--path", "/", "-A"),
			Entry(desc, "workspace", "open-port", "ws1", "--user", "user1", "--namespace", "cosmo-user-user1", "--name", "nw11", "--port", "3000", "--path", "/"),
			Entry(desc, "workspace", "open-port", "ws1", "--namespace", "xxxxx", "--name", "nw11", "--port", "3000", "--path", "/"),
			Entry(desc, "workspace", "open-port"),
			Entry(desc, "workspace", "open-port", "ws1", "--user", "user1", "--port", "3000", "--path", "/"),
			Entry(desc, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw11", "--path", "/"),
			Entry(desc, "workspace", "open-port", "ws1", "--user", "xxxxx", "--name", "nw11", "--port", "3000", "--path", "/"),
			Entry(desc, "workspace", "open-port", "xxx", "--user", "user1", "--name", "nw11", "--port", "3000", "--path", "/"),
			Entry(desc, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw11", "--port", "1111", "--path", "/"),
			Entry(desc, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw1", "--port", "1111", "--path", "/", "--group", "gp1"),
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
				rootCmd.SetArgs(args)
				err := rootCmd.Execute()
				Ω(err).Should(HaveOccurred())
				Expect(consoleOut()).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw12", "--port", "4000", "--path", "/def"),
		)
	})

	//==================================================================================
	Describe("[close-port]", func() {

		run_test := func(args ...string) {
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			Expect(consoleOut()).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")
				test_createNetworkRule("user1", "ws1", "nw2", 2222, "gp1", "/")
				run_test(args...)
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), args[2], "user1")
				Expect(err).NotTo(HaveOccurred())
				Ω(wsv1Workspace.Spec).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "close-port", "ws1", "--user", "user1", "--port-name", "nw1"),
		)

		DescribeTable("❌ fail with invalid args:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")
				run_test(args...)
			},
			Entry(desc, "workspace", "close-port", "ws1", "--user", "user1", "--port-name", "nw11", "-A"),
			Entry(desc, "workspace", "close-port", "ws1", "--user", "user1", "--namespace", "cosmo-user-user1", "--port-name", "nw11"),
			Entry(desc, "workspace", "close-port", "ws1", "--namespace", "xxxxx", "--port-name", "nw11"),
			Entry(desc, "workspace", "close-port"),
			Entry(desc, "workspace", "close-port", "ws1", "--user", "user1"),
			Entry(desc, "workspace", "close-port", "ws1", "--user", "xxxxx", "--port-name", "nw11"),
			Entry(desc, "workspace", "close-port", "xxx", "--user", "user1", "--port-name", "nw11"),
			Entry(desc, "workspace", "close-port", "ws1", "--user", "user1", "--port-name", "main"),
			Entry(desc, "workspace", "close-port", "ws1", "--user", "user1", "--port-name", "xxxx"),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")
				clientMock.SetUpdateError("\\.RunE$", errors.New("mock update error"))
				run_test(args...)
			},
			Entry(desc, "workspace", "close-port", "ws1", "--user", "user1", "--port-name", "nw1"),
		)
	})

	//==================================================================================
	Describe("[run-instance]", func() {

		run_test := func(args ...string) {
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			Expect(consoleOut()).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_StopWorkspace("user1", "ws1")
				run_test(args...)
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), args[2], "user1")
				Expect(err).NotTo(HaveOccurred())
				Ω(wsv1Workspace.Spec).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "run-instance", "ws1", "--user", "user1"),
		)

		DescribeTable("❌ fail with invalid args:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_StopWorkspace("user1", "ws1")
				test_CreateWorkspace("user1", "ws2", "template1", nil)
				run_test(args...)
			},
			Entry(desc, "workspace", "run-instance", "ws1", "--user", "user1", "-A"),
			Entry(desc, "workspace", "run-instance", "ws1", "--user", "user1", "--namespace", "cosmo-user-user1"),
			Entry(desc, "workspace", "run-instance", "ws1", "--namespace", "xxxxx"),
			Entry(desc, "workspace", "run-instance"),
			Entry(desc, "workspace", "run-instance", "ws1", "--user", "xxxxx"),
			Entry(desc, "workspace", "run-instance", "xxx", "--user", "user1"),
			Entry(desc, "workspace", "run-instance", "ws2", "--user", "user1"),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_StopWorkspace("user1", "ws1")
				clientMock.SetUpdateError("\\.RunE$", errors.New("mock update error"))
				run_test(args...)
			},
			Entry(desc, "workspace", "run-instance", "ws1", "--user", "user1"),
		)
	})

	//==================================================================================
	Describe("[stop-instance]", func() {

		run_test := func(args ...string) {
			By("---------------test start----------------")
			rootCmd.SetArgs(args)
			err := rootCmd.Execute()
			Expect(consoleOut()).To(MatchSnapShot())
			Ω(errSnap(err)).To(MatchSnapShot())
			By("---------------test end---------------")
		}

		DescribeTable("✅ success in normal context:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				run_test(args...)
				wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), args[2], "user1")
				Expect(err).NotTo(HaveOccurred())
				Ω(workspaceSnap(wsv1Workspace)).To(MatchSnapShot())
			},
			Entry(desc, "workspace", "stop-instance", "ws1", "--user", "user1"),
		)

		DescribeTable("❌ fail with invalid args:",
			func(args ...string) {
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				test_CreateWorkspace("user1", "ws2", "template1", nil)
				test_StopWorkspace("user1", "ws2")
				run_test(args...)
			},
			Entry(desc, "workspace", "stop-instance", "ws1", "--user", "user1", "-A"),
			Entry(desc, "workspace", "stop-instance", "ws1", "--user", "user1", "--namespace", "cosmo-user-user1"),
			Entry(desc, "workspace", "stop-instance", "ws1", "--namespace", "xxxxx"),
			Entry(desc, "workspace", "stop-instance"),
			Entry(desc, "workspace", "stop-instance", "ws1", "--user", "xxxxx"),
			Entry(desc, "workspace", "stop-instance", "xxx", "--user", "user1"),
			Entry(desc, "workspace", "stop-instance", "ws2", "--user", "user1"),
		)

		DescribeTable("❌ fail with an unexpected error at update:",
			func(args ...string) {
				clientMock.SetUpdateError("\\.RunE$", errors.New("mock update error"))
				test_CreateWorkspace("user1", "ws1", "template1", nil)
				run_test(args...)
			},
			Entry(desc, "workspace", "stop-instance", "ws1", "--user", "user1"),
		)
	})

})
