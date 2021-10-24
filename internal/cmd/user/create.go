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
)

type createOption struct {
	*cmdutil.CliOptions

	UserID      string
	DisplayName string
	Role        string
	Admin       bool

	user *wsv1alpha1.User
}

func createCmd(cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &createOption{CliOptions: cliOpt}
	cmd := &cobra.Command{
		Use:               "create USER_ID --role cosmo-admin",
		Short:             "Create user",
		PersistentPreRunE: o.PreRunE,
		RunE:              o.RunE,
	}
	cmd.Flags().StringVar(&o.DisplayName, "name", "", "user display name (default: same as USER_ID)")
	cmd.Flags().StringVar(&o.Role, "role", "", "user role")
	cmd.Flags().BoolVar(&o.Admin, "admin", false, "user admin role")
	return cmd
}

func (o *createOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *createOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	if o.Role != "" {
		if o.Admin {
			return errors.New("--role and --admin is not used at the same time")
		}
		if !wsv1alpha1.UserRole(o.Role).IsValid() {
			return fmt.Errorf("role %s is invalid", o.Role)
		}
	}
	return nil
}

func (o *createOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}

	o.UserID = args[0]

	user := wsv1alpha1.User{}
	user.ID = o.UserID

	if o.DisplayName != "" {
		user.DisplayName = o.DisplayName
	} else {
		user.DisplayName = o.UserID
	}

	if o.Admin {
		o.Role = wsv1alpha1.UserAdminRole.String()
	}

	user.Role = wsv1alpha1.UserRole(o.Role)
	o.user = &user
	return nil
}

func (o *createOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	if _, err := c.CreateUser(ctx, o.user); err != nil {
		return err
	}

	// Wait until user created
	tk := time.NewTicker(time.Second)
	defer tk.Stop()
	var defaultPassword *string

UserCreationWaitLoop:
	for {
		p, err := c.GetDefaultPassword(ctx, o.UserID)
		if err == nil {
			tk.Stop()
			defaultPassword = p
			break UserCreationWaitLoop
		}

		select {
		case <-ctx.Done():
			tk.Stop()
			return fmt.Errorf("reached to timeout in user creation")

		default:
			<-tk.C
		}
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully created user %s\n", o.UserID)
	fmt.Fprintln(o.Out, "Default password:", *defaultPassword)
	return nil
}
