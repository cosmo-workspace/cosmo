package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type CreateOption struct {
	*cmdutil.CliOptions

	UserName      string
	DisplayName   string
	Roles         []string
	AuthType      string
	Admin         bool
	Addons        []string
	ClusterAddons []string

	userAddons []cosmov1alpha1.UserAddon
}

func CreateCmd(cmd *cobra.Command, cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &CreateOption{CliOptions: cliOpt}
	cmd.PersistentPreRunE = o.PreRunE
	cmd.RunE = cmdutil.RunEHandler(o.RunE)
	cmd.Flags().StringVar(&o.DisplayName, "name", "", "user display name (default: same as USER_NAME)")
	cmd.Flags().StringSliceVar(&o.Roles, "role", nil, "user roles")
	cmd.Flags().StringVar(&o.AuthType, "auth-type", cosmov1alpha1.UserAuthTypePasswordSecert.String(), "user auth type 'password-secret'(default),'ldap'")
	cmd.Flags().BoolVar(&o.Admin, "admin", false, "user admin role")
	cmd.Flags().StringArrayVar(&o.Addons, "addon", nil, "user addons by Template, which created in UserNamespace\nformat is '--addon TEMPLATE_NAME1,KEY:VAL,KEY:VAL --addon TEMPLATE_NAME2,KEY:VAL ...' ")
	cmd.Flags().StringArrayVar(&o.ClusterAddons, "cluster-addon", nil, "user addons by ClusterTemplate\nformat is '--cluster-addon TEMPLATE_NAME1,KEY:VAL,KEY:VAL --cluster-addon TEMPLATE_NAME2,KEY:VAL ...' ")
	return cmd
}

func (o *CreateOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *CreateOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Validate(cmd, args); err != nil {
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
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}

	o.UserName = args[0]

	if o.Admin {
		o.Roles = []string{cosmov1alpha1.PrivilegedRoleName}
	}

	o.userAddons = make([]cosmov1alpha1.UserAddon, 0, len(o.Addons)+len(o.ClusterAddons))
	if len(o.Addons) > 0 {
		userAddons, err := parseUserAddonOptions(o.Addons, false)
		if err != nil {
			return err
		}
		o.userAddons = append(o.userAddons, userAddons...)
	}
	if len(o.ClusterAddons) > 0 {
		userAddons, err := parseUserAddonOptions(o.ClusterAddons, true)
		if err != nil {
			return err
		}
		o.userAddons = append(o.userAddons, userAddons...)
	}

	return nil
}

func parseUserAddonOptions(rawAddonOptionArray []string, isClusterScope bool) ([]cosmov1alpha1.UserAddon, error) {
	// format
	//   TEMPLATE_NAME
	//   TEMPLATE_NAME,KEY1:XXX,KEY2:YYY ZZZ,KEY3:
	r1 := regexp.MustCompile(`^[^: ,]+(,([^: ,]+):([^,]*))*$`)
	r2 := regexp.MustCompile(`^([^: ,]+):([^,]*)$`)

	userAddons := make([]cosmov1alpha1.UserAddon, 0, len(rawAddonOptionArray))

	for _, addonParm := range rawAddonOptionArray {
		if !r1.MatchString(addonParm) {
			return nil, fmt.Errorf("invalid addon vars format: %s", addonParm)
		}

		addonSplits := strings.Split(addonParm, ",")

		userAddon := cosmov1alpha1.UserAddon{
			Template: cosmov1alpha1.UserAddonTemplateRef{
				Name:          addonSplits[0],
				ClusterScoped: isClusterScope,
			},
			Vars: make(map[string]string, len(addonSplits)-1),
		}

		for _, k_v := range addonSplits[1:] {
			kv := r2.FindStringSubmatch(k_v)
			userAddon.Vars[kv[1]] = kv[2]
		}
		userAddons = append(userAddons, userAddon)
	}
	return userAddons, nil
}

func (o *CreateOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	if _, err := o.Client.CreateUser(ctx, o.UserName, o.DisplayName, o.Roles, o.AuthType, o.userAddons); err != nil {
		return err
	}

	if o.AuthType == cosmov1alpha1.UserAuthTypePasswordSecert.String() {
		defaultPassword, err := o.Client.GetDefaultPasswordAwait(ctx, o.UserName)
		if err != nil {
			return err
		}
		cmdutil.PrintfColorInfo(o.Out, "Successfully created user %s\n", o.UserName)
		fmt.Fprintln(o.Out, "Default password:", *defaultPassword)
	} else {
		cmdutil.PrintfColorInfo(o.Out, "Successfully created user %s\n", o.UserName)
	}
	return nil
}
