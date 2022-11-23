package netrule

import (
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, co *cmdutil.CliOptions) {
	netruleCmd := &cobra.Command{
		Use:   "networkrule",
		Short: "Manipulate NetworkRule of Workspace resource",
		Long: `
Workspace network rule utility command
`,
		Aliases: []string{"netrule", "net"},
	}

	o := cmdutil.NewUserNamespacedCliOptions(co)

	netruleCmd.PersistentFlags().StringVarP(&o.User, "user", "u", "", "user name")
	netruleCmd.PersistentFlags().StringVarP(&o.Namespace, "namespace", "n", "", "namespace")

	netruleCmd.AddCommand(CreateCmd(&cobra.Command{
		Use:     "create NETWORK_RULE_NAME --workspace WORKSPACE_NAME --port PORT_NUMBER",
		Short:   "Create or update workspace network rule",
		Aliases: []string{"add"},
	}, o))
	netruleCmd.AddCommand(DeleteCmd(&cobra.Command{
		Use:     "delete NETWORK_RULE_NAME --workspace WORKSPACE_NAME",
		Short:   "Delete workspace network rule",
		Aliases: []string{"rm"},
	}, o))

	cmd.AddCommand(netruleCmd)
}
