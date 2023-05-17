package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type resetPasswordOption struct {
	*cmdutil.CliOptions

	UserName string
	Password string
}

func resetPasswordCmd(cmd *cobra.Command, cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &resetPasswordOption{CliOptions: cliOpt}
	cmd.PersistentPreRunE = o.PreRunE
	cmd.RunE = cmdutil.RunEHandler(o.RunE)
	cmd.Flags().StringVar(&o.Password, "password", "", "new password (default: random string)")
	return cmd
}

func (o *resetPasswordOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *resetPasswordOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *resetPasswordOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.UserName = args[0]
	return nil
}

func (o *resetPasswordOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	if o.Password == "" {
		if err := c.ResetPassword(ctx, o.UserName); err != nil {
			return err
		}
	} else {
		if err := c.RegisterPassword(ctx, o.UserName, []byte(o.Password)); err != nil {
			return err
		}
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully reset password: user %s\n", o.UserName)

	if o.Password == "" {
		pass, err := c.GetDefaultPassword(ctx, o.UserName)
		if err != nil {
			return err
		}
		fmt.Fprintln(o.Out, "New password:", *pass)
	}

	return nil
}
