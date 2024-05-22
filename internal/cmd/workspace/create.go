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
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type CreateOption struct {
	*cli.RootOptions

	WorkspaceName string
	UserName      string
	Template      string
	TemplateVars  []string
	Force         bool

	vars map[string]string
}

func CreateCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &CreateOption{RootOptions: cliOpt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().StringVarP(&o.UserName, "user", "u", "", "user name (defualt: login user)")
	cmd.Flags().StringVarP(&o.Template, "template", "t", "", "template name (Required)")
	cmd.MarkFlagRequired("template")
	cmd.Flags().StringSliceVar(&o.TemplateVars, "set", []string{}, "template vars. the format is VarName=VarValue (example: --set VAR1=VAL1 --set VAR2=VAL2)")
	cmd.Flags().BoolVar(&o.Force, "force", false, "not ask confirmation")

	return cmd
}

func (o *CreateOption) Validate(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	if o.UseKubeAPI && o.UserName == "" {
		return fmt.Errorf("user name is required")
	}
	return nil
}

func (o *CreateOption) Complete(cmd *cobra.Command, args []string) error {
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

func (o *CreateOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	o.Logr.Info("creating workspace", "user", o.UserName, "name", o.WorkspaceName, "template", o.Template, "vars", o.TemplateVars)

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

	var (
		ws  *dashv1alpha1.Workspace
		err error
	)
	if o.UseKubeAPI {
		ws, err = o.CreateWorkspaceWithKubeClient(ctx)
	} else {
		ws, err = o.CreateWorkspaceWithDashClient(ctx)
	}
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("Successfully created workspace %s", o.WorkspaceName))
	OutputTable(cmd.OutOrStdout(), []*dashv1alpha1.Workspace{ws})

	return nil
}

func (o *CreateOption) CreateWorkspaceWithDashClient(ctx context.Context) (*dashv1alpha1.Workspace, error) {
	req := &dashv1alpha1.CreateWorkspaceRequest{
		WsName:   o.WorkspaceName,
		UserName: o.UserName,
		Template: o.Template,
		Vars:     o.vars,
	}
	c := o.CosmoDashClient
	res, err := c.WorkspaceServiceClient.CreateWorkspace(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("WorkspaceServiceClient.CreateWorkspace", "res", res)

	return res.Msg.Workspace, nil
}

func (o *CreateOption) CreateWorkspaceWithKubeClient(ctx context.Context) (*dashv1alpha1.Workspace, error) {
	c := o.KosmoClient
	ws, err := c.CreateWorkspace(ctx, o.UserName, o.WorkspaceName, o.Template, o.vars)
	if err != nil {
		return nil, err
	}
	return apiconv.C2D_Workspace(*ws), nil
}
