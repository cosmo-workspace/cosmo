package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/printers"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type getOption struct {
	*cmdutil.CliOptions
}

func getCmd(cliOpt *cmdutil.CliOptions) *cobra.Command {
	o := &getOption{CliOptions: cliOpt}

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get users",
		Long: `
Get Users. This command is similar to "kubectl get namespace"
`,
		PersistentPreRunE: o.PreRunE,
		RunE:              cmdutil.RunEHandler(o.RunE),
	}
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
	return nil
}

func (o *getOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CliOptions.Complete(cmd, args); err != nil {
		return err
	}
	return nil
}

func (o *getOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	users, err := c.ListUsers(ctx)
	if err != nil {
		return err
	}
	o.Logr.DebugAll().Info("ListUsers", "users", users)

	w := printers.GetNewTabWriter(o.Out)
	defer w.Flush()

	columnNames := []string{"ID", "NAME", "ROLE", "NAMESPACE", "STATUS"}
	fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))
	for _, v := range users {
		rowdata := []string{v.Name, v.Spec.DisplayName, v.Spec.Role.String(), v.Status.Namespace.Name, string(v.Status.Phase)}
		fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
	}

	return nil
}
