package template

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/printers"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/wscfg"
)

type getOption struct {
	*cmdutil.CliOptions
	TemplateNames []string
	TypeWorkspace bool

	tmpltype string
}

func getCmd(cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &getOption{CliOptions: cliOpt}
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get templates",
		Long: `Get Templates

Basically it is similar to "kubectl get template"

For type workspace template, use with --workspace flag to see more information. 
`,
		PersistentPreRunE: o.PreRunE,
		RunE:              cmdutil.RunEHandler(o.RunE),
	}
	cmd.PersistentFlags().BoolVar(&o.TypeWorkspace, "workspace", false, "show type workspace template")
	return cmd
}

func (o *getOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *getOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if o.TypeWorkspace {
		o.tmpltype = wsv1alpha1.TemplateTypeWorkspace
	}
	return nil
}

func (o *getOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}
	if len(args) > 0 {
		o.TemplateNames = args
	}
	return nil
}

func (o *getOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()

	var tmpls []cosmov1alpha1.TemplateObject

	o.Logr.Debug().Info("options", "templateNames", o.TemplateNames)

	if o.tmpltype != "" {
		ts, err := kubeutil.ListTemplateObjectsByType(ctx, o.Client, []string{wsv1alpha1.TemplateTypeWorkspace})
		if err != nil {
			return err
		}
		tmpls = ts
	} else {
		ts, err := kubeutil.ListTemplateObjects(ctx, o.Client)
		if err != nil {
			return err
		}
		tmpls = ts
	}
	o.Logr.DebugAll().Info("ListTemplates", "tmplList", tmpls)

	w := printers.GetNewTabWriter(o.Out)
	defer w.Flush()

	switch o.tmpltype {
	case wsv1alpha1.TemplateTypeWorkspace:

		columnNames := []string{"NAME", "REQUIRED-VARS", "DEPLOYMENT/SERVICE/INGRESS", "URLBASE"}
		fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))

		for _, v := range tmpls {
			cfg, err := wscfg.ConfigFromTemplateAnnotations(v.(*cosmov1alpha1.Template))
			if err != nil {
				o.Logr.Error(err, "failed to get workspace config", "template", v.GetName())
				continue
			}

			vars := make([]string, 0, len(v.GetSpec().RequiredVars))
			for _, t := range v.GetSpec().RequiredVars {
				vars = append(vars, t.Var)
			}
			rawTmplVars := strings.Join(vars, ",")

			resources := fmt.Sprintf("%s/%s/%s", cfg.DeploymentName, cfg.ServiceName, cfg.IngressName)
			rowdata := []string{v.GetName(), rawTmplVars, resources, cfg.URLBase}
			fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
		}

	default:
		columnNames := []string{"NAME", "REQUIRED-VARS", "TYPE", "IsClusterScope"}
		fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))

		for _, v := range tmpls {
			vars := make([]string, 0, len(v.GetSpec().RequiredVars))
			for _, t := range v.GetSpec().RequiredVars {
				vars = append(vars, t.Var)
			}
			rawTmplVars := strings.Join(vars, ",")

			tmplType := v.GetLabels()[cosmov1alpha1.TemplateLabelKeyType]
			rowdata := []string{v.GetName(), rawTmplVars, tmplType, strconv.FormatBool(v.GetScope() == meta.RESTScopeRoot)}
			fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
		}
	}

	return nil
}
