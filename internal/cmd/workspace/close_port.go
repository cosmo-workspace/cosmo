package workspace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type closePortOption struct {
	*cmdutil.UserNamespacedCliOptions

	WorkspaceName string
	PortName      string
}

func closePortCmd(cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &closePortOption{UserNamespacedCliOptions: cliOpt}

	cmd := &cobra.Command{
		Use:               "close-port WORKSPACE_NAME --port-name PORT_NAME",
		Short:             "Remove workspace network port",
		PersistentPreRunE: o.PreRunE,
		RunE:              cmdutil.RunEHandler(o.RunE),
	}
	cmd.Flags().StringVar(&o.PortName, "port-name", "", "port name (Required)")
	return cmd
}

func (o *closePortOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *closePortOption) Validate(cmd *cobra.Command, args []string) error {
	if o.AllNamespace {
		return errors.New("--all-namespaces is not supported in this command")
	}
	if err := o.UserNamespacedCliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	if o.PortName == "" {
		return errors.New("port name is required")
	}
	return nil
}

func (o *closePortOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.WorkspaceName = args[0]
	return nil
}

func (o *closePortOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	if _, err := c.DeleteNetworkRule(ctx, o.WorkspaceName, o.User, o.PortName); err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully closed port '%s' for workspace '%s'\n", o.PortName, o.WorkspaceName)
	return nil
}
