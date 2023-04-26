package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

type updateOption struct {
	*cmdutil.CliOptions

	UserName string
	Name     string
	Role     []string

	displayName *string
}

func updateCmd(cmd *cobra.Command, cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &updateOption{CliOptions: cliOpt}
	cmd.PersistentPreRunE = o.PreRunE
	cmd.RunE = cmdutil.RunEHandler(o.RunE)
	cmd.Flags().StringVar(&o.Name, "name", "-", "user name")
	cmd.Flags().StringSliceVar(&o.Role, "role", []string{"-"}, "user role")
	return cmd
}

func (o *updateOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *updateOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *updateOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.UserName = args[0]

	if o.Name != "-" {
		o.displayName = &o.Name
	}
	return nil
}

func (o *updateOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()

	ctx = clog.IntoContext(ctx, o.Logr)
	_, err := o.Client.UpdateUser(ctx, o.UserName, kosmo.UpdateUserOpts{
		DisplayName: o.displayName,
		UserRoles:   o.Role,
	})
	if err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully updated user %s\n", o.UserName)
	return nil
}
