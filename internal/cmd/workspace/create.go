package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type CreateOption struct {
	*cmdutil.UserNamespacedCliOptions

	WorkspaceName string
	Template      string
	RawVars       string
	DryRun        bool

	vars map[string]string
}

func CreateCmd(cmd *cobra.Command, cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &CreateOption{UserNamespacedCliOptions: cliOpt}

	cmd.PersistentPreRunE = o.PreRunE
	cmd.RunE = cmdutil.RunEHandler(o.RunE)
	cmd.Flags().StringVarP(&o.Template, "template", "t", "", "template name")
	cmd.Flags().StringVar(&o.RawVars, "vars", "", "template vars. the format is VarName:VarValue. also it can be set multiple vars by conma separated list. (example: VAR1:VAL1,VAR2:VAL2)")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", false, "dry run")

	return cmd
}

func (o *CreateOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *CreateOption) Validate(cmd *cobra.Command, args []string) error {
	if o.AllNamespace {
		return errors.New("--all-namespaces is not supported in this command")
	}
	if err := o.UserNamespacedCliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	if o.Template == "" {
		return errors.New("--template is required")
	}
	return nil
}

func (o *CreateOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.WorkspaceName = args[0]

	if o.RawVars != "" {
		vars := make(map[string]string)
		varAndVals := strings.Split(o.RawVars, ",")
		for _, v := range varAndVals {
			varAndVal := strings.Split(v, ":")
			if len(varAndVal) != 2 {
				return fmt.Errorf("vars format error: vars %s must be 'VAR:VAL'", v)
			}
			vars[varAndVal[0]] = varAndVal[1]
		}
		o.vars = vars
	}

	return nil
}

func (o *CreateOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	if o.DryRun {
		ws, err := c.CreateWorkspace(ctx, o.User, o.WorkspaceName, o.Template, o.vars, client.DryRunAll)
		if err != nil {
			return err
		}

		gvk, err := apiutil.GVKForObject(ws, o.Scheme)
		if err != nil {
			return err
		}
		ws.SetGroupVersionKind(gvk)
		if out, err := yaml.Marshal(ws); err == nil {
			fmt.Fprintln(o.Out, string(out))
		}

		cmdutil.PrintfColorInfo(o.ErrOut, "Successfully created workspace %s (dry-run)\n", o.WorkspaceName)

	} else {
		if _, err := c.CreateWorkspace(ctx, o.User, o.WorkspaceName, o.Template, o.vars); err != nil {
			return err
		}

		cmdutil.PrintfColorInfo(o.ErrOut, "Successfully created workspace %s\n", o.WorkspaceName)
	}

	return nil
}
