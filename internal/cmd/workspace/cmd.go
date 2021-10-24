package workspace

import (
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, co *cmdutil.CliOptions) {
	workspaceCmd := &cobra.Command{
		Use:   "workspace",
		Short: "Workspace utility command",
		Long: `
Workspace utility command. Manipulate Workspaces like COSMO Dashboard UI.

For Workspace detailed status or trouble shooting, 
use "kubectl describe workspace" or "kubectl describe instance" and see controller's events.
`,
		Aliases: []string{"ws"},
	}

	o := cmdutil.NewUserNamespacedCliOptions(co)

	workspaceCmd.PersistentFlags().StringVarP(&o.User, "user", "u", "", "user ID")
	workspaceCmd.PersistentFlags().StringVarP(&o.Namespace, "namespace", "n", "", "namespace")
	workspaceCmd.PersistentFlags().BoolVarP(&o.AllNamespace, "all-namespaces", "A", false, "all namespaces")

	workspaceCmd.AddCommand(getCmd(o))
	workspaceCmd.AddCommand(createCmd(o))
	workspaceCmd.AddCommand(deleteCmd(o))
	workspaceCmd.AddCommand(runInstanceCmd(o))
	workspaceCmd.AddCommand(stopInstanceCmd(o))
	workspaceCmd.AddCommand(openPortCmd(o))
	workspaceCmd.AddCommand(closePortCmd(o))

	cmd.AddCommand(workspaceCmd)
}
