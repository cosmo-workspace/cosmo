package template

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	cmdutil "github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

type generateOption struct {
	*cmdutil.CliOptions
	wsConfig wsv1alpha1.Config

	Name         string
	OutputFile   string
	RequiredVars string
	Desc         string

	TypeWorkspace                bool
	DisableInjectAuthProxy       bool
	InjectAuthProxyImage         string
	InjectAuthProxyTLSSecretName string
	ServiceAccount               string

	TypeUserAddon       bool
	SetDefaultUserAddon bool
	SetSysnsUserAddon   string

	tmpl cosmov1alpha1.Template
}

func generateCmd(cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &generateOption{CliOptions: cliOpt}
	cmd := &cobra.Command{
		Use:     "generate --name TEMPLATE_NAME [< Input via Stdin or pipe]",
		Aliases: []string{"gen"},
		Short:   "Generate Template",
		Long: `Generate Template

For create generated template, just do "kubectl create -f cosmo-template.yaml"

Example:
  * Pipe from kustomize build and apply to your cluster in a single line 
	
      kustomize build ./kubernetes/ | cosmoctl template generate --name TEMPLATE_NAME | kubectl apply -f -

  * Pipe from helm template and generate Workspace Template with cosmo-auth-proxy injection
	
  	  helm template code-server ci/helm-chart \
		| cosmoctl template generate --name TEMPLATE_NAME --workspace \
			--workspace-urlbase 'https://{{NETRULE_GROUP}}-{{INSTANCE}}-{{NAMESPACE}}.yourdomain:443'

  * Input merged config file (kustomize build ... or helm template ... etc.) and save it to file

      cosmoctl template generate --name TEMPLATE_NAME -o cosmo-template.yaml < merged.yaml
`,
		PersistentPreRunE: o.PreRunE,
		RunE:              o.RunE,
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "template name (use directory name if not specified)")
	cmd.Flags().StringVarP(&o.OutputFile, "output", "o", "", "write output into file (default: Stdout)")
	cmd.Flags().StringVar(&o.RequiredVars, "required-vars", "", "template custom vars to be replaced by instance. format --required-vars VAR1,VAR2:default-value")
	cmd.Flags().StringVar(&o.Desc, "desc", "", "template description")

	cmd.Flags().BoolVar(&o.TypeWorkspace, "workspace", false, "template as type workspace")
	cmd.Flags().BoolVar(&o.DisableInjectAuthProxy, "disable-inject-auth-proxy", false, "disable injection cosmo-auth-proxy sidecar")
	cmd.Flags().StringVar(&o.InjectAuthProxyImage, "inject-auth-proxy-image", "ghcr.io/cosmo-workspace/cosmo-auth-proxy:latest", "cosmo-auth-proxy sidecar image. use with --workspace")
	cmd.Flags().StringVar(&o.InjectAuthProxyTLSSecretName, "inject-auth-proxy-tls-secret", "", "TLS secret name for https sidecar cosmo-auth-proxy. Be empty if http. use with --workspace")
	cmd.Flags().StringVar(&o.ServiceAccount, "serviceaccount", "default", "service account name for cosmo-auth-proxy rolebinding")

	cmd.Flags().StringVar(&o.wsConfig.DeploymentName, "workspace-deployment-name", "", "Deployment name for Workspace. use with --workspace")
	cmd.Flags().StringVar(&o.wsConfig.ServiceName, "workspace-service-name", "", "Service name for Workspace. use with --workspace")
	cmd.Flags().StringVar(&o.wsConfig.IngressName, "workspace-ingress-name", "", "Ingress name for Workspace. use with --workspace")
	cmd.Flags().StringVar(&o.wsConfig.ServiceMainPortName, "workspace-main-service-port-name", "", "ServicePort name for Workspace main container port. use with --workspace")
	cmd.Flags().StringVar(&o.wsConfig.URLBase, "workspace-urlbase", "", "Workspace URLBase. use with --workspace")

	cmd.Flags().BoolVar(&o.TypeUserAddon, "user-addon", false, "template as type user-addon")
	cmd.Flags().BoolVar(&o.SetDefaultUserAddon, "set-default-user-addon", false, "set default user addon")
	cmd.Flags().StringVar(&o.SetSysnsUserAddon, "set-sysns-user-addon", "", "user addon in system namespace")

	return cmd
}

func (o *generateOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *generateOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Validate(cmd, args); err != nil {
		return err
	}

	if o.TypeWorkspace {
		if o.wsConfig.URLBase == "" {
			return errors.New("--workspace-urlbase is required")
		}
	}

	if o.TypeWorkspace && o.TypeUserAddon {
		return errors.New("--workspace and --user-addon is incompatible")
	}

	return nil
}

func (o *generateOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}

	if o.Name == "" {
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		o.Name = filepath.Base(dir)
	}

	if o.OutputFile != "" {
		var err error
		o.OutputFile, err = filepath.Abs(o.OutputFile)
		if err != nil {
			return err
		}
	}

	if o.RequiredVars != "" {
		varsList := strings.Split(o.RequiredVars, ",")

		vars := make([]cosmov1alpha1.RequiredVarSpec, 0, len(varsList))
		for _, v := range varsList {
			vcol := strings.Split(v, ":")
			varSpec := cosmov1alpha1.RequiredVarSpec{Var: vcol[0]}
			if len(vcol) > 1 {
				varSpec.Default = vcol[1]
			}
			vars = append(vars, varSpec)
		}
		o.tmpl.Spec.RequiredVars = vars
	}

	o.tmpl.Name = o.Name
	o.tmpl.Spec.Description = o.Desc

	gvk, err := apiutil.GVKForObject(&o.tmpl, o.Scheme)
	if err != nil {
		return err
	}
	o.tmpl.SetGroupVersionKind(gvk)

	if o.TypeWorkspace {
		template.SetTemplateType(&o.tmpl, wsv1alpha1.TemplateTypeWorkspace)
	} else if o.TypeUserAddon {
		template.SetTemplateType(&o.tmpl, wsv1alpha1.TemplateTypeUserAddon)

		ann := o.tmpl.GetAnnotations()
		if ann == nil {
			ann = make(map[string]string)
		}
		if o.SetDefaultUserAddon {
			ann[wsv1alpha1.TemplateAnnKeyDefaultUserAddon] = strconv.FormatBool(true)
		}
		if o.SetSysnsUserAddon != "" {
			ann[wsv1alpha1.TemplateAnnKeySysNsUserAddon] = o.SetSysnsUserAddon
		}
		o.tmpl.SetAnnotations(ann)
	}

	return nil
}

func (o *generateOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if isatty.IsTerminal(os.Stdin.Fd()) {
		return fmt.Errorf("no input via stdin")
	}

	// input data from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read input file : %w", err)
	}
	if len(input) == 0 {
		return fmt.Errorf("no input")
	}
	o.Logr.DebugAll().Info(string(input), "obj", "input k8s configs")

	// create tmp dir
	tmpDir, err := ioutil.TempDir(os.TempDir(), "cosmoctl-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir : %w", err)
	}
	defer os.RemoveAll(tmpDir)
	o.Logr.Debug().Info("tmpDir created", "path", tmpDir)

	// save it as "packaged" file
	if err := cmdutil.CreateFile(tmpDir, DefaultPackagedFile, input); err != nil {
		return err
	}
	o.Logr.Debug().Info(fmt.Sprintf("%s created", DefaultPackagedFile))

	// if type workspace, validate and set label
	o.Logr.Debug().Info("template type", "workspace", o.TypeWorkspace)
	unsts, err := preTemplateBuild(string(input))
	if err != nil {
		return fmt.Errorf("failed to pre-build template: %w", err)
	}

	if o.TypeWorkspace {
		if err := completeWorkspaceConfig(&o.wsConfig, unsts); err != nil {
			return fmt.Errorf("type workspace validation failed: %w", err)
		}
		wsv1alpha1.SetConfigOnTemplateAnnotations(&o.tmpl, o.wsConfig)
	}

	kust := NewKustomize()

	// inject cosmo-auth-proxy if enabled
	o.Logr.Debug().Info("inject cosmo-auth-proxy", "enabled", o.TypeWorkspace, "image", o.InjectAuthProxyImage)

	if o.TypeWorkspace && !o.DisableInjectAuthProxy {
		// patch deployment
		deploy := deploymentAuthProxyPatch(o.wsConfig.DeploymentName, o.InjectAuthProxyImage, o.InjectAuthProxyTLSSecretName)
		rawDeploy := StructToYaml(deploy)
		err := cmdutil.CreateFile(tmpDir, AuthProxyPatchFile, rawDeploy)
		if err != nil {
			return err
		}
		o.Logr.DebugAll().Info(string(rawDeploy), "obj", "cosmo-auth-proxy deployment patch", "file", AuthProxyPatchFile)

		addPatchesStrategicMerges(kust, AuthProxyPatchFile)

		// add auth proxy rolebindings
		if o.ServiceAccount != "default" {
			roleb := wsv1alpha1.AuthProxyRoleBindingApplyConfiguration(o.ServiceAccount, template.DefaultVarsNamespace)
			rawRoleb := StructToYaml(roleb)
			if err := cmdutil.CreateFile(tmpDir, AuthProxyRoleBFile, rawRoleb); err != nil {
				return err
			}
			o.Logr.DebugAll().Info(string(rawRoleb), "obj", "cosmo-auth-proxy rolebinding", "file", AuthProxyRoleBFile)

			kust.Resources = append(kust.Resources, AuthProxyRoleBFile)
		}
	}

	// run kustomize
	out, err := cmdutil.ExecKustomize(ctx, tmpDir, kust)
	if err != nil {
		return err
	}
	o.Logr.DebugAll().Info(string(out), "obj", "updated k8s configs")

	o.tmpl.Spec.RawYaml = string(out)

	outtmpl, _ := yaml.Marshal(&o.tmpl)

	output := append([]byte("# Generated by cosmoctl template command\n"), outtmpl...)

	// output to Stdout or write the output to file given by option
	if o.OutputFile == "" {
		fmt.Fprintln(o.Out, string(output))
	} else {
		if err := cmdutil.CreateFile(filepath.Dir(o.OutputFile), filepath.Base(o.OutputFile), output); err != nil {
			return err
		}
	}
	return nil
}

func preTemplateBuild(rawTmpl string) ([]unstructured.Unstructured, error) {
	var inst cosmov1alpha1.Instance
	inst.SetName("dummy")
	inst.SetNamespace("dummy")

	builder := template.NewUnstructuredBuilder(rawTmpl, &inst)
	return builder.Build()
}
