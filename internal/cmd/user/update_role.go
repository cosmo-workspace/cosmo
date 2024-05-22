package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type UpdateRoleOption struct {
	*cli.RootOptions

	UserName       string
	Roles          []string
	PrivilegedRole bool
	Force          bool

	userAddons []*dashv1alpha1.UserAddon
}

func UpdateRoleCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &UpdateRoleOption{RootOptions: cliOpt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().StringSliceVar(&o.Roles, "role", nil, "user roles")
	cmd.MarkFlagsOneRequired("role")
	cmd.Flags().BoolVar(&o.PrivilegedRole, "privileged", false, "add cosmo-admin role (privileged)")
	cmd.Flags().BoolVar(&o.Force, "force", false, "not ask confirmation")
	return cmd
}

func (o *UpdateRoleOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *UpdateRoleOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}

	o.UserName = args[0]

	if o.PrivilegedRole {
		o.Roles = []string{cosmov1alpha1.PrivilegedRoleName}
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return nil
}

func (o *UpdateRoleOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	var (
		currentUser *dashv1alpha1.User
		err         error
	)
	if o.UseKubeAPI {
		currentUser, err = o.GetUserWithKubeClient(ctx)
	} else {
		currentUser, err = o.GetUserWithDashClient(ctx)
	}
	if err != nil {
		return err
	}

	o.Logr.Info("updating user roles", "userName", o.UserName, "currentRole", currentUser.Roles, "newRole", o.Roles)

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

	var user *dashv1alpha1.User
	if o.UseKubeAPI {
		user, err = o.UpdateUserRoleWithKubeClient(ctx)
	} else {
		user, err = o.UpdateUserRoleWithDashClient(ctx)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("Successfully updated user %s", o.UserName))
	OutputWideTable(cmd.OutOrStdout(), []*dashv1alpha1.User{user})

	return nil
}

func (o *UpdateRoleOption) UpdateUserRoleWithDashClient(ctx context.Context) (*dashv1alpha1.User, error) {
	req := &dashv1alpha1.UpdateUserRoleRequest{
		UserName: o.UserName,
		Roles:    o.Roles,
	}
	c := o.CosmoDashClient
	res, err := c.UserServiceClient.UpdateUserRole(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("UserServiceClient.UpdateUserRole", "res", res)

	return res.Msg.User, nil
}

func (o *UpdateRoleOption) UpdateUserRoleWithKubeClient(ctx context.Context) (*dashv1alpha1.User, error) {
	c := o.KosmoClient
	opts := kosmo.UpdateUserOpts{
		UserRoles: apiconv.S2C_UserRoles(o.Roles),
	}
	user, err := c.UpdateUser(ctx, o.UserName, opts)
	if err != nil {
		return nil, err
	}
	d := apiconv.C2D_User(*user)

	return d, nil
}

func (o *UpdateRoleOption) GetUserWithDashClient(ctx context.Context) (*dashv1alpha1.User, error) {
	req := &dashv1alpha1.GetUserRequest{
		UserName: o.UserName,
	}
	c := o.CosmoDashClient
	res, err := c.UserServiceClient.GetUser(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("UserServiceClient.GetUser", "res", res)

	return res.Msg.User, nil
}

func (o *UpdateRoleOption) GetUserWithKubeClient(ctx context.Context) (*dashv1alpha1.User, error) {
	c := o.KosmoClient
	user, err := c.GetUser(ctx, o.UserName)
	if err != nil {
		return nil, err
	}
	d := apiconv.C2D_User(*user)

	return d, nil
}
