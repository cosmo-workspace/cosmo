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

type DeleteOption struct {
	*cmdutil.UserNamespacedCliOptions

	WorkspaceName string
	DryRun        bool
}

func DeleteCmd(cmd *cobra.Command, cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &DeleteOption{UserNamespacedCliOptions: cliOpt}

	cmd.PersistentPreRunE = o.PreRunE
	cmd.RunE = cmdutil.RunEHandler(o.RunE)
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", false, "dry run")
	return cmd
}

func (o *DeleteOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *DeleteOption) Validate(cmd *cobra.Command, args []string) error {
	if o.AllNamespace {
		return errors.New("--all-namespaces is not supported in this command")
	}
	if err := o.UserNamespacedCliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *DeleteOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.WorkspaceName = args[0]
	return nil
}

func (o *DeleteOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()

	if o.DryRun {
		if _, err := o.Client.DeleteWorkspace(ctx, o.WorkspaceName, o.User, client.DryRunAll); err != nil {
			return err
		}
		cmdutil.PrintfColorInfo(o.ErrOut, "Successfully deleted workspace %s (dry-run)\n", o.WorkspaceName)

	} else {
		if _, err := o.Client.DeleteWorkspace(ctx, o.WorkspaceName, o.User); err != nil {
			return err
		}
		cmdutil.PrintfColorInfo(o.ErrOut, "Successfully deleted workspace %s\n", o.WorkspaceName)
	}

	return nil
}
