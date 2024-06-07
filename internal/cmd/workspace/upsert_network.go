package workspace

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type UpsertNetworkOption struct {
	*cli.RootOptions

	WorkspaceName    string
	UserName         string
	CustomHostPrefix string
	PortNumber       int32
	HTTPPath         string
	Public           bool

	rule cosmov1alpha1.NetworkRule
}

func UpsertNetworkCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &UpsertNetworkOption{RootOptions: cliOpt}

	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().StringVarP(&o.UserName, "user", "u", "", "user name (defualt: login user)")
	cmd.Flags().Int32Var(&o.PortNumber, "port", 0, "serivce port number (Required)")
	cmd.MarkFlagRequired("port")
	cmd.Flags().StringVar(&o.CustomHostPrefix, "host-prefix", "", "custom host prefix")
	cmd.Flags().StringVar(&o.HTTPPath, "path", "/", "path for Ingress path when using ingress")
	cmd.Flags().BoolVar(&o.Public, "public", false, "disable authentication for this port")

	return cmd
}

func (o *UpsertNetworkOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	if o.UseKubeAPI && o.UserName == "" {
		return fmt.Errorf("user name is required")
	}
	return nil
}

func (o *UpsertNetworkOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}
	if len(args) > 0 {
		o.WorkspaceName = args[0]
	} else if cli.UseServiceAccount(o.CliConfig) {
		o.WorkspaceName = cli.GetCurrentWorkspaceName()
		o.Logr.Info("Workspace name is auto detected from hostname", "name", o.WorkspaceName)
	}
	if !o.UseKubeAPI && o.UserName == "" {
		o.UserName = o.CliConfig.User
	}

	o.rule = cosmov1alpha1.NetworkRule{
		CustomHostPrefix: o.CustomHostPrefix,
		PortNumber:       o.PortNumber,
		HTTPPath:         o.HTTPPath,
		Public:           o.Public,
	}
	o.rule.Default()

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return nil
}

func (o *UpsertNetworkOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	var (
		rule *dashv1alpha1.NetworkRule
		err  error
	)
	if o.UseKubeAPI {
		rule, err = o.UpsertNetworkRuleByKubeClient(ctx)
	} else {
		rule, err = o.UpsertNetworkRuleWithDashClient(ctx)
	}
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("Successfully upsert network rule for workspace '%s'", o.WorkspaceName))

	o.OutputTable(cmd.OutOrStdout(), rule)
	return nil
}

func (o *UpsertNetworkOption) UpsertNetworkRuleWithDashClient(ctx context.Context) (*dashv1alpha1.NetworkRule, error) {
	reqGet := &dashv1alpha1.GetWorkspaceRequest{
		WsName:   o.WorkspaceName,
		UserName: o.UserName,
	}
	c := o.CosmoDashClient
	resGet, err := c.WorkspaceServiceClient.GetWorkspace(ctx, cli.NewRequestWithToken(reqGet, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}

	rules := apiconv.D2C_NetworkRules(resGet.Msg.Workspace.Spec.Network)
	index := cosmov1alpha1.GetNetworkRuleIndex(rules, o.rule)

	req := &dashv1alpha1.UpsertNetworkRuleRequest{
		WsName:      o.WorkspaceName,
		UserName:    o.UserName,
		NetworkRule: apiconv.C2D_NetworkRule(o.rule),
		Index:       int32(index),
	}
	o.Logr.DebugAll().Info("WorkspaceServiceClient.UpsertNetworkRule", "req", req)
	res, err := c.WorkspaceServiceClient.UpsertNetworkRule(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("WorkspaceServiceClient.UpsertNetworkRule", "res", res)
	return res.Msg.NetworkRule, nil
}

func (o *UpsertNetworkOption) UpsertNetworkRuleByKubeClient(ctx context.Context) (*dashv1alpha1.NetworkRule, error) {
	c := o.KosmoClient

	ws, err := c.GetWorkspaceByUserName(ctx, o.WorkspaceName, o.UserName)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace: %v", err)
	}
	index := cosmov1alpha1.GetNetworkRuleIndex(ws.Spec.Network, o.rule)

	o.Logr.DebugAll().Info("AddNetworkRule", "workspaceName", o.WorkspaceName, "userName", o.UserName, "rule", o.rule, "index", index)
	cr, err := c.AddNetworkRule(ctx, o.WorkspaceName, o.UserName, o.rule, index)
	if err != nil {
		return nil, err
	}
	return apiconv.C2D_NetworkRule(*cr), nil
}

func (o *UpsertNetworkOption) OutputTable(w io.Writer, v *dashv1alpha1.NetworkRule) {
	data := [][]string{
		{fmt.Sprintf("%d", v.PortNumber), v.CustomHostPrefix, v.HttpPath, strconv.FormatBool(v.Public), v.Url},
	}
	cli.OutputTable(w,
		[]string{"PORT", "CUSTOM_HOST_PREFIX", "HTTP_PATH", "PUBLIC", "URL"},
		data)
}
