package run

import (
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, co *cmdutil.CliOptions) {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run workload resources",
		Long: `
Run cosmo workload resources
`,
	}

	o := cmdutil.NewUserNamespacedCliOptions(co)

	runCmd.PersistentFlags().StringVarP(&o.User, "user", "u", "", "user name")
	runCmd.PersistentFlags().StringVarP(&o.Namespace, "namespace", "n", "", "namespace")

	runCmd.AddCommand(workspace.RunInstanceCmd(&cobra.Command{
		Use:     "workspace WORKSPACE_NAME",
		Aliases: []string{"ws", "inst", "instance"},
		Short:   "Run workspace instance",
	}, o))

	cmd.AddCommand(runCmd)
}
