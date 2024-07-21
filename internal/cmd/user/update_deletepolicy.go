package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/utils/ptr"

	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type UpdateDeletePolicyOption struct {
	*cli.RootOptions

	UserName     string
	DeletePolicy string
}

func UpdateDeletePolicyCmd(cmd *cobra.Command, cliOpt *cli.RootOptions) *cobra.Command {
	o := &UpdateDeletePolicyOption{RootOptions: cliOpt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	return cmd
}

func (o *UpdateDeletePolicyOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 2 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *UpdateDeletePolicyOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}

	o.UserName = args[0]
	o.DeletePolicy = args[1]

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return nil
}

func (o *UpdateDeletePolicyOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	o.Logr.Info("updating user delete polocy", "userName", o.UserName, "deletepolicy", o.DeletePolicy)

	var (
		user *dashv1alpha1.User
		err  error
	)
	if o.UseKubeAPI {
		user, err = o.UpdateUserDeletePolicyWithKubeClient(ctx)
	} else {
		user, err = o.UpdateUserDeletePolicyWithDashClient(ctx)
	}
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("Successfully updated user %s", o.UserName))
	OutputWideTable(cmd.OutOrStdout(), []*dashv1alpha1.User{user})

	return nil
}

func (o *UpdateDeletePolicyOption) UpdateUserDeletePolicyWithDashClient(ctx context.Context) (*dashv1alpha1.User, error) {
	delPolicy := apiconv.C2D_DeletePolicy(o.DeletePolicy)
	if delPolicy == nil {
		delPolicy = ptr.To(dashv1alpha1.DeletePolicy_delete)
	}

	req := &dashv1alpha1.UpdateUserDeletePolicyRequest{
		UserName:     o.UserName,
		DeletePolicy: *delPolicy,
	}
	o.Logr.DebugAll().Info("UserServiceClient.UpdateUserDeletePolicy", "req", req)
	c := o.CosmoDashClient
	res, err := c.UserServiceClient.UpdateUserDeletePolicy(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("UserServiceClient.UpdateUserDeletePolicy", "res", res)

	return res.Msg.User, nil
}

func (o *UpdateDeletePolicyOption) UpdateUserDeletePolicyWithKubeClient(ctx context.Context) (*dashv1alpha1.User, error) {
	c := o.KosmoClient
	opts := kosmo.UpdateUserOpts{
		DeletePolicy: &o.DeletePolicy,
	}
	o.Logr.DebugAll().Info("UpdateUser", "opts", opts)
	user, err := c.UpdateUser(ctx, o.UserName, opts)
	if err != nil {
		return nil, err
	}
	d := apiconv.C2D_User(*user)

	return d, nil
}
