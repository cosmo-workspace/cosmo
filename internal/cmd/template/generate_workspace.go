package template

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

type generateWorkspaceOption struct {
	*cli.RootOptions
	Name               string
	OutputFile         string
	RequiredVars       []string
	Desc               string
	NoHeader           bool
	UserRoles          []string
	RequiredUserAddons []string
	wsConfig           cosmov1alpha1.Config
}

func generateWorkspaceCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &generateWorkspaceOption{RootOptions: cliOpt}
	cmd.RunE = cli.ConnectErrorHandler(o)

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "template name (use directory name if not specified)")
	cmd.Flags().StringVarP(&o.OutputFile, "output", "o", "", "write output into file (default: Stdout)")
	cmd.Flags().StringSliceVar(&o.RequiredVars, "var", []string{}, "indicate template custom vars. format --var=VAR1 --var=VAR2:default-value")
	cmd.Flags().StringVar(&o.Desc, "desc", "", "template description")
	cmd.Flags().BoolVar(&o.NoHeader, "no-header", false, "no output headers")
	cmd.Flags().StringSliceVar(&o.UserRoles, "userroles", []string{}, "user roles only to show this template (e.g. 'teama-*', 'teamb-admin', etc.)")
	cmd.Flags().StringSliceVar(&o.RequiredUserAddons, "required-useraddons", []string{}, "add dependency to use this useraddon")

	cmd.Flags().StringVar(&o.wsConfig.DeploymentName, "deployment", "", "Deployment name for Workspace (auto detected if not specified)")
	cmd.Flags().StringVar(&o.wsConfig.ServiceName, "service", "", "Service name for Workspace (auto detected if not specified)")
	cmd.Flags().StringVar(&o.wsConfig.ServiceMainPortName, "main-service-port", "", "ServicePort name for Workspace main container port (auto detected if not specified)")

	return cmd
}

func (o *generateWorkspaceOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	return nil
}

func (o *generateWorkspaceOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.CompleteWithoutClient(cmd, args); err != nil {
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
		outFile, err := filepath.Abs(o.OutputFile)
		if err != nil {
			return err
		}
		o.OutputFile = outFile
	}
	return nil
}

func (o *generateWorkspaceOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	input, err := cli.ReadFromPipedStdin()
	if err != nil {
		return err
	}
	o.Logr.Debug().Info(input)

	unsts, err := template.NewRawYAMLBuilder(input, nil).Build()
	if err != nil {
		return fmt.Errorf("failed to build template: %w", err)
	}
	if err := completeWorkspaceConfig(&o.wsConfig, unsts); err != nil {
		return err
	}

	builder := NewTemplateObjectBuilder(false).
		Name(o.Name).
		Description(o.Desc).
		RequiredVars(o.RequiredVars).
		SetRequiredAddons(o.RequiredUserAddons).
		SetUserRoles(o.UserRoles).
		TypeWorkspace(o.wsConfig).
		Resources(input)

	if !o.NoHeader {
		builder.SetHeader(o.Versions)
	}

	out, err := builder.Build(o.Ctx)
	if err != nil {
		return err
	}

	// output to Stdout or write the output to file given by option
	if o.OutputFile == "" {
		fmt.Fprintln(cmd.OutOrStdout(), string(out))
	} else {
		if err := os.WriteFile(o.OutputFile, out, 0644); err != nil {
			return fmt.Errorf("failed to write file %s : %w", o.OutputFile, err)
		}
	}
	return nil
}

func completeWorkspaceConfig(wsConfig *cosmov1alpha1.Config, unst []unstructured.Unstructured) error {
	if wsConfig == nil || len(unst) == 0 {
		return errors.New("invalid args")
	}

	dps := make([]unstructured.Unstructured, 0)
	svcs := make([]unstructured.Unstructured, 0)

	for _, u := range unst {
		if kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.DeploymentGVK) {
			dps = append(dps, u)
		} else if kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.ServiceGVK) {
			svcs = append(svcs, u)
		}
	}

	// complete deployment name
	if wsConfig.DeploymentName == "" {
		if len(dps) != 1 {
			return errors.New("no deployment")
		}
		wsConfig.DeploymentName = dps[0].GetName()
	}

	// validate deployment
	var validDep, validSvc bool
	for _, v := range dps {
		if wsConfig.DeploymentName == v.GetName() {
			validDep = true
		}
	}
	if !validDep {
		return fmt.Errorf("deployment '%s' is not found", wsConfig.DeploymentName)
	}

	// complete service name
	if wsConfig.ServiceName == "" {
		if len(svcs) != 1 {
			return errors.New("no service")
		}
		wsConfig.ServiceName = svcs[0].GetName()
	}

	// validate service
	var svc corev1.Service
	for _, v := range svcs {
		if wsConfig.ServiceName == v.GetName() {
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(v.Object, &svc)
			if err != nil {
				return err
			}
			validSvc = true
		}
	}
	if !validSvc {
		return fmt.Errorf("service '%s' is not found", wsConfig.ServiceName)
	}

	// complete service main port
	if wsConfig.ServiceMainPortName == "" {
		if len(svc.Spec.Ports) != 1 {
			return errors.New("failed to specify the service port")
		}
		wsConfig.ServiceMainPortName = svc.Spec.Ports[0].Name
	}

	// validate service main port
	var mainServicePort int32
	for _, port := range svc.Spec.Ports {
		if port.Name == wsConfig.ServiceMainPortName {
			mainServicePort = port.Port
		}
	}
	if mainServicePort == 0 {
		return fmt.Errorf("service '%s' is not found", wsConfig.ServiceName)
	}

	return nil
}
