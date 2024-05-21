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
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type CreateOption struct {
	*cli.RootOptions

	UserName       string
	DisplayName    string
	Roles          []string
	AuthType       string
	PrivilegedRole bool
	Addons         []string
	Force          bool

	userAddons []*dashv1alpha1.UserAddon
}

func CreateCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &CreateOption{RootOptions: cliOpt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().StringVar(&o.DisplayName, "display-name", "", "user display name (default: same as USER_NAME)")
	cmd.Flags().StringSliceVar(&o.Roles, "role", nil, "user roles")
	cmd.Flags().StringVar(&o.AuthType, "auth-type", cosmov1alpha1.UserAuthTypePasswordSecert.String(), "user auth type 'password-secret'(default),'ldap'")
	cmd.Flags().BoolVar(&o.PrivilegedRole, "privileged", false, "add cosmo-admin role (privileged)")
	cmd.Flags().StringArrayVar(&o.Addons, "addon", nil, "user addons\nformat is '--addon TEMPLATE_NAME1,KEY=VAL,KEY=VAL --addon TEMPLATE_NAME2,KEY=VAL ...' ")
	cmd.Flags().BoolVar(&o.Force, "force", false, "not ask confirmation")
	return cmd
}

func (o *CreateOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	if !cosmov1alpha1.UserAuthType(o.AuthType).IsValid() {
		return fmt.Errorf("invalid auth-type: %s", o.AuthType)
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *CreateOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}

	o.UserName = args[0]

	if o.PrivilegedRole {
		o.Roles = []string{cosmov1alpha1.PrivilegedRoleName}
	}

	o.userAddons = make([]*dashv1alpha1.UserAddon, 0, len(o.Addons))
	if len(o.Addons) > 0 {
		userAddons, err := apiconv.S2D_UserAddons(o.Addons)
		if err != nil {
			return err
		}
		o.userAddons = append(o.userAddons, userAddons...)
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return nil
}

func (o *CreateOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	o.Logr.Info("creating user", "userName", o.UserName, "displayName", o.DisplayName, "roles", o.Roles, "authType", o.AuthType, "addons", o.Addons)

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

	var (
		user *dashv1alpha1.User
		err  error
	)
	if o.UseKubeAPI {
		user, err = o.CreateUserWithKubeClient(ctx)
	} else {
		user, err = o.CreateUserWithDashClient(ctx)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("Successfully created user %s", o.UserName))
	OutputTable(cmd.OutOrStdout(), []*dashv1alpha1.User{user})

	if o.AuthType == cosmov1alpha1.UserAuthTypePasswordSecert.String() {
		fmt.Fprintln(cmd.OutOrStdout(), "Default password:", user.DefaultPassword)
	}
	return nil
}

func (o *CreateOption) CreateUserWithDashClient(ctx context.Context) (*dashv1alpha1.User, error) {
	req := &dashv1alpha1.CreateUserRequest{
		UserName:    o.UserName,
		DisplayName: o.DisplayName,
		Roles:       o.Roles,
		AuthType:    o.AuthType,
		Addons:      o.userAddons,
	}
	c := o.CosmoDashClient
	res, err := c.UserServiceClient.CreateUser(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("UserServiceClient.CreateUser", "res", res)

	return res.Msg.User, nil
}

func (o *CreateOption) CreateUserWithKubeClient(ctx context.Context) (*dashv1alpha1.User, error) {
	c := o.KosmoClient
	user, err := c.CreateUser(ctx, o.UserName, o.DisplayName, o.Roles, o.AuthType, apiconv.D2C_UserAddons(o.userAddons))
	if err != nil {
		return nil, err
	}
	d := apiconv.C2D_User(*user)

	if o.AuthType == cosmov1alpha1.UserAuthTypePasswordSecert.String() {
		defaultPassword, err := c.GetDefaultPasswordAwait(ctx, o.UserName)
		if err != nil {
			return nil, err
		}
		d.DefaultPassword = *defaultPassword
	}
	return d, nil
}
