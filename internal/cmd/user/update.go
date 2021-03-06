package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

type updateOption struct {
	*cmdutil.CliOptions

	UserID string
	Name   string
	Role   string
	role   wsv1alpha1.UserRole
}

func updateCmd(cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &updateOption{CliOptions: cliOpt}
	cmd := &cobra.Command{
		Use:               "update USER_ID --role ROLE --name NAME",
		Short:             "Update user",
		PersistentPreRunE: o.PreRunE,
		RunE:              cmdutil.RunEHandler(o.RunE),
	}
	cmd.Flags().StringVar(&o.Name, "name", "", "user name")
	cmd.Flags().StringVar(&o.Role, "role", "-", "user role")
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
	if o.Role != "-" {
		if o.role = wsv1alpha1.UserRole(o.Role); !o.role.IsValid() {
			return fmt.Errorf("role %s is invalid", o.Role)
		}
	}
	return nil
}

func (o *updateOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.UserID = args[0]
	return nil
}

func (o *updateOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()

	ctx = clog.IntoContext(ctx, o.Logr)
	_, err := o.Client.UpdateUser(ctx, o.UserID, kosmo.UpdateUserOpts{
		DisplayName: &o.Name,
		UserRole:    &o.Role,
	})
	if err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully updated user %s\n", o.UserID)
	return nil
}
