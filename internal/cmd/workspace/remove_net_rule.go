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

type RemoveNetRuleOption struct {
	*cmdutil.UserNamespacedCliOptions

	WorkspaceName string
	NetRuleName   string
}

func removeNetRuleCmd(cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &RemoveNetRuleOption{UserNamespacedCliOptions: cliOpt}

	cmd := &cobra.Command{
		Use:               "remove-net-rule WORKSPACE_NAME --name NETWORK_RULE_NAME",
		Short:             "Remove workspace network rule",
		PersistentPreRunE: o.PreRunE,
		RunE:              cmdutil.RunEHandler(o.RunE),
	}
	cmd.Flags().StringVar(&o.NetRuleName, "name", "", "network rule name (Required)")
	return cmd
}

func (o *RemoveNetRuleOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *RemoveNetRuleOption) Validate(cmd *cobra.Command, args []string) error {
	if o.AllNamespace {
		return errors.New("--all-namespaces is not supported in this command")
	}
	if err := o.UserNamespacedCliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	if o.NetRuleName == "" {
		return errors.New("network rule name is required")
	}
	return nil
}

func (o *RemoveNetRuleOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.WorkspaceName = args[0]
	return nil
}

func (o *RemoveNetRuleOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	if _, err := c.DeleteNetworkRule(ctx, o.WorkspaceName, o.User, o.NetRuleName); err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully remove network rule '%s' for workspace '%s'\n", o.NetRuleName, o.WorkspaceName)
	return nil
}
