package workspace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type UpdateDeletePolicyOption struct {
	*cli.RootOptions

	WorkspaceName string
	UserName      string
	DeletePolicy  string
}

func UpdateDeletePolicyCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &UpdateDeletePolicyOption{RootOptions: cliOpt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().StringVarP(&o.UserName, "user", "u", "", "user name (defualt: login user)")
	return cmd
}

func (o *UpdateDeletePolicyOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 2 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *UpdateDeletePolicyOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}

	o.WorkspaceName = args[0]
	o.DeletePolicy = args[1]

	if !o.UseKubeAPI && o.UserName == "" {
		o.UserName = o.CliConfig.User
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return nil
}

func (o *UpdateDeletePolicyOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	o.Logr.Info("updating delete polocy", "workspace", o.WorkspaceName, "userName", o.UserName, "deletepolicy", o.DeletePolicy)

	var (
		ws  *dashv1alpha1.Workspace
		err error
	)
	if o.UseKubeAPI {
		ws, err = o.UpdateWorkspaceWithKubeClient(ctx)
	} else {
		ws, err = o.UpdateWorkspaceWithDashClient(ctx)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("Successfully updated workspace %s", o.UserName))
	OutputWideTable(cmd.OutOrStdout(), o.UserName, []*dashv1alpha1.Workspace{ws})

	return nil
}

func (o *UpdateDeletePolicyOption) UpdateWorkspaceWithDashClient(ctx context.Context) (*dashv1alpha1.Workspace, error) {
	req := &dashv1alpha1.UpdateWorkspaceRequest{
		WsName:       o.WorkspaceName,
		UserName:     o.UserName,
		DeletePolicy: apiconv.C2D_DeletePolicy(o.DeletePolicy),
	}
	c := o.CosmoDashClient
	o.Logr.DebugAll().Info("WorkspaceServiceClient.UpdateWorkspace", "req", req)
	res, err := c.WorkspaceServiceClient.UpdateWorkspace(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("WorkspaceServiceClient.UpdateWorkspace", "res", res)

	return res.Msg.Workspace, nil
}

func (o *UpdateDeletePolicyOption) UpdateWorkspaceWithKubeClient(ctx context.Context) (*dashv1alpha1.Workspace, error) {
	opts := kosmo.UpdateWorkspaceOpts{
		DeletePolicy: &o.DeletePolicy,
	}
	c := o.KosmoClient
	o.Logr.DebugAll().Info("UpdateWorkspace", "userName", o.UserName, "workspaceName", o.WorkspaceName, "opts", opts)
	ws, err := c.UpdateWorkspace(ctx, o.UserName, o.WorkspaceName, opts)
	if err != nil {
		return nil, err
	}
	return apiconv.C2D_Workspace(*ws), nil
}
