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
)

type resetPasswordOption struct {
	*changePasswordOption

	Force  bool
	Silent bool
}

func resetPasswordCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &resetPasswordOption{changePasswordOption: &changePasswordOption{RootOptions: cliOpt}}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().BoolVar(&o.Force, "force", false, "not ask confirmation")
	cmd.Flags().BoolVar(&o.Silent, "silent", false, "only output new password")
	return cmd
}

func (o *resetPasswordOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	if !o.UseKubeAPI {
		return errors.New("force reset is only available with -k")
	}
	return nil
}

func (o *resetPasswordOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.UserName = args[0]

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return nil
}

func (o *resetPasswordOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	if err := o.ValidateUser(ctx); err != nil {
		return err
	}

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

	newPassword, err := o.resetPasswordWithKubeClient(ctx)
	if err != nil {
		return err
	}

	if o.Silent {
		fmt.Fprintln(cmd.OutOrStdout(), *newPassword)
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("Successfully reset password: user %s", o.UserName))
		fmt.Fprintln(cmd.OutOrStdout(), "New password:", *newPassword)
	}

	return nil
}

func (o *resetPasswordOption) resetPasswordWithKubeClient(ctx context.Context) (*string, error) {
	c := o.KosmoClient
	if err := c.ResetPassword(ctx, o.UserName); err != nil {
		return nil, err
	}
	pass, err := c.GetDefaultPassword(ctx, o.UserName)
	if err != nil {
		return nil, err
	}
	if pass == nil {
		return nil, errors.New("password is nil")
	}
	return pass, nil
}
