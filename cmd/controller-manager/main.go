package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/cosmo-workspace/cosmo/internal/controllers"
	"github.com/cosmo-workspace/cosmo/internal/webhooks"
	"github.com/cosmo-workspace/cosmo/pkg/clog"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
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
	MetricsAddr             string
	ProbeAddr               string
	EnableLeaderElection    bool
	StatusCheckIntervals    int64
	CertDir                 string
	WorkspaceDefaultURLBase string
}

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = cosmov1alpha1.AddToScheme(scheme)
	_ = wsv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	flag.Int64Var(&o.StatusCheckIntervals, "statuscheck-interval-seconds", 5, "Status check interval seconds")
	flag.StringVar(&o.MetricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&o.ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&o.CertDir, "cert-dir", "/tmp/k8s-webhook-server/serving-certs", "Certificate dir. The server key and certificate must be named tls.key and tls.crt")
	flag.StringVar(&o.WorkspaceDefaultURLBase, "workspace-default-urlbase", "https://{{NETRULE_GROUP}}-{{INSTANCE}}-{{NAMESPACE}}.example.com", "workspace default urlbase. used if not set on template")
	flag.BoolVar(&o.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	printVersion()
	printOptions()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     o.MetricsAddr,
		Port:                   9443,
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
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
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
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", clusterTmplController)
		os.Exit(1)
	}
	if err = (&controllers.WorkspaceReconciler{
		Client:   mgr.GetClient(),
		Recorder: mgr.GetEventRecorderFor(wsController),
		Scheme:   mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", wsController)
		os.Exit(1)
	}
	if err = (&controllers.WorkspaceStatusReconciler{
		Client:         mgr.GetClient(),
		Recorder:       mgr.GetEventRecorderFor(wsStatController),
		Scheme:         mgr.GetScheme(),
		DefaultURLBase: o.WorkspaceDefaultURLBase,
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
		Log:    clog.NewLogger(ctrl.Log.WithName("InstanceMutationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)
	(&webhooks.InstanceValidationWebhookHandler{
		Client:       mgr.GetClient(),
		Log:          clog.NewLogger(ctrl.Log.WithName("InstanceValidationWebhookHandler")),
		FieldManager: controllerFieldManager,
	}).SetupWebhookWithManager(mgr)

	(&webhooks.TemplateMutationWebhookHandler{
		Client:         mgr.GetClient(),
		Log:            clog.NewLogger(ctrl.Log.WithName("TemplateMutationWebhookHandler")),
		DefaultURLBase: o.WorkspaceDefaultURLBase,
	}).SetupWebhookWithManager(mgr)
	(&webhooks.TemplateValidationWebhookHandler{
		Client:       mgr.GetClient(),
		Log:          clog.NewLogger(ctrl.Log.WithName("TemplateValidationWebhookHandler")),
		FieldManager: controllerFieldManager,
	}).SetupWebhookWithManager(mgr)

	(&webhooks.WorkspaceMutationWebhookHandler{
		Client: mgr.GetClient(),
		Log:    clog.NewLogger(ctrl.Log.WithName("WorkspaceMutationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)
	(&webhooks.WorkspaceValidationWebhookHandler{
		Client: mgr.GetClient(),
		Log:    clog.NewLogger(ctrl.Log.WithName("WorkspaceValidationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)

	(&webhooks.UserMutationWebhookHandler{
		Client: mgr.GetClient(),
		Log:    clog.NewLogger(ctrl.Log.WithName("UserMutationWebhookHandler")),
	}).SetupWebhookWithManager(mgr)
	(&webhooks.UserValidationWebhookHandler{
		Client: mgr.GetClient(),
		Log:    clog.NewLogger(ctrl.Log.WithName("UserValidationWebhookHandler")),
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

func printVersion() {
	fmt.Println("cosmo-controller-manager - cosmo v0.4.0 cosmo-workspace 2021")
}
