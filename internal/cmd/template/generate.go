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

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/internal/cmd/version"
	cmdutil "github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	"github.com/cosmo-workspace/cosmo/pkg/workspace"
)

type generateOption struct {
	*cmdutil.CliOptions
	wsConfig cosmov1alpha1.Config

	Name         string
	OutputFile   string
	RequiredVars string
	Desc         string

	TypeWorkspace bool
	TypeUserAddon bool

	SetDefaultUserAddon bool
	DisableNamePrefix   bool
	ClusterScope        bool
	UserRoles           string
	ForbiddenUserRoles  string

	tmpl cosmov1alpha1.TemplateObject
}

func generateCmd(cmd *cobra.Command, cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &generateOption{CliOptions: cliOpt}
	cmd.PersistentPreRunE = o.PreRunE
	cmd.RunE = cmdutil.RunEHandler(o.RunE)

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "template name (use directory name if not specified)")
	cmd.Flags().StringVarP(&o.OutputFile, "output", "o", "", "write output into file (default: Stdout)")
	cmd.Flags().StringVar(&o.RequiredVars, "required-vars", "", "template custom vars to be replaced by instance. format --required-vars VAR1,VAR2:default-value")
	cmd.Flags().StringVar(&o.Desc, "desc", "", "template description")

	cmd.Flags().BoolVar(&o.TypeWorkspace, "workspace", false, "template as type workspace")
	cmd.Flags().StringVar(&o.wsConfig.DeploymentName, "workspace-deployment-name", "", "Deployment name for Workspace. use with --workspace (auto detected if not specified)")
	cmd.Flags().StringVar(&o.wsConfig.ServiceName, "workspace-service-name", "", "Service name for Workspace. use with --workspace (auto detected if not specified)")
	cmd.Flags().StringVar(&o.wsConfig.ServiceMainPortName, "workspace-main-service-port-name", "", "ServicePort name for Workspace main container port. use with --workspace (auto detected if not specified)")

	cmd.Flags().BoolVar(&o.TypeUserAddon, "user-addon", false, "template as type useraddon")
	cmd.Flags().BoolVar(&o.TypeUserAddon, "useraddon", false, "template as type useraddon")
	cmd.Flags().BoolVar(&o.SetDefaultUserAddon, "set-default-user-addon", false, "set default user addon")
	cmd.Flags().BoolVar(&o.DisableNamePrefix, "disable-nameprefix", false, "disable adding instance name prefix on child resource name")

	cmd.Flags().BoolVar(&o.ClusterScope, "cluster-scope", false, "generate ClusterTemplate (default generate namespaced Template)")
	cmd.Flags().StringVar(&o.UserRoles, "userroles", "", "user roles to show this template (e.g. 'teama-*', 'teamb-admin', etc.)")
	cmd.Flags().StringVar(&o.ForbiddenUserRoles, "forbidden-userroles", "", "user roles NOT to show this template (e.g. 'teama-*', 'teamb-admin', etc.)")

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

	if o.TypeWorkspace && o.TypeUserAddon {
		return errors.New("--workspace and --user-addon cannot be specified concurrently")
	}

	if o.TypeWorkspace && o.ClusterScope {
		return errors.New("workspace template cannot be cluster-scoped")
	}

	return nil
}

func (o *generateOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}

	if o.ClusterScope {
		o.tmpl = &cosmov1alpha1.ClusterTemplate{}
	} else {
		o.tmpl = &cosmov1alpha1.Template{}
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
		o.tmpl.GetSpec().RequiredVars = vars
	}

	o.tmpl.SetName(o.Name)
	o.tmpl.GetSpec().Description = o.Desc

	gvk, err := apiutil.GVKForObject(o.tmpl, o.Scheme)
	if err != nil {
		return err
	}
	o.tmpl.SetGroupVersionKind(gvk)

	// annotations
	ann := o.tmpl.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string)
	}

	if o.TypeWorkspace {
		template.SetTemplateType(o.tmpl, cosmov1alpha1.TemplateLabelEnumTypeWorkspace)
	} else if o.TypeUserAddon {
		template.SetTemplateType(o.tmpl, cosmov1alpha1.TemplateLabelEnumTypeUserAddon)

		if o.SetDefaultUserAddon {
			ann[cosmov1alpha1.UserAddonTemplateAnnKeyDefaultUserAddon] = strconv.FormatBool(true)
		}
		if o.DisableNamePrefix {
			ann[cosmov1alpha1.TemplateAnnKeyDisableNamePrefix] = strconv.FormatBool(true)
		}
	}

	if o.UserRoles != "" {
		ann[cosmov1alpha1.TemplateAnnKeyUserRoles] = o.UserRoles
	}
	if o.ForbiddenUserRoles != "" {
		ann[cosmov1alpha1.TemplateAnnKeyForbiddenUserRoles] = o.ForbiddenUserRoles
	}

	o.tmpl.SetAnnotations(ann)

	return nil
}

func (o *generateOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()

	if isatty.IsTerminal(os.Stdin.Fd()) {
		return fmt.Errorf("no input via stdin")
	}

	// input data from stdin
	input, err := io.ReadAll(o.In)
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
		workspace.SetConfigOnTemplateAnnotations(o.tmpl, o.wsConfig)
	}

	kust := NewKustomize(o.DisableNamePrefix)

	// run kustomize
	out, err := cmdutil.ExecKustomize(ctx, tmpDir, kust)
	if err != nil {
		return err
	}
	o.Logr.Debug().Info(string(out), "obj", "updated k8s configs")

	o.tmpl.GetSpec().RawYaml = string(out)

	outtmpl, _ := yaml.Marshal(&o.tmpl)

	output := append([]byte("# Generated by "+version.Footprint+"\n"), outtmpl...)

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

	builder := template.NewRawYAMLBuilder(rawTmpl, &inst)
	return builder.Build()
}
