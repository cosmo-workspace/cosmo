package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type createOption struct {
	*cmdutil.CliOptions

	UserID        string
	DisplayName   string
	Role          string
	Admin         bool
	Addons        []string
	ClusterAddons []string

	userAddons []wsv1alpha1.UserAddon
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
	cmd.Flags().StringArrayVar(&o.Addons, "addon", nil, "user addons by Template, which created in UserNamespace\nformat is '--addon TEMPLATE_NAME1,KEY:VAL,KEY:VAL --addon TEMPLATE_NAME2,KEY:VAL ...' ")
	cmd.Flags().StringArrayVar(&o.ClusterAddons, "cluster-addon", nil, "user addons by ClusterTemplate\nformat is '--cluster-addon TEMPLATE_NAME1,KEY:VAL,KEY:VAL --cluster-addon TEMPLATE_NAME2,KEY:VAL ...' ")
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

	o.userAddons = make([]wsv1alpha1.UserAddon, 0, len(o.Addons)+len(o.ClusterAddons))
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

func parseUserAddonOptions(rawAddonOptionArray []string, isClusterScope bool) ([]wsv1alpha1.UserAddon, error) {
	// format
	//   TEMPLATE_NAME
	//   TEMPLATE_NAME,KEY1:XXX,KEY2:YYY ZZZ,KEY3:
	r1 := regexp.MustCompile(`^[^: ,]+(,([^: ,]+):([^,]*))*$`)
	r2 := regexp.MustCompile(`^([^: ,]+):([^,]*)$`)

	userAddons := make([]wsv1alpha1.UserAddon, 0, len(rawAddonOptionArray))

	for _, addonParm := range rawAddonOptionArray {
		if !r1.MatchString(addonParm) {
			return nil, fmt.Errorf("invalid addon vars format: %s", addonParm)
		}

		addonSplits := strings.Split(addonParm, ",")

		userAddon := wsv1alpha1.UserAddon{
			Template: cosmov1alpha1.TemplateRef{
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

func (o *createOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	if _, err := o.Client.CreateUser(ctx, o.UserID, o.DisplayName, o.Role, "", o.userAddons); err != nil {
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
