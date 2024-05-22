package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type changePasswordOption struct {
	*cli.RootOptions

	UserName      string
	PasswordStdin bool

	currentPassword string
	newPassword     string
}

func changePasswordCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &changePasswordOption{RootOptions: cliOpt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().BoolVar(&o.PasswordStdin, "password-stdin", false, "input new password from stdin pipe")
	return cmd
}

func (o *changePasswordOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	if o.UseKubeAPI && len(args) < 1 {
		return fmt.Errorf("user name is required")
	}
	return nil
}

func (o *changePasswordOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}
	if len(args) > 0 {
		o.UserName = args[0]
	}
	if !o.UseKubeAPI && o.UserName == "" {
		o.UserName = o.CliConfig.User
		o.Logr.Info(fmt.Sprintf("Change login user password: %s", o.UserName))
	}

	if err := o.ValidateUser(o.Ctx); err != nil {
		return err
	}

	if o.PasswordStdin {
		if !o.UseKubeAPI {
			return errors.New("--password-stdin is only supported with -k")
		}
		input, err := cli.ReadFromPipedStdin()
		if err != nil {
			return fmt.Errorf("failed to read from stdin pipe: %w", err)
		}
		o.newPassword = input

	} else {
		input, err := cli.AskInput("Current password: ", true)
		if err != nil {
			return err
		}
		o.currentPassword = input

		input, err = cli.AskInput("New password    : ", true)
		if err != nil {
			return err
		}
		o.newPassword = input
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return nil
}

func (o *changePasswordOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	if o.UseKubeAPI {
		if err := o.changePasswordWithKubeClient(ctx); err != nil {
			return err
		}
	} else {
		if err := o.changePasswordWithDashClient(ctx); err != nil {
			return err
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("Successfully changed password: %s", o.UserName))

	return nil
}

func (o *changePasswordOption) ValidateUser(ctx context.Context) error {
	var (
		user *dashv1alpha1.User
		err  error
	)
	if o.UseKubeAPI {
		user, err = o.getUserWithKubeClient(ctx, o.UserName)
	} else {
		user, err = o.getUserWithDashClient(ctx, o.UserName)
	}
	if err != nil {
		return err
	}
	if cosmov1alpha1.UserAuthType(user.AuthType) != cosmov1alpha1.UserAuthTypePasswordSecert {
		return fmt.Errorf("password cannot be changed if auth-type is '%s'", user.AuthType)
	}
	return nil
}

func (o *changePasswordOption) getUserWithKubeClient(ctx context.Context, userName string) (*dashv1alpha1.User, error) {
	c := o.KosmoClient
	user, err := c.GetUser(ctx, userName)
	if err != nil {
		return nil, err
	}
	return apiconv.C2D_User(*user), nil
}

func (o *changePasswordOption) getUserWithDashClient(ctx context.Context, userName string) (*dashv1alpha1.User, error) {
	c := o.CosmoDashClient
	res, err := c.UserServiceClient.GetUser(ctx, cli.NewRequestWithToken(&dashv1alpha1.GetUserRequest{UserName: userName}, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("UserServiceClient.GetUser", "res", res)
	return res.Msg.User, nil
}

func (o *changePasswordOption) changePasswordWithKubeClient(ctx context.Context) error {
	c := o.KosmoClient
	if err := c.RegisterPassword(ctx, o.UserName, []byte(o.newPassword)); err != nil {
		return err
	}
	return nil
}

func (o *changePasswordOption) changePasswordWithDashClient(ctx context.Context) error {
	req := &dashv1alpha1.UpdateUserPasswordRequest{
		UserName:        o.UserName,
		CurrentPassword: o.currentPassword,
		NewPassword:     o.newPassword,
	}
	c := o.CosmoDashClient
	res, err := c.UserServiceClient.UpdateUserPassword(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("UserServiceClient.UpdateUserPassword", "res", res)
	return nil
}
