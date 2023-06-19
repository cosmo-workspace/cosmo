package template

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	cmdutil "github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	"github.com/cosmo-workspace/cosmo/pkg/transformer"
)

type validateOption struct {
	*cmdutil.CliOptions

	File               string
	RawVars            string
	DryrunOnClientSide bool

	input []byte
	tmpl  cosmov1alpha1.Template
	vars  map[string]string
}

func validateCmd(cmd *cobra.Command, cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &validateOption{CliOptions: cliOpt}
	cmd.PersistentPreRunE = o.PreRunE
	cmd.RunE = cmdutil.RunEHandler(o.RunE)
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "input COSMO Template file yaml path. when specified '-', input from Stdin")
	cmd.Flags().StringVar(&o.RawVars, "vars", "", "template vars. the format is VarName:VarValue. also it can be set multiple vars by conma separated list. (example: VAR1:VAL1,VAR2:VAL2)")
	cmd.Flags().BoolVar(&o.DryrunOnClientSide, "client", false, "dry-run on client-side. kubectl is required to be executable in PATH")

	return cmd
}

func (o *validateOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *validateOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if o.File == "" {
		return errors.New("--file is required")
	}
	return nil
}

func (o *validateOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}

	// load template
	var input []byte
	var err error
	if o.File == "-" {
		if isatty.IsTerminal(os.Stdin.Fd()) {
			return fmt.Errorf("no input via stdin")
		}
		// input data from stdin
		input, err = io.ReadAll(o.In)
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

	} else {
		input, err = os.ReadFile(o.File)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
	}

	if len(input) == 0 {
		return fmt.Errorf("no input")
	}
	o.Logr.DebugAll().Info(string(input))
	o.input = input

	// unmarshal to Template struct
	var tmpl cosmov1alpha1.Template
	if err := yaml.Unmarshal(input, &tmpl); err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %w", err)
	}
	o.tmpl = tmpl

	// parse vars
	vars := make(map[string]string)
	if o.RawVars != "" {
		varAndVals := strings.Split(o.RawVars, ",")
		for _, v := range varAndVals {
			varAndVal := strings.Split(v, ":")
			if len(varAndVal) != 2 {
				return fmt.Errorf("vars format error: vars %s must be 'VAR:VAL'", v)
			}
			vars[varAndVal[0]] = varAndVal[1]
		}
	}

	for _, rqvar := range tmpl.Spec.RequiredVars {
		if _, exist := vars[rqvar.Var]; !exist {
			if rqvar.Default == "" {
				return fmt.Errorf("required vars not given. set --var %s:<TEST_VAR>", rqvar.Var)
			} else {
				vars[rqvar.Var] = rqvar.Default
			}
		}
	}
	o.vars = vars

	return nil
}

func (o *validateOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()

	o.Logr.Info("smoke test: dryrun apply template")
	_, tmplUnst, err := template.StringToUnstructured(string(o.input))
	if err != nil {
		return err
	}
	if err := o.dryrunApplyOnServer(ctx, tmplUnst); err != nil {
		return err
	}

	// gen dummy instance sufix
	g, err := password.NewGenerator(&password.GeneratorInput{Symbols: "", Digits: ""})
	if err != nil {
		return fmt.Errorf("failed to create password generator: %w", err)
	}

	sufix, err := g.Generate(8, 0, 0, true, true)
	if err != nil {
		return fmt.Errorf("failed to generate random string: %w", err)
	}

	dummyInst := cosmov1alpha1.Instance{}
	dummyInst.SetName(fmt.Sprintf("cosmoctl-validate-%s", sufix))
	dummyInst.SetNamespace("default")
	dummyInst.SetUID(types.UID(sufix))
	dummyInst.Spec.Vars = o.vars

	o.Logr.Info("smoke test: create dummy instance to apply each resources", "instance", dummyInst.GetName())
	o.Logr.Debug().DumpObject(o.Scheme, &dummyInst, "test instance")

	builts, err := template.BuildObjects(o.tmpl.Spec, &dummyInst)
	if err != nil {
		return fmt.Errorf("failed to build test instance: %w", err)
	}
	// only apply MetadataTransformer
	ts := []transformer.Transformer{transformer.NewMetadataTransformer(&dummyInst, o.Scheme, template.IsDisableNamePrefix(&o.tmpl))}
	builts, err = transformer.ApplyTransformers(ctx, ts, builts)
	if err != nil {
		return fmt.Errorf("failed to transform objects: %w", err)
	}

	w := printers.GetNewTabWriter(o.Out)
	defer w.Flush()
	columnNames := []string{"APIVERSION", "KIND", "NAME", "RESULT", "MESSAGE"}
	fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))

	for _, built := range builts {
		o.Logr.Info("smoke test: dryrun applying dummy resource",
			"apiVersion", built.GetAPIVersion(), "kind", built.GetKind())

		o.Logr.Debug().DumpObject(o.Scheme, &built, "validating object")
		if o.DryrunOnClientSide {
			err = o.kubectlDryrunApplyOnClient(ctx, &built)
		} else {
			err = o.dryrunApplyOnServer(ctx, &built)
		}
		rowdata := resultRow(built.GetAPIVersion(), built.GetKind(), built.GetName(), err)
		fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
	}

	return nil
}

func (o *validateOption) dryrunApplyOnServer(ctx context.Context, obj client.Object) error {
	options := &client.PatchOptions{
		FieldManager: "cosmoctl-validate",
		Force:        pointer.Bool(true),
		DryRun:       []string{metav1.DryRunAll},
	}

	if err := o.Client.Patch(ctx, obj, client.Apply, options); err != nil {
		return fmt.Errorf("dryrun failed: %w", err)
	}
	return nil
}

func (o *validateOption) kubectlDryrunApplyOnClient(ctx context.Context, obj runtime.Object) error {
	b, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(b)

	cmd := exec.CommandContext(ctx, "kubectl", "apply", "--dry-run=client", "-f", "-")
	cmd.Stdin = buf

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to exec kubectl: %w : %s", err, out)
	}
	return nil
}

func resultRow(apiVersion, kind, name string, err error) []string {
	var result, errMsg string
	if err == nil {
		result = "OK"
	} else {
		result = "NG"
		errMsg = err.Error()
	}
	return []string{apiVersion, kind, name, result, errMsg}
}
