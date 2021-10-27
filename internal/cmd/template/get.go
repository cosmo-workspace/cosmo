package template

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/printers"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type getOption struct {
	*cmdutil.CliOptions
	TemplateName  string
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
		RunE:              o.RunE,
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
		o.TemplateName = args[0]
	}
	return nil
}

func (o *getOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	c := o.Client

	var tmpls []cosmov1alpha1.Template

	o.Logr.Debug().Info("options", "templateName", o.TemplateName)

	if o.TemplateName != "" {
		tmpl, err := c.GetTemplate(ctx, o.TemplateName)
		if err != nil {
			return err
		}
		tmpls = []cosmov1alpha1.Template{*tmpl}
		o.Logr.DebugAll().Info("GetTemplate", "tmpls", tmpls)

	} else {
		switch o.tmpltype {
		case wsv1alpha1.TemplateTypeWorkspace:
			ts, err := c.ListTemplatesByType(ctx, []string{wsv1alpha1.TemplateTypeWorkspace})
			if err != nil {
				return err
			}
			tmpls = ts

		default:
			ts, err := c.ListTemplates(ctx)
			if err != nil {
				return err
			}
			tmpls = ts
		}
		o.Logr.DebugAll().Info("ListTemplates", "tmplList", tmpls)
	}

	w := printers.GetNewTabWriter(o.Out)
	defer w.Flush()

	switch o.tmpltype {
	case wsv1alpha1.TemplateTypeWorkspace:

		columnNames := []string{"NAME", "REQUIRED-VARS", "DEPLOYMENT/SERVICE/INGRESS", "URLBASE"}
		fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))

		for _, v := range tmpls {
			cfg, err := wsv1alpha1.ConfigFromTemplateAnnotations(&v)
			if err != nil {
				o.Logr.Error(err, "failed to get workspace config", "template", v.GetName())
				continue
			}

			vars := make([]string, 0, len(v.Spec.RequiredVars))
			for _, t := range v.Spec.RequiredVars {
				vars = append(vars, t.Var)
			}
			rawTmplVars := strings.Join(vars, ",")

			resources := fmt.Sprintf("%s/%s/%s", cfg.DeploymentName, cfg.ServiceName, cfg.IngressName)
			rowdata := []string{v.Name, rawTmplVars, resources, cfg.URLBase}
			fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
		}

	default:
		columnNames := []string{"NAME", "REQUIRED-VARS", "TYPE"}
		fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))

		for _, v := range tmpls {
			vars := make([]string, 0, len(v.Spec.RequiredVars))
			for _, t := range v.Spec.RequiredVars {
				vars = append(vars, t.Var)
			}
			rawTmplVars := strings.Join(vars, ",")

			tmplType := v.Labels[cosmov1alpha1.LabelKeyTemplateType]
			rowdata := []string{v.Name, rawTmplVars, tmplType}
			fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
		}
	}

	return nil
}
