package workspace

import (
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, co *cmdutil.CliOptions) {
	workspaceCmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manipulate Workspace resource",
		Long: `
Workspace utility command. Manipulate Workspaces like COSMO Dashboard UI.

For Workspace detailed status or trouble shooting, 
use "kubectl describe workspace" or "kubectl describe instance" and see controller's events.
`,
		Aliases: []string{"ws"},
	}

	o := cmdutil.NewUserNamespacedCliOptions(co)

	workspaceCmd.PersistentFlags().StringVarP(&o.User, "user", "u", "", "user name")
	workspaceCmd.PersistentFlags().StringVarP(&o.Namespace, "namespace", "n", "", "namespace")
	workspaceCmd.PersistentFlags().BoolVarP(&o.AllNamespace, "all-namespaces", "A", false, "all namespaces")

	workspaceCmd.AddCommand(GetCmd(&cobra.Command{
		Use:   "get [WORKSPACE_NAME]",
		Short: "Get workspaces",
		Long: `
Get workspaces

This command is like "kubectl get workspace" but show more information.

But for Workspace detailed status or trouble shooting, 
use "kubectl describe workspace" or "kubectl describe instance" and see controller's events.
`,
	}, o))
	workspaceCmd.AddCommand(CreateCmd(&cobra.Command{
		Use:     "create WORKSPACE_NAME --template TEMPLATE_NAME",
		Short:   "Create workspace",
		Example: "create my-code-server --user example-user --template code-server --vars PVC_SIZE_Gi:10",
	}, o))
	workspaceCmd.AddCommand(DeleteCmd(&cobra.Command{
		Use:     "delete WORKSPACE_NAME",
		Aliases: []string{"del"},
		Short:   "Delete workspace",
	}, o))
	workspaceCmd.AddCommand(RunInstanceCmd(&cobra.Command{
		Use:     "run-instance WORKSPACE_NAME",
		Aliases: []string{"run"},
		Short:   "Run workspace instance",
	}, o))
	workspaceCmd.AddCommand(StopInstanceCmd(&cobra.Command{
		Use:     "stop-instance WORKSPACE_NAME",
		Aliases: []string{"stop"},
		Short:   "Stop workspace instance",
	}, o))

	cmd.AddCommand(workspaceCmd)
}
