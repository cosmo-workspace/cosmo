package user

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/utils/ptr"

	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type GetAddonsOption struct {
	*cli.RootOptions
	AddonNames   []string
	Filter       []string
	OutputFormat string

	filters []cli.Filter
}

func GetAddonsCmd(cmd *cobra.Command, opt *cli.RootOptions) *cobra.Command {
	o := &GetAddonsOption{RootOptions: opt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().StringSliceVar(&o.Filter, "filter", nil, "filter option. available columns are ['NAME', 'USERROLE', 'REQUIRED_USERADDON']. available operators are ['==', '!=']. value format is filepath. e.g. '--filter USERROLE==*-dev --filter USERROLE!=team-a'")
	cmd.Flags().StringVarP(&o.OutputFormat, "output", "o", "table", "output format. available values are ['table', 'yaml']")
	return cmd
}

func (o *GetAddonsOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	switch o.OutputFormat {
	case "table", "yaml":
	default:
		return fmt.Errorf("invalid output format: %s", o.OutputFormat)
	}
	return nil
}

func (o *GetAddonsOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}
	if len(args) > 0 {
		o.AddonNames = args
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

func (o *GetAddonsOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*30)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	var (
		tmpls []*dashv1alpha1.Template
		err   error
	)
	if o.UseKubeAPI {
		tmpls, err = o.ListUserAddonsByKubeClient(ctx, o.OutputFormat == "yaml")
	} else {
		tmpls, err = o.ListUserAddonsWithDashClient(ctx, o.OutputFormat == "yaml")
	}
	if err != nil {
		return err
	}
	o.Logr.Debug().Info("UserAddon templates", "templates", tmpls)

	tmpls = o.ApplyFilters(tmpls)

	if o.OutputFormat == "yaml" {
		o.OutputYAML(cmd.OutOrStdout(), tmpls)
		return nil
	} else {
		o.OutputTable(cmd.OutOrStdout(), tmpls)
		return nil
	}
}

func (o *GetAddonsOption) ListUserAddonsWithDashClient(ctx context.Context, withRaw bool) ([]*dashv1alpha1.Template, error) {
	req := &dashv1alpha1.GetUserAddonTemplatesRequest{
		UseRoleFilter: ptr.To(false),
		WithRaw:       ptr.To(withRaw),
	}
	c := o.CosmoDashClient
	res, err := c.TemplateServiceClient.GetUserAddonTemplates(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to connect dashboard server: %w", err)
	}
	o.Logr.DebugAll().Info("TemplateServiceClient.GetUserAddonTemplates", "res", res)
	return res.Msg.Items, nil
}

func (o *GetAddonsOption) ApplyFilters(tmpls []*dashv1alpha1.Template) []*dashv1alpha1.Template {
	for _, f := range o.filters {
		o.Logr.Debug().Info("applying filter", "key", f.Key, "value", f.Value, "op", f.Operator)

		switch strings.ToUpper(f.Key) {
		case "NAME":
			tmpls = cli.DoFilter(tmpls, func(u *dashv1alpha1.Template) []string {
				return []string{u.Name}
			}, f)
		case "USERROLE", "USERROLES", "REQUIRED_USERROLES":
			tmpls = cli.DoFilter(tmpls, func(u *dashv1alpha1.Template) []string {
				arr := make([]string, 0, len(u.Userroles))
				arr = append(arr, u.Userroles...)
				return arr
			}, f)
		case "REQUIRED_USERADDONS":
			tmpls = cli.DoFilter(tmpls, func(u *dashv1alpha1.Template) []string {
				arr := make([]string, 0, len(u.RequiredUseraddons))
				arr = append(arr, u.RequiredUseraddons...)
				return arr
			}, f)
		default:
			o.Logr.Info("WARNING: unknown filter key", "key", f.Key)
		}
	}

	if len(o.AddonNames) > 0 {
		ts := make([]*dashv1alpha1.Template, 0, len(o.AddonNames))
	UserLoop:
		// Or loop
		for _, t := range tmpls {
			for _, selected := range o.AddonNames {
				if selected == t.GetName() {
					ts = append(ts, t)
					continue UserLoop
				}
			}
		}
		tmpls = ts
	}
	return tmpls
}

func (o *GetAddonsOption) OutputYAML(w io.Writer, tmpls []*dashv1alpha1.Template) {
	docs := make([]string, len(tmpls))
	for i, t := range tmpls {
		docs[i] = *t.Raw
	}
	fmt.Fprintln(w, strings.Join(docs, "---\n"))
}

func (o *GetAddonsOption) OutputTable(w io.Writer, tmpls []*dashv1alpha1.Template) {
	data := [][]string{}

	for _, v := range tmpls {
		rawRequiredAddons := strings.Join(v.RequiredUseraddons, ",")
		rawUserroles := strings.Join(v.Userroles, ",")

		var isDefaultUserAddon bool
		if v.IsDefaultUserAddon != nil {
			isDefaultUserAddon = *v.IsDefaultUserAddon
		}
		data = append(data, []string{v.GetName(), strconv.FormatBool(isDefaultUserAddon), requiredVars(v.RequiredVars), rawUserroles, rawRequiredAddons})
	}

	cli.OutputTable(w,
		[]string{"NAME", "DEFAULT", "REQUIRED_VARS(default)", "USERROLE", "REQUIRED_USERADDON"},
		data)
}

func requiredVars(vs []*dashv1alpha1.TemplateRequiredVars) string {
	var s []string
	for _, v := range vs {
		data := v.VarName
		if v.DefaultValue != "" {
			data += fmt.Sprintf("(%s)", v.DefaultValue)
		}
		s = append(s, data)
	}
	return strings.Join(s, ",")
}

func (o *GetAddonsOption) ListUserAddonsByKubeClient(ctx context.Context, withRaw bool) ([]*dashv1alpha1.Template, error) {
	c := o.KosmoClient
	tmpls, err := c.ListUserAddonTemplates(ctx)
	if err != nil {
		return nil, err
	}
	return apiconv.C2D_Templates(tmpls, apiconv.WithTemplateRaw(&withRaw)), nil
}
