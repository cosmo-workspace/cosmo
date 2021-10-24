package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

type createOption struct {
	*cmdutil.UserNamespacedCliOptions

	WorkspaceName string
	Template      string
	RawVars       string
	DryRun        bool

	vars map[string]string
}

func createCmd(cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &createOption{UserNamespacedCliOptions: cliOpt}

	cmd := &cobra.Command{
		Use:               "create WORKSPACE_NAME --template TEMPLATE_NAME",
		Short:             "Create workspace",
		PersistentPreRunE: o.PreRunE,
		RunE:              o.RunE,
		Example:           "create my-code-server --user example-user --template code-server --vars PVC_SIZE_Gi:10",
	}
	cmd.Flags().StringVarP(&o.Template, "template", "t", "", "template name")
	cmd.Flags().StringVar(&o.RawVars, "vars", "", "template vars. the format is VarName:VarValue. also it can be set multiple vars by conma separated list. (example: VAR1:VAL1,VAR2:VAL2)")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", false, "dry run")

	return cmd
}

func (o *createOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *createOption) Validate(cmd *cobra.Command, args []string) error {
	if o.AllNamespace {
		return errors.New("--all-namespace is not supported in this command")
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

func (o *createOption) Complete(cmd *cobra.Command, args []string) error {
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

func (o *createOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	ws := wsv1alpha1.Workspace{}
	ws.SetName(o.WorkspaceName)
	ws.SetNamespace(wsv1alpha1.UserNamespace(o.User))
	ws.Spec = wsv1alpha1.WorkspaceSpec{
		Template: cosmov1alpha1.TemplateRef{
			Name: o.Template,
		},
		Vars: o.vars,
	}

	o.Logr.Debug().Info("creating workspace", "ws", ws, "dryrun", o.DryRun)

	if o.DryRun {
		if err := c.Create(ctx, &ws, client.DryRunAll); err != nil {
			return err
		}

		kosmo.FillTypeMeta(&ws, wsv1alpha1.GroupVersion)
		if out, err := yaml.Marshal(ws); err == nil {
			fmt.Fprintln(o.Out, string(out))
		}

		cmdutil.PrintfColorInfo(o.ErrOut, "Successfully created workspace %s (dry-run)\n", o.WorkspaceName)

	} else {
		if err := c.Create(ctx, &ws); err != nil {
			return err
		}
		cmdutil.PrintfColorInfo(o.ErrOut, "Successfully created workspace %s\n", o.WorkspaceName)
	}

	return nil
}
