package workspace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type deleteOption struct {
	*cmdutil.UserNamespacedCliOptions

	WorkspaceName string
	DryRun        bool
}

func deleteCmd(cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &deleteOption{UserNamespacedCliOptions: cliOpt}

	cmd := &cobra.Command{
		Use:               "delete WORKSPACE_NAME",
		Aliases:           []string{"del"},
		Short:             "Delete workspace",
		PersistentPreRunE: o.PreRunE,
		RunE:              o.RunE,
	}
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", false, "dry run")
	return cmd
}

func (o *deleteOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *deleteOption) Validate(cmd *cobra.Command, args []string) error {
	if o.AllNamespace {
		return errors.New("--all-namespace is not supported in this command")
	}
	if err := o.UserNamespacedCliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *deleteOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.WorkspaceName = args[0]
	return nil
}

func (o *deleteOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()

	inst, err := o.Client.GetInstance(ctx, o.WorkspaceName, o.Namespace)
	if err != nil {
		return err
	}

	o.Logr.Debug().Info("deleting workspace", "inst", inst, "dryrun", o.DryRun)
	if o.DryRun {
		if err := o.Client.Delete(ctx, inst, client.DryRunAll); err != nil {
			return err
		}
		cmdutil.PrintfColorInfo(o.ErrOut, "Successfully deleted workspace %s (dry-run)\n", o.WorkspaceName)

	} else {
		if err := o.Client.Delete(ctx, inst); err != nil {
			return err
		}
		cmdutil.PrintfColorInfo(o.ErrOut, "Successfully deleted workspace %s\n", o.WorkspaceName)
	}

	return nil
}
