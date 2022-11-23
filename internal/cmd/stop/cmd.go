package stop

import (
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, co *cmdutil.CliOptions) {
	runCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop workload resources",
		Long: `
Stop cosmo workload resources
`,
	}

	o := cmdutil.NewUserNamespacedCliOptions(co)

	runCmd.PersistentFlags().StringVarP(&o.User, "user", "u", "", "user name")
	runCmd.PersistentFlags().StringVarP(&o.Namespace, "namespace", "n", "", "namespace")

	runCmd.AddCommand(workspace.StopInstanceCmd(&cobra.Command{
		Use:     "workspace WORKSPACE_NAME",
		Aliases: []string{"ws", "inst", "instance"},
		Short:   "Stop workspace instance",
	}, o))

	cmd.AddCommand(runCmd)
}
