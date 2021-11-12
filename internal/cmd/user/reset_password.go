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

	UserID string
}

func resetPasswordCmd(cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &resetPasswordOption{CliOptions: cliOpt}
	cmd := &cobra.Command{
		Use:               "reset-password USER_ID",
		Short:             "Reset user password",
		PersistentPreRunE: o.PreRunE,
		RunE:              o.RunE,
	}
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
	o.UserID = args[0]
	return nil
}

func (o *resetPasswordOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	if err := c.ResetPassword(ctx, o.UserID); err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully reset password: user %s\n", o.UserID)

	pass, err := c.GetDefaultPassword(ctx, o.UserID)
	if err != nil {
		return err
	}

	fmt.Fprintln(o.Out, "New password:", *pass)

	return nil
}
