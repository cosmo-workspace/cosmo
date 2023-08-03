package user

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/printers"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type GetOption struct {
	*cmdutil.CliOptions

	UserNames []string
	Filter    []string

	roleFilter  []string
	addonFilter []string
}

func GetCmd(cmd *cobra.Command, cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &GetOption{CliOptions: cliOpt}

	cmd.PersistentPreRunE = o.PreRunE
	cmd.RunE = cmdutil.RunEHandler(o.RunE)
	cmd.Flags().StringSliceVar(&o.Filter, "filter", nil, "filter option. 'role' and 'addon' are available for now. e.g. 'role=x', 'addon=y'")
	return cmd
}

func (o *GetOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *GetOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Validate(cmd, args); err != nil {
		return err
	}
	return nil
}

func (o *GetOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}
	if len(args) > 0 {
		o.UserNames = args
	}
	if len(o.Filter) > 0 {
		for _, f := range o.Filter {
			s := strings.Split(f, "=")
			if len(s) != 2 {
				return fmt.Errorf("invalid filter expression: %s", f)
			}
			switch s[0] {
			case "addon":
				o.addonFilter = append(o.addonFilter, s[1])
			case "role":
				o.roleFilter = append(o.roleFilter, s[1])
			default:
				o.Logr.Info("invalid filter expression", "filter", f)
				return fmt.Errorf("invalid filter expression: %s", f)
			}
		}
	}
	return nil
}

func (o *GetOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	users, err := c.ListUsers(ctx)
	if err != nil {
		return err
	}
	o.Logr.DebugAll().Info("ListUsers", "users", users)
	o.Logr.Debug().Info("filter", "role", o.roleFilter, "addon", o.addonFilter)

	if len(o.roleFilter) > 0 {
		// And loop
		for _, selected := range o.roleFilter {
			ts := make([]cosmov1alpha1.User, 0)
			for _, t := range users {
			RoleFilterLoop:
				for _, v := range t.Spec.Roles {
					if matched, err := filepath.Match(selected, v.Name); err == nil && matched {
						ts = append(ts, t)
						break RoleFilterLoop
					}
				}
			}
			users = ts
		}
	}
	if len(o.addonFilter) > 0 {
		// And loop
		for _, selected := range o.addonFilter {
			ts := make([]cosmov1alpha1.User, 0, len(o.UserNames))
			for _, t := range users {
			AddonsLoop:
				for _, v := range t.Spec.Addons {
					if matched, err := filepath.Match(selected, v.Template.Name); err == nil && matched {
						ts = append(ts, t)
						break AddonsLoop
					}
				}
			}
			users = ts
		}
	}

	if len(o.UserNames) > 0 {
		ts := make([]cosmov1alpha1.User, 0, len(o.UserNames))
	UserLoop:
		// Or loop
		for _, t := range users {
			for _, selected := range o.UserNames {
				if selected == t.GetName() {
					ts = append(ts, t)
					continue UserLoop
				}
			}
		}
		users = ts
	}

	w := printers.GetNewTabWriter(o.Out)
	defer w.Flush()

	columnNames := []string{"NAME", "ROLES", "AUTHTYPE", "NAMESPACE", "PHASE", "ADDONS"}
	fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))
	for _, v := range users {
		role := make([]string, 0, len(v.Spec.Roles))
		for _, v := range v.Spec.Roles {
			role = append(role, v.Name)
		}
		addons := make([]string, 0, len(v.Spec.Addons))
		for _, v := range v.Spec.Addons {
			addons = append(addons, v.Template.Name)
		}
		rowdata := []string{v.Name, strings.Join(role, ","), v.Spec.AuthType.String(), v.Status.Namespace.Name, string(v.Status.Phase), strings.Join(addons, ",")}
		fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
	}

	return nil
}
