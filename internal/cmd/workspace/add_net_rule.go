package workspace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/utils/pointer"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type addNetRuleOption struct {
	*cmdutil.UserNamespacedCliOptions

	WorkspaceName string
	NetRuleName   string
	PortNumber    int
	Group         string
	HTTPPath      string
	Public        bool

	rule wsv1alpha1.NetworkRule
}

func addNetRuleCmd(cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &addNetRuleOption{UserNamespacedCliOptions: cliOpt}

	cmd := &cobra.Command{
		Use:               "add-net-rule WORKSPACE_NAME --name NETWORK_RULE_NAME --port PORT_NUMBER",
		Short:             "Update or insert workspace network rule",
		Aliases:           []string{"add-net"},
		PersistentPreRunE: o.PreRunE,
		RunE:              cmdutil.RunEHandler(o.RunE),
	}
	cmd.Flags().StringVar(&o.NetRuleName, "name", "", "network rule name (Required)")
	cmd.Flags().IntVar(&o.PortNumber, "port", 0, "serivce port number (Required)")
	cmd.Flags().StringVar(&o.Group, "group", "", "group of ports for URLVar. Ports in the same group are treated as the same domain. set 'name' value if empty")
	cmd.Flags().StringVar(&o.HTTPPath, "path", "/", "path for Ingress path when using ingress")
	cmd.Flags().BoolVar(&o.Public, "public", false, "disable authentication for this port")

	return cmd
}

func (o *addNetRuleOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *addNetRuleOption) Validate(cmd *cobra.Command, args []string) error {
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
		return errors.New("--name is required")
	}
	if o.PortNumber == 0 {
		return errors.New("--port is required")
	}
	return nil
}

func (o *addNetRuleOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.WorkspaceName = args[0]

	if o.Group == "" {
		o.Group = o.NetRuleName
	}

	o.rule = wsv1alpha1.NetworkRule{
		PortName:   o.NetRuleName,
		PortNumber: o.PortNumber,
		HTTPPath:   o.HTTPPath,
		Group:      pointer.String(o.Group),
		Public:     o.Public,
	}
	return nil
}

func (o *addNetRuleOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	if _, err := c.AddNetworkRule(ctx, o.WorkspaceName, o.User, o.rule.PortName,
		o.rule.PortNumber, o.rule.Group, o.rule.HTTPPath, o.rule.Public); err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully add network rule '%s' for workspace '%s'\n", o.NetRuleName, o.WorkspaceName)
	return nil
}
