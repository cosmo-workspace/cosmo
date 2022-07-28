package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
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
	Addons      []string
	UserAddons  []wsv1alpha1.UserAddon

	//user *wsv1alpha1.User
}

func createCmd(cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &createOption{CliOptions: cliOpt}
	cmd := &cobra.Command{
		Use:               "create USER_ID --role cosmo-admin",
		Short:             "Create user",
		PersistentPreRunE: o.PreRunE,
		RunE:              cmdutil.RunEHandler(o.RunE),
	}
	cmd.Flags().StringVar(&o.DisplayName, "name", "", "user display name (default: same as USER_ID)")
	cmd.Flags().StringVar(&o.Role, "role", "", "user role")
	cmd.Flags().BoolVar(&o.Admin, "admin", false, "user admin role")
	cmd.Flags().StringArrayVar(&o.Addons, "addon", nil, "user addons, which created after UserNamespace created.\nformat is '--addon TEMPLATE_NAME1,KEY:VAL,KEY:VAL --addon TEMPLATE_NAME2,KEY:VAL ...' ")
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

	if o.Admin {
		o.Role = wsv1alpha1.UserAdminRole.String()
	}

	if len(o.Addons) > 0 {
		// format
		//   ADDON_TEMPLATE_NAME1
		//   ADDON_TEMPLATE_NAME1,KEY1:XXX,KEY2:YYY ZZZ,KEY3:
		r1 := regexp.MustCompile(`^[^: ,]+(,([^: ,]+):([^,]*))*$`)
		r2 := regexp.MustCompile(`^([^: ,]+):([^,]*)$`)

		o.UserAddons = make([]wsv1alpha1.UserAddon, 0, len(o.Addons))

		for _, addonParm := range o.Addons {
			if !r1.MatchString(addonParm) {
				return fmt.Errorf("invalid addon vars format: %s", addonParm)
			}

			addonSplits := strings.Split(addonParm, ",")

			var userAddon wsv1alpha1.UserAddon
			userAddon.Template.Name = addonSplits[0]
			userAddon.Vars = make(map[string]string, len(addonSplits)-1)
			for _, k_v := range addonSplits[1:] {
				kv := r2.FindStringSubmatch(k_v)
				userAddon.Vars[kv[1]] = kv[2]
			}
			o.UserAddons = append(o.UserAddons, userAddon)
		}
	}

	return nil
}

func (o *createOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	if _, err := o.Client.CreateUser(ctx, o.UserID, o.DisplayName, o.Role, "", o.UserAddons); err != nil {
		return err
	}

	defaultPassword, err := o.Client.GetDefaultPasswordAwait(ctx, o.UserID)
	if err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully created user %s\n", o.UserID)
	fmt.Fprintln(o.Out, "Default password:", *defaultPassword)
	return nil
}
