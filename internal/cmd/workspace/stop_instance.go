package workspace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/utils/pointer"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

type stopInstanceOption struct {
	*cmdutil.UserNamespacedCliOptions

	InstanceName string
}

func stopInstanceCmd(cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &stopInstanceOption{UserNamespacedCliOptions: cliOpt}

	cmd := &cobra.Command{
		Use:               "stop-instance WORKSPACE_NAME",
		Aliases:           []string{"stop"},
		Short:             "Stop workspace instance",
		PersistentPreRunE: o.PreRunE,
		RunE:              cmdutil.RunEHandler(o.RunE),
	}
	return cmd
}

func (o *stopInstanceOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *stopInstanceOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *stopInstanceOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.InstanceName = args[0]
	return nil
}

func (o *stopInstanceOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	if _, err := c.UpdateWorkspace(ctx, o.InstanceName, o.User, kosmo.UpdateWorkspaceOpts{Replicas: pointer.Int64(0)}); err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully stopped workspace %s\n", o.InstanceName)
	return nil
}
