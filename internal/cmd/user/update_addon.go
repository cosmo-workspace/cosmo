package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type UpdateAddonOption struct {
	*cli.RootOptions

	UserName string
	Addons   []string
	Force    bool

	userAddons []*dashv1alpha1.UserAddon
}

func UpdateAddonCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &UpdateAddonOption{RootOptions: cliOpt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().StringArrayVar(&o.Addons, "addon", nil, "user addons\nformat is '--addon TEMPLATE_NAME1,KEY=VAL,KEY=VAL --addon TEMPLATE_NAME2,KEY=VAL ...' ")
	cmd.MarkFlagsOneRequired("addon")
	cmd.Flags().BoolVar(&o.Force, "force", false, "not ask confirmation")
	return cmd
}

func (o *UpdateAddonOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *UpdateAddonOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}

	o.UserName = args[0]

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

func (o *UpdateAddonOption) RunE(cmd *cobra.Command, args []string) error {
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

	o.Logr.Info("updating user", "userName", o.UserName, "currentAddons", apiconv.D2S_UserAddons(currentUser.Addons), "newAddons", o.Addons)

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
		user, err = o.UpdateUserWithKubeClient(ctx)
	} else {
		user, err = o.UpdateUserWithDashClient(ctx)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("Successfully updated user %s", o.UserName))
	OutputWideTable(cmd.OutOrStdout(), []*dashv1alpha1.User{user})

	return nil
}

func (o *UpdateAddonOption) UpdateUserWithDashClient(ctx context.Context) (*dashv1alpha1.User, error) {
	c := o.CosmoDashClient

	req := &dashv1alpha1.UpdateUserAddonsRequest{
		UserName: o.UserName,
		Addons:   o.userAddons,
	}
	res, err := c.UserServiceClient.UpdateUserAddons(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("UserServiceClient.UpdateUserAddons", "res", res)

	return res.Msg.User, nil
}

func (o *UpdateAddonOption) UpdateUserWithKubeClient(ctx context.Context) (*dashv1alpha1.User, error) {
	c := o.KosmoClient
	opts := kosmo.UpdateUserOpts{
		UserAddons: apiconv.D2C_UserAddons(o.userAddons),
	}

	user, err := c.UpdateUser(ctx, o.UserName, opts)
	if err != nil {
		return nil, err
	}
	d := apiconv.C2D_User(*user)

	return d, nil
}

func (o *UpdateAddonOption) GetUserWithDashClient(ctx context.Context) (*dashv1alpha1.User, error) {
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

func (o *UpdateAddonOption) GetUserWithKubeClient(ctx context.Context) (*dashv1alpha1.User, error) {
	c := o.KosmoClient
	user, err := c.GetUser(ctx, o.UserName)
	if err != nil {
		return nil, err
	}
	d := apiconv.C2D_User(*user)

	return d, nil
}
