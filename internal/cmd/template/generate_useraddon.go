package template

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
)

type generateUserAddonOption struct {
	*cli.RootOptions

	Name               string
	OutputFile         string
	RequiredVars       []string
	Desc               string
	NoHeader           bool
	UserRoles          []string
	RequiredUserAddons []string
	SetDefault         bool
	SetUserNamePrefix  bool
	ClusterScope       bool
}

func generateUserAddonCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &generateUserAddonOption{RootOptions: cliOpt}
	cmd.RunE = cli.ConnectErrorHandler(o)

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "template name (use directory name if not specified)")
	cmd.Flags().StringVarP(&o.OutputFile, "output", "o", "", "write output into file (default: Stdout)")
	cmd.Flags().StringSliceVar(&o.RequiredVars, "var", []string{}, "template custom vars. format --var=VAR1 --var=VAR2:default-value")
	cmd.Flags().StringVar(&o.Desc, "desc", "", "template description")
	cmd.Flags().BoolVar(&o.NoHeader, "no-header", false, "no output headers")
	cmd.Flags().StringSliceVar(&o.UserRoles, "userroles", []string{}, "user roles only to show this template (e.g. 'teama-*', 'teamb-admin', etc.)")
	cmd.Flags().StringSliceVar(&o.RequiredUserAddons, "required-useraddons", []string{}, "add dependency to use this useraddon")

	cmd.Flags().BoolVar(&o.SetDefault, "default", false, "set default. default user addon is applied to all users")
	cmd.Flags().BoolVar(&o.SetUserNamePrefix, "user-prefix", false, "adding user name prefix on child resource name. default false but true if --cluster-scope is specified")
	cmd.Flags().BoolVarP(&o.ClusterScope, "cluster-scope", "c", false, "include cluster-scoped resoure like ClusterRoleBindings, PersistentVolume etc.")
	return cmd
}

func (o *generateUserAddonOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	return nil
}

func (o *generateUserAddonOption) Complete(cmd *cobra.Command, args []string) error {
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

	if o.ClusterScope {
		o.SetUserNamePrefix = true
	}
	return nil
}

func (o *generateUserAddonOption) RunE(cmd *cobra.Command, args []string) error {
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

	builder := NewTemplateObjectBuilder(o.ClusterScope).
		Name(o.Name).
		Description(o.Desc).
		RequiredVars(o.RequiredVars).
		SetRequiredAddons(o.RequiredUserAddons).
		SetUserRoles(o.UserRoles).
		TypeUserAddon(o.SetDefault).
		Resources(input)

	if !o.SetUserNamePrefix {
		builder = builder.DisableNamePrefix()
	}

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
