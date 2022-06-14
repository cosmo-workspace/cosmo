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
	Describe("[create]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
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
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "create", "ws1", "--user", "user1", "--template", "template1", "--vars", "HOGE:HOGEHOGE"),
				Entry(nil, "workspace", "create", "ws1", "--user", "user1", "--template", "template1"),
			)

			DescribeTable("with dry-run:",
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
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "create", "ws1", "--user", "user1", "--template", "template1", "--vars", "HOGE:HOGEHOGE", "--dry-run"),
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
				Entry(nil, "workspace", "create"),
				Entry(nil, "workspace", "create", "ws1", "--user", "user1", "--template", "template1", "--all-namespaces"),
				Entry(nil, "workspace", "create", "ws1", "--user", "xxxxx", "--template", "template1"),
				Entry(nil, "workspace", "create", "ws1", "--user", "user1", "--namespace", "user1", "--template", "template1"),
				Entry(nil, "workspace", "create", "ws1", "--namespace", "xxxx", "--template", "template1"),
				Entry(nil, "workspace", "create", "--user", "user1", "--template", "template1"),
				Entry(nil, "workspace", "create", "ws1", "--user", "--template", "template1"),
				Entry(nil, "workspace", "create", "ws1", "--user", "user1", "--template"),
				Entry(nil, "workspace", "create", "ws1", "--user", "xxxxx", "--template", "template1", "--dry-run"),
				Entry(nil, "workspace", "create", "ws1", "--user", "user1", "--template", "template1", "--vars", "HOGE"),
			)
		})
	})

	//==================================================================================
	Describe("[get]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					test_CreateWorkspace("user1", "ws2", "template1", nil)
					test_createNetworkRule("user1", "ws2", "nw1", 1111, "gp1", "/")
					test_createNetworkRule("user1", "ws2", "nw3", 2222, "gp1", "/")
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
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "get", "--user", "user1"),
				Entry(nil, "workspace", "get", "--user", "user1", "ws2"),
				Entry(nil, "workspace", "get", "--namespace", "cosmo-user-user1"),
				Entry(nil, "workspace", "get", "--namespace", "cosmo-user-user1", "ws2"),
				Entry(nil, "workspace", "get", "-A"),
				Entry(nil, "workspace", "get", "-A", "-o", "yaml"),
				Entry(nil, "workspace", "get", "-A", "-o", "wide"),
				Entry(nil, "workspace", "get", "-A", "--network"),
				Entry(nil, "workspace", "get", "-A", "--network", "-o", "yaml"),
				Entry(nil, "workspace", "get", "-A", "--network", "-o", "wide"),
			)

			DescribeTable("when workspace is empty:",
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
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "get", "--user", "user1"),
				Entry(nil, "workspace", "get", "--namespace", "cosmo-user-user1"),
				Entry(nil, "workspace", "get", "--all-namespaces"),
				Entry(nil, "workspace", "get", "-A", "-o", "yaml"),
				Entry(nil, "workspace", "get", "-A", "-o", "wide"),
				Entry(nil, "workspace", "get", "-A", "--network"),
				Entry(nil, "workspace", "get", "-A", "--network", "-o", "yaml"),
				Entry(nil, "workspace", "get", "-A", "--network", "-o", "wide"),
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
				Entry(nil, "workspace", "get", "-A", "ws1"),
				Entry(nil, "workspace", "get", "--namespace", "cosmo-user-user1", "--user", "user1"),
				Entry(nil, "workspace", "get", "--namespace", "xxx"),
				Entry(nil, "workspace", "get", "-A", "--user", "user1"),
				Entry(nil, "workspace", "get", "-A", "-o", "xxxx"),
				Entry(nil, "workspace", "get", "--user", "user1", "xxx"),
				Entry(nil, "workspace", "get", "--user", "xxxx"),
			)

			DescribeTable("with an unexpected error at list users:",
				func(args ...string) {
					clientMock.ListMock = func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.ListUsers$") {
							return true, errors.New("ListUsers error")
						}
						return false, nil
					}
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "get", "-A"),
			)

			DescribeTable("with an unexpected error at list workspace:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					test_CreateWorkspace("user1", "ws2", "template1", nil)
					clientMock.ListMock = func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.ListWorkspacesByUserID$") {
							return true, errors.New("ListWorkspacesByUserID error")
						}
						return false, nil
					}
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "get", "-A"),
				Entry(nil, "workspace", "get", "--user", "user1"),
			)

		})
	})

	//==================================================================================
	Describe("[delete]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws2", "template1", nil)
					test_createNetworkRule("user1", "ws2", "nw1", 1111, "gp1", "/")
					test_createNetworkRule("user1", "ws2", "nw3", 2222, "gp1", "/")
					test_createNetworkRule("user1", "ws2", "nw3", 2222, "gp1", "/")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())

					_, err = k8sClient.GetWorkspaceByUserID(context.Background(), args[2], "user1")
					Expect(err).To(HaveOccurred()) // deleted
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "delete", "ws2", "--user", "user1"),
				Entry(nil, "workspace", "delete", "ws2", "--namespace", "cosmo-user-user1"),
			)

			DescribeTable("with dry-run:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws2", "template1", nil)
					test_createNetworkRule("user1", "ws2", "nw1", 1111, "gp1", "/")
					test_createNetworkRule("user1", "ws2", "nw3", 2222, "gp1", "/")
					test_createNetworkRule("user1", "ws2", "nw3", 2222, "gp1", "/")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())

					_, err = k8sClient.GetWorkspaceByUserID(context.Background(), args[2], "user1")
					Expect(err).NotTo(HaveOccurred()) // undeleted
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "delete", "ws2", "--dry-run", "--user", "user1"),
				Entry(nil, "workspace", "delete", "ws2", "--dry-run", "--namespace", "cosmo-user-user1"),
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
				Entry(nil, "workspace", "delete", "ws1", "--user", "user1", "-A"),
				Entry(nil, "workspace", "delete", "ws1", "--namespace", "cosmo-user-user1", "--user", "user1"),
				Entry(nil, "workspace", "delete", "ws1", "--namespace", "xxxx"),
				Entry(nil, "workspace", "delete"),
				Entry(nil, "workspace", "delete", "xxxx", "--user", "user1", "-A"),
				Entry(nil, "workspace", "delete", "ws1", "--user", "user1", "xxx"),
			)

			DescribeTable("with an unexpected error at delete:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					clientMock.DeleteMock = func(ctx context.Context, obj client.Object, opts ...client.DeleteOption) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.RunE$") {
							return true, errors.New("Delete error")
						}
						return false, nil
					}
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "delete", "ws1", "--user", "user1"),
				Entry(nil, "workspace", "delete", "ws1", "--dry-run", "--user", "user1"),
			)

		})
	})

	//==================================================================================
	Describe("[open-port]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
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
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw11", "--port", "3000", "--path", "/abc", "--group", "gp11"),
				Entry(nil, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw12", "--port", "4000", "--path", "/def"),
			)
		})

		Describe("fail", func() {

			DescribeTable("with invalid args:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw11", "--port", "3000", "--path", "/", "-A"),
				Entry(nil, "workspace", "open-port", "ws1", "--user", "user1", "--namespace", "cosmo-user-user1", "--name", "nw11", "--port", "3000", "--path", "/"),
				Entry(nil, "workspace", "open-port", "ws1", "--namespace", "xxxxx", "--name", "nw11", "--port", "3000", "--path", "/"),
				Entry(nil, "workspace", "open-port"),
				Entry(nil, "workspace", "open-port", "ws1", "--user", "user1", "--port", "3000", "--path", "/"),
				Entry(nil, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw11", "--path", "/"),
				Entry(nil, "workspace", "open-port", "ws1", "--user", "xxxxx", "--name", "nw11", "--port", "3000", "--path", "/"),
				Entry(nil, "workspace", "open-port", "xxx", "--user", "user1", "--name", "nw11", "--port", "3000", "--path", "/"),
				Entry(nil, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw11", "--port", "1111", "--path", "/"),
				Entry(nil, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw1", "--port", "1111", "--path", "/", "--group", "gp1"),
			)

			DescribeTable("with an unexpected error at update:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					clientMock.UpdateMock = func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.RunE$") {
							return true, errors.New("update error")
						}
						return false, nil
					}
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "open-port", "ws1", "--user", "user1", "--name", "nw12", "--port", "4000", "--path", "/def"),
			)
		})
	})

	//==================================================================================
	Describe("[close-port]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")
					test_createNetworkRule("user1", "ws1", "nw2", 2222, "gp1", "/")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())

					wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), args[2], "user1")
					Expect(err).NotTo(HaveOccurred())
					Ω(wsv1Workspace.Spec).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "close-port", "ws1", "--user", "user1", "--port-name", "nw1"),
			)
		})

		Describe("fail", func() {

			DescribeTable("with invalid args:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "close-port", "ws1", "--user", "user1", "--port-name", "nw11", "-A"),
				Entry(nil, "workspace", "close-port", "ws1", "--user", "user1", "--namespace", "cosmo-user-user1", "--port-name", "nw11"),
				Entry(nil, "workspace", "close-port", "ws1", "--namespace", "xxxxx", "--port-name", "nw11"),
				Entry(nil, "workspace", "close-port"),
				Entry(nil, "workspace", "close-port", "ws1", "--user", "user1"),
				Entry(nil, "workspace", "close-port", "ws1", "--user", "xxxxx", "--port-name", "nw11"),
				Entry(nil, "workspace", "close-port", "xxx", "--user", "user1", "--port-name", "nw11"),
				Entry(nil, "workspace", "close-port", "ws1", "--user", "user1", "--port-name", "main"),
				Entry(nil, "workspace", "close-port", "ws1", "--user", "user1", "--port-name", "xxxx"),
			)

			DescribeTable("with an unexpected error at update:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					test_createNetworkRule("user1", "ws1", "nw1", 1111, "gp1", "/")
					clientMock.UpdateMock = func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.RunE$") {
							return true, errors.New("update error")
						}
						return false, nil
					}
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "close-port", "ws1", "--user", "user1", "--port-name", "nw1"),
			)
		})
	})

	//==================================================================================
	Describe("[run-instance]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					test_StopWorkspace("user1", "ws1")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())

					wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), args[2], "user1")
					Expect(err).NotTo(HaveOccurred())
					Ω(wsv1Workspace.Spec).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "run-instance", "ws1", "--user", "user1"),
			)
		})

		Describe("fail", func() {

			DescribeTable("with invalid args:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					test_StopWorkspace("user1", "ws1")
					test_CreateWorkspace("user1", "ws2", "template1", nil)

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "run-instance", "ws1", "--user", "user1", "-A"),
				Entry(nil, "workspace", "run-instance", "ws1", "--user", "user1", "--namespace", "cosmo-user-user1"),
				Entry(nil, "workspace", "run-instance", "ws1", "--namespace", "xxxxx"),
				Entry(nil, "workspace", "run-instance"),
				Entry(nil, "workspace", "run-instance", "ws1", "--user", "xxxxx"),
				Entry(nil, "workspace", "run-instance", "xxx", "--user", "user1"),
				Entry(nil, "workspace", "run-instance", "ws2", "--user", "user1"),
			)

			DescribeTable("with an unexpected error at update:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					test_StopWorkspace("user1", "ws1")
					clientMock.UpdateMock = func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.RunE$") {
							return true, errors.New("update error")
						}
						return false, nil
					}
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "run-instance", "ws1", "--user", "user1"),
			)
		})
	})

	//==================================================================================
	Describe("[stop-instance]", func() {

		Describe("succeed", func() {

			DescribeTable("in normal context:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).ShouldNot(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())

					wsv1Workspace, err := k8sClient.GetWorkspaceByUserID(context.Background(), args[2], "user1")
					Expect(err).NotTo(HaveOccurred())
					Ω(wsv1Workspace.Spec).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "stop-instance", "ws1", "--user", "user1"),
			)
		})

		Describe("fail", func() {

			DescribeTable("with invalid args:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					test_CreateWorkspace("user1", "ws2", "template1", nil)
					test_StopWorkspace("user1", "ws2")

					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "stop-instance", "ws1", "--user", "user1", "-A"),
				Entry(nil, "workspace", "stop-instance", "ws1", "--user", "user1", "--namespace", "cosmo-user-user1"),
				Entry(nil, "workspace", "stop-instance", "ws1", "--namespace", "xxxxx"),
				Entry(nil, "workspace", "stop-instance"),
				Entry(nil, "workspace", "stop-instance", "ws1", "--user", "xxxxx"),
				Entry(nil, "workspace", "stop-instance", "xxx", "--user", "user1"),
				Entry(nil, "workspace", "stop-instance", "ws2", "--user", "user1"),
			)

			DescribeTable("with an unexpected error at update:",
				func(args ...string) {
					test_CreateWorkspace("user1", "ws1", "template1", nil)
					clientMock.UpdateMock = func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) (mocked bool, err error) {
						if clientMock.IsCallingFrom("\\.RunE$") {
							return true, errors.New("update error")
						}
						return false, nil
					}
					rootCmd.SetArgs(args)
					err := rootCmd.Execute()
					Ω(err).Should(HaveOccurred())
					Expect(consoleOut()).To(MatchSnapShot())
				},
				func(args ...string) string { return strings.Join(args, " ") },
				Entry(nil, "workspace", "stop-instance", "ws1", "--user", "user1"),
			)
		})
	})

})
