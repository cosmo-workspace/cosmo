package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type DeleteOption struct {
	*cli.RootOptions

	UserNames []string
	Force     bool
}

func DeleteCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &DeleteOption{RootOptions: cliOpt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().BoolVar(&o.Force, "force", false, "not ask confirmation")
	return cmd
}

func (o *DeleteOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *DeleteOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.UserNames = args

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return nil
}

func (o *DeleteOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	o.Logr.Info("deleting users", "users", o.UserNames)

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

	for _, v := range o.UserNames {
		if o.UseKubeAPI {
			if err := o.DeleteUserWithKubeClient(ctx, v); err != nil {
				return err
			}
		} else {
			if err := o.DeleteUserWithDashClient(ctx, v); err != nil {
				return err
			}
		}
		fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("Successfully deleted user %s", v))
	}

	return nil
}

func (o *DeleteOption) DeleteUserWithDashClient(ctx context.Context, userName string) error {
	req := &dashv1alpha1.DeleteUserRequest{
		UserName: userName,
	}
	c := o.CosmoDashClient
	res, err := c.UserServiceClient.DeleteUser(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("UserServiceClient.DeleteUser", "res", res)

	return nil
}

func (o *DeleteOption) DeleteUserWithKubeClient(ctx context.Context, userName string) error {
	c := o.KosmoClient
	if _, err := c.DeleteUser(ctx, userName); err != nil {
		return err
	}
	return nil
}
