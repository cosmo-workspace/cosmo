package user

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	connect_go "github.com/bufbuild/connect-go"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"k8s.io/utils/ptr"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type GetOption struct {
	*cli.RootOptions

	UserNames    []string
	Filter       []string
	OutputFormat string

	filters []cli.Filter
}

func GetCmd(cmd *cobra.Command, opt *cli.RootOptions) *cobra.Command {
	o := &GetOption{RootOptions: opt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().StringSliceVar(&o.Filter, "filter", nil, "filter option. available columns are ['NAME', 'ROLE', 'ADDON', 'AUTHTYPE', 'PHASE']. available operators are ['==', '!=']. value format is filepath. e.g. '--filter ROLE==*-dev --filter ROLE!=team-a'")
	cmd.Flags().StringVarP(&o.OutputFormat, "output", "o", "table", "output format. available values are ['table', 'yaml', 'wide']")
	return cmd
}

func (o *GetOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	switch o.OutputFormat {
	case "table", "yaml", "wide":
	default:
		return fmt.Errorf("invalid output format: %s", o.OutputFormat)
	}
	return nil
}

func (o *GetOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}
	if len(args) > 0 {
		o.UserNames = args
	}
	if len(o.Filter) > 0 {
		f, err := cli.ParseFilters(o.Filter)
		if err != nil {
			return err
		}
		o.filters = f
	}
	for _, f := range o.filters {
		o.Logr.Debug().Info("filter", "key", f.Key, "value", f.Value, "op", f.Operator)
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return nil
}

func (o *GetOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*30)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	var users []*dashv1alpha1.User
	var err error
	if o.UseKubeAPI {
		users, err = o.ListUsersByKubeClient(ctx)
		if err != nil {
			return err
		}
	} else {
		users, err = o.ListUsersWithDashClient(ctx)
		if err != nil {
			if connect_go.CodeOf(err) == connect_go.CodePermissionDenied {

				if len(o.UserNames) == 0 {
					fmt.Fprintln(cmd.ErrOrStderr(), color.YellowString("WARNING: Without Admin roles, you can get only login user"))
				} else {
					for _, v := range o.UserNames {
						if v != o.CliConfig.User {
							return fmt.Errorf("permission denied: failed to get user: %s", v)
						}
					}
				}
				me, err := o.GetUserWithDashClient(ctx, o.CliConfig.User)
				if err != nil {
					return err
				}
				users = []*dashv1alpha1.User{me}
			} else {
				return err
			}
		}
	}
	o.Logr.Debug().Info("Users", "users", users)

	users = o.ApplyFilters(users)

	if o.OutputFormat == "yaml" {
		o.OutputYAML(cmd.OutOrStdout(), users)
		return nil
	} else if o.OutputFormat == "wide" {
		OutputWideTable(cmd.OutOrStdout(), users)
		return nil
	} else {
		OutputTable(cmd.OutOrStdout(), users)
		return nil
	}
}

func (o *GetOption) ListUsersWithDashClient(ctx context.Context) ([]*dashv1alpha1.User, error) {
	c := o.CosmoDashClient
	res, err := c.UserServiceClient.GetUsers(ctx, cli.NewRequestWithToken(&dashv1alpha1.GetUsersRequest{
		WithRaw: ptr.To(o.OutputFormat == "yaml"),
	}, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("UserServiceClient.GetUsers", "res", res)
	return res.Msg.Items, nil
}

func (o *GetOption) GetUserWithDashClient(ctx context.Context, userName string) (*dashv1alpha1.User, error) {
	c := o.CosmoDashClient
	res, err := c.UserServiceClient.GetUser(ctx, cli.NewRequestWithToken(&dashv1alpha1.GetUserRequest{
		UserName: userName,
		WithRaw:  ptr.To(o.OutputFormat == "yaml"),
	}, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("UserServiceClient.GetUser", "res", res)
	return res.Msg.User, nil
}

func (o *GetOption) ApplyFilters(users []*dashv1alpha1.User) []*dashv1alpha1.User {
	for _, f := range o.filters {
		o.Logr.Debug().Info("applying filter", "key", f.Key, "value", f.Value, "op", f.Operator)

		switch strings.ToUpper(f.Key) {
		case "NAME":
			users = cli.DoFilter(users, func(u *dashv1alpha1.User) []string {
				return []string{u.Name}
			}, f)
		case "ROLE", "ROLES":
			users = cli.DoFilter(users, func(u *dashv1alpha1.User) []string {
				arr := make([]string, 0, len(u.Roles))
				arr = append(arr, u.Roles...)
				return arr
			}, f)
		case "ADDON", "ADDONS":
			users = cli.DoFilter(users, func(u *dashv1alpha1.User) []string {
				arr := make([]string, 0, len(u.Addons))
				for _, a := range u.Addons {
					arr = append(arr, a.Template)
				}
				return arr
			}, f)
		case "AUTHTYPE":
			users = cli.DoFilter(users, func(u *dashv1alpha1.User) []string {
				return []string{u.AuthType}
			}, f)
		case "PHASE":
			users = cli.DoFilter(users, func(u *dashv1alpha1.User) []string {
				return []string{u.Status}
			}, f)
		default:
			o.Logr.Info("WARNING: unknown filter key", "key", f.Key)
		}
	}

	// name filter
	for _, userName := range o.UserNames {
		users = cli.DoFilter(users, func(u *dashv1alpha1.User) []string {
			return []string{u.Name}
		}, cli.Filter{Operator: cli.OperatorEqual, Value: userName})
	}

	return users
}

func (o *GetOption) OutputYAML(w io.Writer, objs []*dashv1alpha1.User) {
	docs := make([]string, len(objs))
	for i, t := range objs {
		docs[i] = *t.Raw
	}
	fmt.Fprintln(w, strings.Join(docs, "---\n"))
}

func printAddons(addons []*dashv1alpha1.UserAddon) string {
	arr := make([]string, len(addons))
	for i, v := range addons {
		arr[i] = v.Template
	}
	return strings.Join(arr, ",")
}

func printAddonWithVars(addons []*dashv1alpha1.UserAddon) string {
	arr := make([]string, len(addons))
	for i, v := range apiconv.D2S_UserAddons(addons) {
		arr[i] = v
	}
	return strings.Join(arr, " ")
}

func OutputTable(out io.Writer, users []*dashv1alpha1.User) {
	data := [][]string{}

	for _, v := range users {
		data = append(data, []string{v.Name, strings.Join(v.Roles, ","), v.AuthType, cosmov1alpha1.UserNamespace(v.Name), v.Status, printAddons(v.Addons)})
	}

	cli.OutputTable(out,
		[]string{"NAME", "ROLES", "AUTHTYPE", "NAMESPACE", "PHASE", "ADDONS"},
		data)
}

func OutputWideTable(out io.Writer, users []*dashv1alpha1.User) {
	data := [][]string{}

	for _, v := range users {
		data = append(data, []string{v.Name, v.DisplayName, strings.Join(v.Roles, ","), v.AuthType, cosmov1alpha1.UserNamespace(v.Name), v.Status, printAddonWithVars(v.Addons)})
	}

	cli.OutputTable(out,
		[]string{"NAME", "DISPLAYNAME", "ROLES", "AUTHTYPE", "NAMESPACE", "PHASE", "ADDONS"},
		data)
}

func (o *GetOption) ListUsersByKubeClient(ctx context.Context) ([]*dashv1alpha1.User, error) {
	c := o.KosmoClient
	users, err := c.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	return apiconv.C2D_Users(users, apiconv.WithUserRaw(ptr.To(o.OutputFormat == "yaml"))), nil
}
