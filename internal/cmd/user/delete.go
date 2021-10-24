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

type deleteOption struct {
	*cmdutil.CliOptions

	UserID string
}

func deleteCmd(cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &deleteOption{CliOptions: cliOpt}
	cmd := &cobra.Command{
		Use:               "delete USER_ID",
		Aliases:           []string{"del"},
		Short:             "Delete user",
		PersistentPreRunE: o.PreRunE,
		RunE:              o.RunE,
	}
	return cmd
}

func (o *deleteOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *deleteOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *deleteOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.UserID = args[0]
	return nil
}

func (o *deleteOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	if _, err := c.DeleteUser(ctx, o.UserID); err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully deleted user %s\n", o.UserID)
	return nil
}
