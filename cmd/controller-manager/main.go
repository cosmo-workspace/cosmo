package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	klog "k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	traefikv1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"

	"github.com/cosmo-workspace/cosmo/internal/controllers"
	"github.com/cosmo-workspace/cosmo/internal/webhooks"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/workspace"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	// +kubebuilder:scaffold:imports
)

const (
	controllerFieldManager string = "cosmo-instance-controller"
)

const (
	instController        string = "cosmo-instance-controller"
	clusterInstController string = "cosmo-cluster-instance-controller"
	tmplController        string = "cosmo-template-controller"
	clusterTmplController string = "cosmo-cluster-template-controller"
	userController        string = "cosmo-user-controller"
	wsController          string = "cosmo-workspace-controller"
	wsStatController      string = "cosmo-workspace-status-controller"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	o        = &option{}
)

type option struct {
	ZapOpts zap.Options

	Port                     int
	MetricsAddr              string
	ProbeAddr                string
	EnableLeaderElection     bool
	StatusCheckIntervals     int64
	CertDir                  string
	WorkspaceURLBaseProtocol string
	TraefikIngressRouteCfg   workspace.TraefikIngressRouteConfig
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme))
	utilruntime.Must(traefikv1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "controller-manager",
		Short: "cosmo controller manager",
		Long: `
cosmo controller manager
Complete documentation is available at http://github.com/cosmo-workspace/cosmo

MIT 2023 cosmo-workspace/cosmo
`,
		Version: "v0.9.0 cosmo-workspace 2023",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			ctrl.SetLogger(zap.New(zap.UseFlagOptions(&o.ZapOpts)))

			printVersion(cmd)
			printOptions()

			mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
				Scheme:                 scheme,
				MetricsBindAddress:     o.MetricsAddr,
				Port:                   o.Port,
				HealthProbeBindAddress: o.ProbeAddr,
				LeaderElection:         o.EnableLeaderElection,
				LeaderElectionID:       "04c57811.cosmo-workspace",
				CertDir:                o.CertDir,
			})
			if err != nil {
				setupLog.Error(err, "unable to start manager")
				os.Exit(1)
			}

			if err = (&controllers.InstanceReconciler{
				Client:   mgr.GetClient(),
				Recorder: mgr.GetEventRecorderFor(instController),
				Scheme:   mgr.GetScheme(),
			}).SetupWithManager(mgr, controllerFieldManager); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", instController)
				os.Exit(1)
			}
			if err = (&controllers.TemplateReconciler{
				Client:       mgr.GetClient(),
				Scheme:       mgr.GetScheme(),
				FieldManager: controllerFieldManager,
			}).SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", tmplController)
				os.Exit(1)
			}
			if err = (&controllers.ClusterInstanceReconciler{
				Client:   mgr.GetClient(),
				Recorder: mgr.GetEventRecorderFor(clusterInstController),
				Scheme:   mgr.GetScheme(),
			}).SetupWithManager(mgr, controllerFieldManager); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", clusterInstController)
				os.Exit(1)
			}
			if err = (&controllers.ClusterTemplateReconciler{
				Client:       mgr.GetClient(),
				Scheme:       mgr.GetScheme(),
				FieldManager: controllerFieldManager,
			}).SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", clusterTmplController)
				os.Exit(1)
			}
			if err = (&controllers.WorkspaceReconciler{
				Client:   mgr.GetClient(),
				Recorder: mgr.GetEventRecorderFor(wsController),
				Scheme:   mgr.GetScheme(),

				TraefikIngressRouteCfg: &o.TraefikIngressRouteCfg,
				URLBaseProtocol:        o.WorkspaceURLBaseProtocol,
			}).SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", wsController)
				os.Exit(1)
			}
			if err = (&controllers.WorkspaceStatusReconciler{
				Client:   mgr.GetClient(),
				Recorder: mgr.GetEventRecorderFor(wsStatController),
				Scheme:   mgr.GetScheme(),
			}).SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", wsStatController)
				os.Exit(1)
			}
			if err = (&controllers.UserReconciler{
				Client:   mgr.GetClient(),
				Recorder: mgr.GetEventRecorderFor(userController),
				Scheme:   mgr.GetScheme(),
			}).SetupWithManager(mgr); err != nil {
				setupLog.Error(err, "unable to create controller", "controller", userController)
				os.Exit(1)
			}
			// +kubebuilder:scaffold:builder

			// Webhook
			(&webhooks.InstanceMutationWebhookHandler{
				Client: mgr.GetClient(),
				Log:    clog.NewLogger(ctrl.Log.WithName("InstanceMutationWebhook")),
			}).SetupWebhookWithManager(mgr)
			(&webhooks.InstanceValidationWebhookHandler{
				Client:       mgr.GetClient(),
				Log:          clog.NewLogger(ctrl.Log.WithName("InstanceValidationWebhook")),
				FieldManager: controllerFieldManager,
			}).SetupWebhookWithManager(mgr)

			(&webhooks.WorkspaceMutationWebhookHandler{
				Client: mgr.GetClient(),
				Log:    clog.NewLogger(ctrl.Log.WithName("WorkspaceMutationWebhook")),
			}).SetupWebhookWithManager(mgr)
			(&webhooks.WorkspaceValidationWebhookHandler{
				Client: mgr.GetClient(),
				Log:    clog.NewLogger(ctrl.Log.WithName("WorkspaceValidationWebhook")),
			}).SetupWebhookWithManager(mgr)

			(&webhooks.UserMutationWebhookHandler{
				Client: mgr.GetClient(),
				Log:    clog.NewLogger(ctrl.Log.WithName("UserMutationWebhook")),
			}).SetupWebhookWithManager(mgr)
			(&webhooks.UserValidationWebhookHandler{
				Client: mgr.GetClient(),
				Log:    clog.NewLogger(ctrl.Log.WithName("UserValidationWebhook")),
			}).SetupWebhookWithManager(mgr)

			ctx := ctrl.SetupSignalHandler()

			if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
				setupLog.Error(err, "unable to set up health check")
				os.Exit(1)
			}
			if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
				setupLog.Error(err, "unable to set up ready check")
				os.Exit(1)
			}

			// Start controller manager
			setupLog.Info("starting manager")
			if err := mgr.Start(ctx); err != nil {
				setupLog.Error(err, "problem running manager")
				os.Exit(1)
			}
			return nil
		},
	}

	goflags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	klog.InitFlags(goflags)
	o.ZapOpts.BindFlags(goflags)
	ctrl.RegisterFlags(goflags)
	rootCmd.PersistentFlags().AddGoFlagSet(goflags)

	rootCmd.PersistentFlags().IntVar(&o.Port, "port", 9443, "Port for webhook server")
	rootCmd.PersistentFlags().Int64Var(&o.StatusCheckIntervals, "statuscheck-interval-seconds", 5, "Status check interval seconds")
	rootCmd.PersistentFlags().StringVar(&o.MetricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	rootCmd.PersistentFlags().StringVar(&o.ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	rootCmd.PersistentFlags().StringVar(&o.CertDir, "cert-dir", "/tmp/k8s-webhook-server/serving-certs", "Certificate dir. The server key and certificate must be named tls.key and tls.crt")
	rootCmd.PersistentFlags().StringVar(&o.WorkspaceURLBaseProtocol, "workspace-urlbase-protocol", "https", "http or https")
	rootCmd.PersistentFlags().StringVar(&o.TraefikIngressRouteCfg.HostBase, "workspace-urlbase-host", "{{NETRULE}}-{{WORKSPACE}}-{{USER}}", "host template. {{NETRULE}}, {{WORKSPACE}} and {{USER}} are replaced for each URL. you can customize like `{{NETRULE}}-{{WORKSPACE}}-{{USER}}-k3d`")
	rootCmd.PersistentFlags().StringVar(&o.TraefikIngressRouteCfg.Domain, "workspace-urlbase-domain", "example.com", "domain for workspace url")
	rootCmd.PersistentFlags().BoolVar(&o.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	rootCmd.PersistentFlags().StringSliceVar(&o.TraefikIngressRouteCfg.Entrypoints, "traefik-entrypoints", []string{"web"}, "Traefik ingress entrypoint")
	rootCmd.PersistentFlags().StringVar(&o.TraefikIngressRouteCfg.AuthenMiddleware.Name, "traefik-authen-middleware", "cosmo-auth", "Traefik authen middleware")
	rootCmd.PersistentFlags().StringVar(&o.TraefikIngressRouteCfg.AuthenMiddleware.Namespace, "traefik-authen-middleware-namespace", "cosmo-system", "Traefik authen middleware namespace")
	rootCmd.PersistentFlags().StringVar(&o.TraefikIngressRouteCfg.UserNameHeaderMiddleware.Name, "traefik-username-header-middleware", "cosmo-username-headers", "Traefik username header middleware")

	if err := rootCmd.Execute(); err != nil {
		setupLog.Error(err, "problem executing command")
		os.Exit(1)
	}
}

func printOptions() {
	rv := reflect.ValueOf(*o)
	rt := rv.Type()
	options := make([]interface{}, rt.NumField()*2)

	for i := 0; i < rt.NumField(); i++ {
		options[i*2] = rt.Field(i).Name
		options[i*2+1] = rv.Field(i).Interface()
	}

	setupLog.Info("options", options...)
}

func printVersion(cmd *cobra.Command) {
	fmt.Fprintf(cmd.OutOrStdout(), "cosmo-controller-manager - cosmo %s\n", cmd.Version)
}
