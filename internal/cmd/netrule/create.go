package netrule

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type CreateOption struct {
	*cmdutil.UserNamespacedCliOptions

	WorkspaceName    string
	CustomHostPrefix string
	PortNumber       int32
	HTTPPath         string
	Public           bool

	rule cosmov1alpha1.NetworkRule
}

func CreateCmd(cmd *cobra.Command, cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &CreateOption{UserNamespacedCliOptions: cliOpt}

	cmd.PersistentPreRunE = o.PreRunE
	cmd.RunE = cmdutil.RunEHandler(o.RunE)
	cmd.Flags().StringVar(&o.WorkspaceName, "workspace", "", "workspace name (Required)")
	cmd.Flags().Int32Var(&o.PortNumber, "port", 0, "serivce port number (Required)")
	cmd.Flags().StringVar(&o.CustomHostPrefix, "host-prefix", "", "custom host prefix")
	cmd.Flags().StringVar(&o.HTTPPath, "path", "/", "path for Ingress path when using ingress")
	cmd.Flags().BoolVar(&o.Public, "public", false, "disable authentication for this port")

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
	if o.WorkspaceName == "" {
		return errors.New("--workspace is required")
	}
	if o.PortNumber == 0 {
		return errors.New("--port is required")
	}
	return nil
}

func (o *CreateOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Complete(cmd, args); err != nil {
		return err
	}

	o.rule = cosmov1alpha1.NetworkRule{
		CustomHostPrefix: o.CustomHostPrefix,
		PortNumber:       o.PortNumber,
		HTTPPath:         o.HTTPPath,
		Public:           o.Public,
	}
	o.rule.Default()
	return nil
}

func (o *CreateOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	ws, err := c.GetWorkspaceByUserName(ctx, o.WorkspaceName, o.User)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %v", err)
	}
	index := -1
	for i, v := range ws.Spec.Network {
		if v.UniqueKey() == o.rule.UniqueKey() {
			index = i
		}
	}

	if _, err := c.AddNetworkRule(ctx, o.WorkspaceName, o.User, o.rule, index); err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully add network rule for workspace '%s'\n", o.WorkspaceName)
	return nil
}
