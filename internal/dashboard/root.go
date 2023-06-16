/*
Copyright © 2023 NAME HERE cosmo-workspace
*/
package dashboard

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	klog "k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/auth"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

type options struct {
	KubeConfigPath string
	KubeContext    string

	ZapOpts zap.Options

	Logr *clog.Logger
	// Client *kosmo.Client

	StaticFileDir           string
	CookieDomain            string
	CookieHashKey           string
	CookieBlockKey          string
	CookieSessionName       string
	ResponseTimeoutSeconds  int64
	GracefulShutdownSeconds int64
	TLSPrivateKeyPath       string
	TLSCertPath             string
	Insecure                bool
	ServerPort              int
	MaxAgeMinutes           int
	LdapURL                 string
	LdapStartTLS            bool
	LdapCaCertPath          string
	LdapInsecureSkipVerify  bool
	LdapBindDN              string
	LdapSearchBindDN        string
	LdapSearchBindPassword  string
	LdapSearchBaseDN        string
	LdapSearchFilter        string
}

func NewRootCmd(o *options) *cobra.Command {

	rootCmd := &cobra.Command{
		Use:   "dashboard",
		Short: "cosmo dashboard server",
		Long: `
cosmo dashboard server
Complete documentation is available at http://github.com/cosmo-workspace/cosmo

MIT 2023 cosmo-workspace/cosmo
`,
		Version:           "v0.9.1 cosmo-workspace 2023",
		PersistentPreRunE: o.PreRunE,
		RunE:              o.RunE,
	}

	goflags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	klog.InitFlags(goflags)
	o.ZapOpts.BindFlags(goflags)
	ctrl.RegisterFlags(goflags)
	rootCmd.PersistentFlags().AddGoFlagSet(goflags)

	rootCmd.PersistentFlags().Int64Var(&o.ResponseTimeoutSeconds, "timeout-seconds", 3, "Timeout seconds for response")
	rootCmd.PersistentFlags().Int64Var(&o.GracefulShutdownSeconds, "graceful-shutdown-seconds", 10, "Graceful shutdown seconds")
	rootCmd.PersistentFlags().StringVar(&o.StaticFileDir, "serve-dir", "/app/public", "Static file dir to serve")
	rootCmd.PersistentFlags().StringVar(&o.CookieDomain, "cookie-domain", "", "Cookie domain name")
	rootCmd.PersistentFlags().StringVar(&o.CookieHashKey, "cookie-hashkey", "", "Cookie hashkey")
	rootCmd.PersistentFlags().StringVar(&o.CookieBlockKey, "cookie-blockkey", "", "Cookie blockkey")
	rootCmd.PersistentFlags().StringVar(&o.CookieSessionName, "cookie-session-name", "cosmo-auth", "Cookie session name")
	rootCmd.PersistentFlags().StringVar(&o.TLSPrivateKeyPath, "tls-key", "tls.key", "TLS key file path")
	rootCmd.PersistentFlags().StringVar(&o.TLSCertPath, "tls-cert", "tls.crt", "TLS certificate file path")
	rootCmd.PersistentFlags().BoolVar(&o.Insecure, "insecure", false, "start http server not https server")
	rootCmd.PersistentFlags().IntVar(&o.ServerPort, "port", 8443, "Port for dashboard server")
	rootCmd.PersistentFlags().IntVar(&o.MaxAgeMinutes, "maxage-minutes", 720, "session maxage minutes")
	rootCmd.PersistentFlags().StringVar(&o.LdapURL, "ldap-url", "", "LDAP URL. ldap[s]://hostname.or.ip[:port]")
	rootCmd.PersistentFlags().BoolVar(&o.LdapStartTLS, "ldap-start-tls", false, "Enables StartTLS functionality")
	rootCmd.PersistentFlags().BoolVar(&o.LdapInsecureSkipVerify, "ldap-insecure-skip-verify", false, "Skip server certificate chain and hostname validation")
	rootCmd.PersistentFlags().StringVar(&o.LdapCaCertPath, "ldap-ca-cert", "", "ca cert file path")
	rootCmd.PersistentFlags().StringVar(&o.LdapBindDN, "ldap-binddn", "", "[bind mode] ex: cn=%s,ou=users,dc=example,dc=com  '%s' is replaced by the userid.")
	rootCmd.PersistentFlags().StringVar(&o.LdapSearchBindDN, "ldap-search-binddn", "", "[search mode] ex: cn=admin,dc=example,dc=com '%s' is replaced by the userid.")
	rootCmd.PersistentFlags().StringVar(&o.LdapSearchBindPassword, "ldap-search-password", "", "[search mode] password for search bindDN.")
	rootCmd.PersistentFlags().StringVar(&o.LdapSearchBaseDN, "ldap-search-basedn", "", "[search mode] ex: dc=example,dc=com")
	rootCmd.PersistentFlags().StringVar(&o.LdapSearchFilter, "ldap-search-filter", "", "[search mode] ex: (uid=%s)  '%s' is replaced by the userid.")

	return rootCmd
}

func (o *options) PreRunE(cmd *cobra.Command, args []string) error {
	cmd.MarkPersistentFlagRequired("cookie-hashkey")
	cmd.MarkPersistentFlagRequired("cookie-blockkey")

	if !o.Insecure {
		cmd.MarkPersistentFlagRequired("tls-key")
		cmd.MarkPersistentFlagRequired("tls-cert")
	}
	if o.LdapURL != "" {
		cmd.MarkPersistentFlagRequired("ldap-user-attr")
		cmd.MarkPersistentFlagRequired("ldap-basedn")
	}
	return nil
}

func (o *options) Validate(cmd *cobra.Command, args []string) error {

	if len(o.CookieHashKey) < 16 {
		return fmt.Errorf("%s is minimum 16 characters", "cookie-hashkey")
	}
	if len(o.CookieBlockKey) < 16 {
		return fmt.Errorf("%s is minimum 16 characters", "cookie-blockkey")
	}
	if o.LdapURL != "" {
		_, err := url.Parse(o.LdapURL)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *options) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *options) newLdapAuthorizer() (*auth.LdapAuthorizer, error) {
	u, _ := url.Parse(o.LdapURL)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: o.LdapInsecureSkipVerify,
		ServerName:         u.Host,
	}
	if o.LdapCaCertPath != "" {
		caCert, err := os.ReadFile(o.LdapCaCertPath)
		if err != nil {
			setupLog.Error(err, "failed to read CA cert file")
			return nil, err
		}
		certPool, err := x509.SystemCertPool()
		if err != nil {
			certPool = x509.NewCertPool()
		}
		certPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = certPool
	}
	autorizer := &auth.LdapAuthorizer{
		URL:                o.LdapURL,
		StartTLS:           o.LdapStartTLS,
		TlsConfig:          tlsConfig,
		BindDN:             o.LdapBindDN,
		SearchBindDN:       o.LdapSearchBindDN,
		SearchBindPassword: o.LdapSearchBindPassword,
		SearchBaseDN:       o.LdapSearchBaseDN,
		SearchFilter:       o.LdapSearchFilter,
	}

	return autorizer, nil
}

func (o *options) RunE(cmd *cobra.Command, args []string) error {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&o.ZapOpts)))

	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	cmd.SilenceUsage = true

	printVersion(cmd, o)
	printOptions(o)

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
		return err
	}

	// Setup server
	klient := kosmo.NewClient(mgr.GetClient())

	auths := make(map[cosmov1alpha1.UserAuthType]auth.Authorizer)
	auths[cosmov1alpha1.UserAuthTypePasswordSecert] = auth.NewPasswordSecretAuthorizer(klient)
	if o.LdapURL != "" {
		auths[cosmov1alpha1.UserAuthTypeLDAP], err = o.newLdapAuthorizer()
		if err != nil {
			return err
		}
	}

	serv := &Server{
		Log:                 clog.NewLogger(ctrl.Log.WithName("dashboard")),
		Klient:              klient,
		GracefulShutdownDur: time.Second * time.Duration(o.GracefulShutdownSeconds),
		ResponseTimeout:     time.Second * time.Duration(o.ResponseTimeoutSeconds),
		StaticFileDir:       o.StaticFileDir,
		Port:                o.ServerPort,
		MaxAgeSeconds:       60 * o.MaxAgeMinutes,
		CookieSessionName:   o.CookieSessionName,
		CookieDomain:        o.CookieDomain,
		CookieHashKey:       o.CookieHashKey,
		CookieBlockKey:      o.CookieBlockKey,
		TLSPrivateKeyPath:   o.TLSPrivateKeyPath,
		TLSCertPath:         o.TLSCertPath,
		Insecure:            o.Insecure,
		Authorizers:         auths,
		http:                &http.Server{Addr: fmt.Sprintf(":%d", o.ServerPort)},
		sessionStore:        nil,
	}

	if err := mgr.Add(serv); err != nil {
		setupLog.Error(err, "failed to add server to controller-manager")
		return err
	}

	// Start server
	setupLog.Info("Start dashboard server")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running dashboard server")
		return err
	}

	return nil
}

func Execute() {
	o := &options{}
	rootCmd := NewRootCmd(o)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(rootCmd.OutOrStdout(), err)
		os.Exit(1)
	}
}

func printOptions(o *options) {
	rv := reflect.ValueOf(*o)
	rt := rv.Type()
	options := make([]interface{}, rt.NumField()*2)

	for i := 0; i < rt.NumField(); i++ {
		options[i*2] = rt.Field(i).Name
		options[i*2+1] = rv.Field(i).Interface()
	}

	setupLog.Info("options", options...)
}

func printVersion(cmd *cobra.Command, o *options) {
	fmt.Fprintf(cmd.OutOrStdout(), "cosmo-dashboard version %s\n", cmd.Version)
}
