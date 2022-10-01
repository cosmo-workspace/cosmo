package main

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/internal/dashboard"
	"github.com/cosmo-workspace/cosmo/pkg/auth"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

const (
	fieldManager string = "cosmo-dashboard"
)

var (
	o        = &options{}
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = cosmov1alpha1.AddToScheme(scheme)
	_ = wsv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type options struct {
	StaticFileDir           string
	ResponseTimeoutSeconds  int64
	GracefulShutdownSeconds int64
	TLSPrivateKeyPath       string
	TLSCertPath             string
	Insecure                bool
	ServerPort              int
	MaxAgeMinutes           int
}

func main() {
	flag.Int64Var(&o.ResponseTimeoutSeconds, "timeout-seconds", 3, "Timeout seconds for response")
	flag.Int64Var(&o.GracefulShutdownSeconds, "graceful-shutdown-seconds", 10, "Graceful shutdown seconds")
	flag.StringVar(&o.StaticFileDir, "serve-dir", "/app/public", "Static file dir to serve")
	flag.StringVar(&o.TLSPrivateKeyPath, "tls-key", "tls.key", "TLS key file path")
	flag.StringVar(&o.TLSCertPath, "tls-cert", "tls.crt", "TLS certificate file path")
	flag.BoolVar(&o.Insecure, "insecure", false, "start http server not https server")
	flag.IntVar(&o.ServerPort, "port", 8443, "Port for dashboard server")
	flag.IntVar(&o.MaxAgeMinutes, "maxage-minutes", 720, "session maxage minutes")

	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	printVersion()
	printOptions()

	ctx := ctrl.SetupSignalHandler()

	// Setup controller manager
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
		Port:               9443,
		LeaderElection:     false,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Setup server
	klient := kosmo.NewClient(mgr.GetClient())

	auths := make(map[wsv1alpha1.UserAuthType]auth.Authorizer)
	auths[wsv1alpha1.UserAuthTypePasswordSecert] = auth.NewPasswordSecretAuthorizer(klient)

	serv := (&dashboard.Server{
		Log:                 clog.NewLogger(ctrl.Log.WithName("dashboard")),
		Klient:              klient,
		GracefulShutdownDur: time.Second * time.Duration(o.GracefulShutdownSeconds),
		ResponseTimeout:     time.Second * time.Duration(o.ResponseTimeoutSeconds),
		StaticFileDir:       o.StaticFileDir,
		Port:                o.ServerPort,
		MaxAgeSeconds:       60 * o.MaxAgeMinutes,
		SessionName:         "cosmo-dashboard",
		TLSPrivateKeyPath:   o.TLSPrivateKeyPath,
		TLSCertPath:         o.TLSCertPath,
		Insecure:            o.Insecure,
		Authorizers:         auths,
	})
	if err := mgr.Add(serv); err != nil {
		setupLog.Error(err, "failed to add server to controller-manager")
		os.Exit(1)
	}

	// Start server
	setupLog.Info("Start controller manager")
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
	fmt.Println("cosmo-dashboard - cosmo v0.7.0 cosmo-workspace 2021")
}
