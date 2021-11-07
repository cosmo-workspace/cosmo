package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/api/equality"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
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
		RunE:              o.RunE,
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	user, err := c.GetUser(ctx, o.UserID)
	if err != nil {
		return err
	}
	o.Logr.DebugAll().Info("GetUser", "user", user)

	before := user.DeepCopy()

	if o.Name != "" {
		user.Spec.DisplayName = o.Name
		o.Logr.Debug().Info("name changed", "name", o.Name)
	}

	if o.Role != "-" {
		user.Spec.Role = o.role
	}
	o.Logr.Debug().Info("role changed", "role", o.role)

	if equality.Semantic.DeepEqual(before, user) {
		return errors.New("no change")
	}

	if err := c.Update(ctx, user); err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully updated user %s\n", o.UserID)
	return nil
}
