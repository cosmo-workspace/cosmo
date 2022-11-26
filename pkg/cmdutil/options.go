package cmdutil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

type CliOptions struct {
	KubeConfigPath string
	KubeContext    string
	LogLevel       int
	In             io.Reader
	Out            io.Writer
	ErrOut         io.Writer

	Ctx    context.Context
	Logr   *clog.Logger
	Client *kosmo.Client
	Scheme *runtime.Scheme
}

type NamespacedCliOptions struct {
	*CliOptions
	Namespace    string
	AllNamespace bool
}

type UserNamespacedCliOptions struct {
	*NamespacedCliOptions
	User string
}

func NewCliOptions() *CliOptions {
	ctx := context.TODO()
	return &CliOptions{Ctx: ctx}
}

func NewNamespacedCliOptions(o *CliOptions) *NamespacedCliOptions {
	return &NamespacedCliOptions{CliOptions: o}
}

func NewUserNamespacedCliOptions(o *CliOptions) *UserNamespacedCliOptions {
	return &UserNamespacedCliOptions{NamespacedCliOptions: NewNamespacedCliOptions(o)}
}

func (o *CliOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *CliOptions) Complete(cmd *cobra.Command, args []string) error {
	if o.LogLevel >= 0 {
		opt := zap.Options{
			Development: true,
			Level:       zapcore.Level(-o.LogLevel),
		}
		o.Logr = clog.NewLogger(zap.New(zap.UseFlagOptions(&opt)))
		o.Ctx = clog.IntoContext(o.Ctx, o.Logr)
	} else {
		o.Logr = clog.NewLogger(logr.Discard())
	}
	debug := o.Logr.WithCaller().DebugAll()

	if o.Client == nil {
		cfgFlg := genericclioptions.NewConfigFlags(true)
		debug.Info("kubeconfigs", "kubeConfigPath", o.KubeConfigPath, "kubeContext", o.KubeContext)

		if o.KubeConfigPath != "" {
			cfgFlg.KubeConfig = &o.KubeConfigPath
		}
		if o.KubeContext != "" {
			cfgFlg.Context = &o.KubeContext
		}

		cfg, err := cfgFlg.ToRESTConfig()
		if err != nil {
			return err
		}
		debug.Info("RestConfig", "cfg", cfg)

		baseclient, err := kosmo.NewClientByRestConfig(cfg, scheme)
		if err != nil {
			return err
		}
		o.Client = &baseclient
		o.Scheme = scheme
	}

	return nil
}

func (o *NamespacedCliOptions) Validate(cmd *cobra.Command, args []string) error {
	if o.AllNamespace && o.Namespace != "" {
		return errors.New("--all-namespaces connot be used with --namespace")
	}
	return o.CliOptions.Validate(cmd, args)
}

func (o *NamespacedCliOptions) Complete(cmd *cobra.Command, args []string) error {
	if !o.AllNamespace && o.Namespace == "" {
		cfg, err := GetKubeConfig(o.KubeConfigPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		o.Namespace = GetDefaultNamespace(cfg, o.KubeContext)
		if o.Namespace == "" {
			return errors.New("failed to get default namespace")
		}
	}
	return o.CliOptions.Complete(cmd, args)
}

func (o *UserNamespacedCliOptions) Validate(cmd *cobra.Command, args []string) error {
	if o.User != "" && o.Namespace != "" {
		return errors.New("--user and --namespace connot be used at the same time")
	}
	if o.AllNamespace && (o.Namespace != "" || o.User != "") {
		return errors.New("--all-namespaces connot be used with --namespace or --user")
	}
	return o.NamespacedCliOptions.Validate(cmd, args)
}

func (o *UserNamespacedCliOptions) Complete(cmd *cobra.Command, args []string) error {
	if !o.AllNamespace {
		if o.Namespace == "" && o.User != "" {
			o.Namespace = cosmov1alpha1.UserNamespace(o.User)
		}
	}
	if err := o.NamespacedCliOptions.Complete(cmd, args); err != nil {
		return err
	}
	if !o.AllNamespace {
		if o.Namespace != "" && o.User == "" {
			userName := cosmov1alpha1.UserNameByNamespace(o.Namespace)
			if userName == "" {
				return fmt.Errorf("namespace %s is not cosmo user's namespace", o.Namespace)
			}
			o.User = userName
		}
	}
	return nil
}
