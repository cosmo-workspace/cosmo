package workspace

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type UpdateVarsOption struct {
	*cli.RootOptions

	WorkspaceName string
	UserName      string
	TemplateVars  []string
	Force         bool

	vars map[string]string
}

func UpdateVarsCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &UpdateVarsOption{RootOptions: cliOpt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().StringVarP(&o.UserName, "user", "u", "", "user name (defualt: login user)")
	cmd.Flags().StringSliceVar(&o.TemplateVars, "set", []string{}, "template vars. the format is VarName=VarValue (example: --set VAR1=VAL1 --set VAR2=VAL2)")
	cmd.Flags().BoolVar(&o.Force, "force", false, "not ask confirmation")

	return cmd
}

func (o *UpdateVarsOption) Validate(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	if o.UseKubeAPI && o.UserName == "" {
		return fmt.Errorf("user name is required")
	}
	return nil
}

func (o *UpdateVarsOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.WorkspaceName = args[0]

	if !o.UseKubeAPI && o.UserName == "" {
		o.UserName = o.CliConfig.User
	}

	if len(o.TemplateVars) > 0 {
		vars := make(map[string]string)
		for _, v := range o.TemplateVars {
			varAndVal := strings.Split(v, "=")
			if len(varAndVal) != 2 {
				return fmt.Errorf("vars format error: vars %s must be 'VAR=VAL'", v)
			}
			vars[varAndVal[0]] = varAndVal[1]
		}
		o.vars = vars
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return nil
}

func (o *UpdateVarsOption) RunE(cmd *cobra.Command, args []string) error {
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
		currentWs *dashv1alpha1.Workspace
		err       error
	)
	if o.UseKubeAPI {
		currentWs, err = o.GetWorkspaceWithKubeClient(ctx)
	} else {
		currentWs, err = o.GetWorkspaceWithDashClient(ctx)
	}
	if err != nil {
		return err
	}

	o.Logr.Info("updating workspace", "user", o.UserName, "name", o.WorkspaceName, "currentVars", currentWs.Spec.Vars, "newVars", o.vars)

	if !o.Force {
	AskLoop:
		for {
			input, err := cli.AskInput("Confirm? [y/n] ", false)
			if err != nil {
				return err
			}
			switch strings.ToLower(input) {
			case "y":
				break AskLoop
			case "n":
				fmt.Println("canceled")
				return nil
			}
		}
	}

	var ws *dashv1alpha1.Workspace
	if o.UseKubeAPI {
		ws, err = o.UpdateWorkspaceWithKubeClient(ctx)
	} else {
		ws, err = o.UpdateWorkspaceWithDashClient(ctx)
	}
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("Successfully updated workspace %s", o.WorkspaceName))
	OutputTable(cmd.OutOrStdout(), o.UserName, []*dashv1alpha1.Workspace{ws})

	return nil
}

func (o *UpdateVarsOption) UpdateWorkspaceWithDashClient(ctx context.Context) (*dashv1alpha1.Workspace, error) {
	req := &dashv1alpha1.UpdateWorkspaceRequest{
		WsName:   o.WorkspaceName,
		UserName: o.UserName,
		Vars:     o.vars,
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

func (o *UpdateVarsOption) UpdateWorkspaceWithKubeClient(ctx context.Context) (*dashv1alpha1.Workspace, error) {
	opts := kosmo.UpdateWorkspaceOpts{
		Vars: o.vars,
	}
	c := o.KosmoClient
	o.Logr.DebugAll().Info("UpdateWorkspace", "userName", o.UserName, "workspaceName", o.WorkspaceName, "opts", opts)
	ws, err := c.UpdateWorkspace(ctx, o.UserName, o.WorkspaceName, opts)
	if err != nil {
		return nil, err
	}
	return apiconv.C2D_Workspace(*ws), nil
}

func (o *UpdateVarsOption) GetWorkspaceWithDashClient(ctx context.Context) (*dashv1alpha1.Workspace, error) {
	req := &dashv1alpha1.GetWorkspaceRequest{
		WsName:   o.WorkspaceName,
		UserName: o.UserName,
	}
	c := o.CosmoDashClient
	o.Logr.DebugAll().Info("WorkspaceServiceClient.GetWorkspace", "req", req)
	res, err := c.WorkspaceServiceClient.GetWorkspace(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("WorkspaceServiceClient.GetWorkspace", "res", res)

	return res.Msg.Workspace, nil
}

func (o *UpdateVarsOption) GetWorkspaceWithKubeClient(ctx context.Context) (*dashv1alpha1.Workspace, error) {
	c := o.KosmoClient
	ws, err := c.GetWorkspace(ctx, o.UserName, o.WorkspaceName)
	if err != nil {
		return nil, err
	}
	return apiconv.C2D_Workspace(*ws), nil
}
