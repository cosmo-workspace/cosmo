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

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/workspace"
)

type GetOption struct {
	*cmdutil.CliOptions
	TemplateNames []string
	TypeWorkspace bool

	tmpltype string
}

func GetCmd(cmd *cobra.Command, cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &GetOption{CliOptions: cliOpt}
	cmd.PersistentPreRunE = o.PreRunE
	cmd.RunE = cmdutil.RunEHandler(o.RunE)
	cmd.PersistentFlags().BoolVar(&o.TypeWorkspace, "workspace", false, "show type workspace template")
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
	if o.TypeWorkspace {
		o.tmpltype = cosmov1alpha1.TemplateLabelEnumTypeWorkspace
	}
	return nil
}

func (o *GetOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}
	if len(args) > 0 {
		o.TemplateNames = args
	}
	return nil
}

func (o *GetOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()

	var tmpls []cosmov1alpha1.TemplateObject

	o.Logr.Debug().Info("options", "templateNames", o.TemplateNames)

	if o.tmpltype != "" {
		ts, err := kubeutil.ListTemplateObjectsByType(ctx, o.Client, []string{cosmov1alpha1.TemplateLabelEnumTypeWorkspace})
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

	if len(o.TemplateNames) > 0 {
		ts := make([]cosmov1alpha1.TemplateObject, 0, len(o.TemplateNames))
		for _, selected := range o.TemplateNames {
			for _, t := range tmpls {
				if selected == t.GetName() {
					ts = append(ts, t)
				}
			}
		}
		tmpls = ts
	}

	w := printers.GetNewTabWriter(o.Out)
	defer w.Flush()

	switch o.tmpltype {
	case cosmov1alpha1.TemplateLabelEnumTypeWorkspace:

		columnNames := []string{"NAME", "REQUIRED-VARS", "DEPLOYMENT/SERVICE"}
		fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))

		for _, v := range tmpls {
			cfg, err := workspace.ConfigFromTemplateAnnotations(v.(*cosmov1alpha1.Template))
			if err != nil {
				o.Logr.Error(err, "failed to get workspace config", "template", v.GetName())
				continue
			}

			vars := make([]string, 0, len(v.GetSpec().RequiredVars))
			for _, t := range v.GetSpec().RequiredVars {
				vars = append(vars, t.Var)
			}
			rawTmplVars := strings.Join(vars, ",")

			resources := fmt.Sprintf("%s/%s", cfg.DeploymentName, cfg.ServiceName)
			rowdata := []string{v.GetName(), rawTmplVars, resources}
			fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
		}

	default:
		columnNames := []string{"NAME", "REQUIREDVARS", "TYPE", "ISCLUSTERSCOPE", "FORROLE", "FORBIDDENROLE"}
		fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))

		for _, v := range tmpls {
			vars := make([]string, 0, len(v.GetSpec().RequiredVars))
			for _, t := range v.GetSpec().RequiredVars {
				vars = append(vars, t.Var)
			}
			rawTmplVars := strings.Join(vars, ",")

			var forRoles, forbiddenRoles string
			ann := v.GetAnnotations()
			if ann != nil {
				forRoles = ann[cosmov1alpha1.TemplateAnnKeyUserRoles]
				forbiddenRoles = ann[cosmov1alpha1.TemplateAnnKeyForbiddenUserRoles]
			}

			tmplType := v.GetLabels()[cosmov1alpha1.TemplateLabelKeyType]
			rowdata := []string{v.GetName(), rawTmplVars, tmplType, strconv.FormatBool(v.GetScope() == meta.RESTScopeRoot), forRoles, forbiddenRoles}
			fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
		}
	}

	return nil
}
