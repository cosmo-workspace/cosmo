package workspace

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/utils/ptr"

	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type GetOption struct {
	*cli.RootOptions

	WorkspaceNames []string
	Filter         []string
	UserName       string
	AllUsers       bool
	OutputFormat   string

	filters []cli.Filter
}

func GetCmd(cmd *cobra.Command, opt *cli.RootOptions) *cobra.Command {
	o := &GetOption{RootOptions: opt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().StringVarP(&o.UserName, "user", "u", "", "user name (defualt: login user)")
	cmd.Flags().StringSliceVar(&o.Filter, "filter", nil, "filter option. available columns are ['NAME', 'TEMPLATE', 'PHASE']. available operators are ['==', '!=']. value format is filepath. e.g. '--filter TEMPLATE==dev-*'")
	cmd.Flags().StringVarP(&o.OutputFormat, "output", "o", "table", "output format. available values are ['table', 'yaml', 'wide']")
	cmd.Flags().BoolVarP(&o.AllUsers, "all-users", "A", false, "get all users workspace")
	return cmd
}

func (o *GetOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	if !o.AllUsers && (o.UseKubeAPI && o.UserName == "") {
		return fmt.Errorf("user name is required")
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
		o.WorkspaceNames = args
	}
	if !o.UseKubeAPI && o.UserName == "" {
		o.UserName = o.CliConfig.User
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

	workspaces := []*dashv1alpha1.Workspace{}
	users := []*dashv1alpha1.User{
		{
			Name: o.UserName,
		},
	}
	if o.AllUsers {
		u, err := o.ListUsers(ctx)
		if err != nil {
			return err
		}
		users = u
	}
	for _, user := range users {
		wss, err := o.ListWorkspaces(ctx, user.Name, !o.AllUsers)
		if err != nil {
			return err
		}
		workspaces = append(workspaces, wss...)
	}

	o.Logr.Debug().Info("Workspaces", "workspaces", workspaces)

	workspaces = o.ApplyFilters(workspaces)

	username := o.UserName
	if o.AllUsers {
		username = ""
	}
	if o.OutputFormat == "yaml" {
		o.OutputYAML(cmd.OutOrStdout(), workspaces)
		return nil
	} else if o.OutputFormat == "wide" {
		OutputWideTable(cmd.OutOrStdout(), username, workspaces)
		return nil
	} else {
		OutputTable(cmd.OutOrStdout(), username, workspaces)
		return nil
	}
}

func (o *GetOption) ListUsers(ctx context.Context) ([]*dashv1alpha1.User, error) {
	if o.UseKubeAPI {
		return o.listUsersByKubeClient(ctx)
	} else {
		return o.listUsersWithDashClient(ctx)
	}
}

func (o *GetOption) ListWorkspaces(ctx context.Context, userName string, includeShared bool) ([]*dashv1alpha1.Workspace, error) {
	if o.UseKubeAPI {
		return o.listWorkspacesByKubeClient(ctx, userName, includeShared)
	} else {
		return o.listWorkspacesWithDashClient(ctx, userName, includeShared)
	}
}

func (o *GetOption) listUsersWithDashClient(ctx context.Context) ([]*dashv1alpha1.User, error) {
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

func (o *GetOption) listWorkspacesWithDashClient(ctx context.Context, userName string, includeShared bool) ([]*dashv1alpha1.Workspace, error) {
	req := &dashv1alpha1.GetWorkspacesRequest{
		UserName:      userName,
		WithRaw:       ptr.To(o.OutputFormat == "yaml"),
		IncludeShared: ptr.To(includeShared),
	}
	c := o.CosmoDashClient
	o.Logr.DebugAll().Info("WorkspaceServiceClient.GetWorkspaces", "req", req)
	res, err := c.WorkspaceServiceClient.GetWorkspaces(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("WorkspaceServiceClient.GetWorkspaces", "res", res)
	return res.Msg.Items, nil
}

func (o *GetOption) ApplyFilters(workspaces []*dashv1alpha1.Workspace) []*dashv1alpha1.Workspace {
	for _, f := range o.filters {
		o.Logr.Debug().Info("applying filter", "key", f.Key, "value", f.Value, "op", f.Operator)

		switch strings.ToUpper(f.Key) {
		case "NAME":
			workspaces = cli.DoFilter(workspaces, func(u *dashv1alpha1.Workspace) []string {
				return []string{u.Name}
			}, f)
		case "TEMPLATE":
			workspaces = cli.DoFilter(workspaces, func(u *dashv1alpha1.Workspace) []string {
				return []string{u.Spec.Template}
			}, f)
		case "PHASE":
			workspaces = cli.DoFilter(workspaces, func(u *dashv1alpha1.Workspace) []string {
				return []string{u.Status.Phase}
			}, f)
		default:
			o.Logr.Info("WARNING: unknown filter key", "key", f.Key)
		}
	}

	// name filter
	for _, wsName := range o.WorkspaceNames {
		workspaces = cli.DoFilter(workspaces, func(u *dashv1alpha1.Workspace) []string {
			return []string{u.Name}
		}, cli.Filter{Operator: cli.OperatorEqual, Value: wsName})
	}

	return workspaces
}

func (o *GetOption) OutputYAML(w io.Writer, objs []*dashv1alpha1.Workspace) {
	docs := make([]string, len(objs))
	for i, t := range objs {
		docs[i] = *t.Raw
	}
	fmt.Fprintln(w, strings.Join(docs, "---\n"))
}

func OutputTable(out io.Writer, username string, workspaces []*dashv1alpha1.Workspace) {
	data := [][]string{}

	for _, v := range workspaces {
		mainURL := v.Status.MainUrl
		if username != "" && v.OwnerName != username {
			mainURL = "[shared workspace. see shared URLs by `cosmoctl ws get-network`]"
		}
		data = append(data, []string{v.OwnerName, v.Name, v.Spec.Template, v.Status.Phase, mainURL})
	}

	cli.OutputTable(out,
		[]string{"USER", "NAME", "TEMPLATE", "PHASE", "MAINURL"},
		data)
}

func OutputWideTable(out io.Writer, username string, workspaces []*dashv1alpha1.Workspace) {
	data := [][]string{}

	for _, v := range workspaces {
		mainURL := v.Status.MainUrl
		if v.OwnerName != username {
			mainURL = `[shared workspace. see shared URLs by "cosmoctl ws get-network"]`
		}
		vars := make([]string, 0, len(v.Spec.Vars))
		for k, vv := range v.Spec.Vars {
			vars = append(vars, fmt.Sprintf("%s=%s", k, vv))
		}
		data = append(data, []string{v.OwnerName, v.Name, v.Spec.Template, strings.Join(vars, ","), v.Status.Phase, mainURL})
	}

	cli.OutputTable(out,
		[]string{"USER", "NAME", "TEMPLATE", "VARS", "PHASE", "MAINURL"},
		data)
}

func (o *GetOption) listWorkspacesByKubeClient(ctx context.Context, userName string, includeShared bool) ([]*dashv1alpha1.Workspace, error) {
	c := o.KosmoClient
	workspaces, err := c.ListWorkspacesByUserName(ctx, userName, func(o *kosmo.ListWorkspacesOptions) {
		o.IncludeShared = includeShared
	})
	if err != nil {
		return nil, err
	}
	return apiconv.C2D_Workspaces(workspaces, apiconv.WithWorkspaceRaw(ptr.To(o.OutputFormat == "yaml"))), nil
}

func (o *GetOption) listUsersByKubeClient(ctx context.Context) ([]*dashv1alpha1.User, error) {
	c := o.KosmoClient
	users, err := c.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	return apiconv.C2D_Users(users, apiconv.WithUserRaw(ptr.To(o.OutputFormat == "yaml"))), nil
}
