package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/cosmo-workspace/cosmo/internal/authproxy"
	"github.com/cosmo-workspace/cosmo/internal/authproxy/proxy"
	"github.com/cosmo-workspace/cosmo/pkg/auth"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	o        = &option{}
	eo       = &envOption{
		Workspace: os.Getenv(authproxy.EnvInstance),
		Namespace: os.Getenv(authproxy.EnvNamespace),
	}
)

type option struct {
	Port                    int
	Insecure                bool
	TLSPrivateKeyPath       string
	TLSCertPath             string
	BackendScheme           string
	GracefulShutdownSeconds int64
	MaxAgeMinutes           int
}

type envOption struct {
	Workspace string
	Namespace string
}

func (eo *envOption) Validate() error {
	if eo.Workspace == "" {
		return errors.New("env Workspace not found")
	}
	if eo.Namespace == "" {
		return errors.New("env Namespace not found")
	}
	return nil
}

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = cosmov1alpha1.AddToScheme(scheme)
	_ = wsv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	flag.IntVar(&o.Port, "port", 9443, "port for controller manager")
	flag.StringVar(&o.BackendScheme, "backend-scheme", "http", "proxy backend scheme. http or https")
	flag.StringVar(&o.TLSPrivateKeyPath, "tls-key", "tls.key", "TLS key file path")
	flag.StringVar(&o.TLSCertPath, "tls-cert", "tls.crt", "TLS certificate file path")
	flag.BoolVar(&o.Insecure, "insecure", false, "start http server not https server")
	flag.Int64Var(&o.GracefulShutdownSeconds, "graceful-shutdown-seconds", 10, "proxy graceful shutdown seconds")
	flag.IntVar(&o.MaxAgeMinutes, "maxage-minutes", 720, "session maxage minutes. if 0, session will be never expired")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	printVersion()
	printOptions()

	if err := eo.Validate(); err != nil {
		setupLog.Error(err, "validation failed")
		os.Exit(1)
	}

	// Setup controller manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
		Port:               o.Port,
		LeaderElection:     false,
		Namespace:          eo.Namespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to initialize controller manager")
		os.Exit(1)
	}

	ownerID := wsv1alpha1.UserIDByNamespace(eo.Namespace)
	if ownerID == "" {
		setupLog.Error(fmt.Errorf("namespace %s is not cosmo user's namespace", eo.Namespace), "invalid namespace")
		os.Exit(1)
	}

	// Setup proxy manager
	logger := clog.NewLogger(ctrl.Log.WithName("proxy-manager"))
	klient := kosmo.NewClient(mgr.GetClient())

	// only support KosmoSecert authorizer for now
	authorizer := auth.NewKosmoSecretAuthorizer(klient)

	proxyManager, err := (&proxy.Manager{
		Log:                      logger,
		ProxyBackendScheme:       o.BackendScheme,
		ProxyGracefulShutdownDur: time.Second * time.Duration(o.GracefulShutdownSeconds),
		ProxyStartupCheckTimeout: time.Second * time.Duration(10),
		Insecure:                 o.Insecure,
		TLSCertPath:              o.TLSCertPath,
		TLSPrivateKeyPath:        o.TLSPrivateKeyPath,
		User:                     ownerID,
		MaxAgeSeconds:            60 * o.MaxAgeMinutes,
		Authorizer:               authorizer,
	}).Initialize()
	if err != nil {
		setupLog.Error(err, "unable to initialize proxy manager")
		os.Exit(1)
	}

	// Setup instance network reconciler
	if err = (&authproxy.NetworkRuleReconciler{
		Client:        klient,
		Recorder:      mgr.GetEventRecorderFor("cosmo-auth-proxy"),
		Scheme:        mgr.GetScheme(),
		ProxyManager:  proxyManager,
		WorkspaceName: eo.Workspace,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NetworkRuleReconciler")
		os.Exit(1)
	}

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
	fmt.Println("cosmo-auth-proxy - cosmo v0.2.0-rc1 cosmo-workspace 2021")
}
