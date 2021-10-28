package user

import (
	"context"
	"errors"
	"fmt"
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

	UserID      string
	DisplayName string
	Role        string
	Admin       bool
	Addons      string
	AddonVars   string

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
	cmd.Flags().StringVar(&o.Addons, "addons", "", "user addons, which created after UserNamespace created. format is '--addons ADDON_TEMPLATE_NAME1,ADDON_TEMPLATE_NAME2 ...' ")
	cmd.Flags().StringVar(&o.Addons, "addon-vars", "", "user addons template vars. format is '--addons-vars Addon=ADDON_TEMPLATE_NAME1,KEY=VAL,KEY=VAL,Addon=ADDON_TEMPLATE_NAME2,KEY=VAL ...' ")
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
	user.SetName(o.UserID)
	user.Spec = wsv1alpha1.UserSpec{}

	if o.DisplayName != "" {
		user.Spec.DisplayName = o.DisplayName
	} else {
		user.Spec.DisplayName = o.UserID
	}

	if o.Admin {
		o.Role = wsv1alpha1.UserAdminRole.String()
	}
	user.Spec.Role = wsv1alpha1.UserRole(o.Role)

	if o.Addons != "" {
		// format is "ADDON_TEMPLATE_NAME1,ADDON_TEMPLATE_NAME2", split by ,
		col := strings.Split(o.Addons, ",")
		addons := make([]wsv1alpha1.UserAddon, len(col))
		for i, a := range col {
			addons[i] = wsv1alpha1.UserAddon{
				Template: cosmov1alpha1.TemplateRef{
					Name: a,
				},
			}
		}
		o.user.Spec.Addons = addons
	}

	if o.AddonVars != "" {
		// format is "Addon=ADDON_TEMPLATE_NAME1,KEY=VAL,KEY=VAL,Addon=ADDON_TEMPLATE_NAME2,KEY=VAL", split by Addon=
		col := strings.Split(o.AddonVars, "Addon=")
		for _, c := range col {
			if c == "" {
				continue
			}

			// format is "ADDON_TEMPLATE_NAME1,KEY=VAL,KEY=VAL"
			varsCol := strings.Split(c, ",")
			if len(varsCol) > 2 {
				return fmt.Errorf("invalid addon vars format: %s", c)
			}
			varsAddonName := varsCol[0]

			for i, addon := range o.user.Spec.Addons {
				if addon.Template.Name == varsAddonName {

					instVars := make(map[string]string)
					for _, kv := range varsCol[1:] {
						if kv == "" {
							continue
						}
						kvcol := strings.Split(kv, "=")
						if len(kvcol) != 2 {
							return fmt.Errorf("invalid addon vars format: %s", c)
						}
						instVars[kvcol[0]] = kvcol[1]
					}
					o.user.Spec.Addons[i].Vars = instVars
				}
			}
		}
	}

	o.user = &user
	return nil
}

func (o *createOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	o.Logr.DebugAll().Info("creating user", "user", o.user)

	if err := c.Create(ctx, o.user); err != nil {
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
