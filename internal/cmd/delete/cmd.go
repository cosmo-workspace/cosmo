package get

import (
	"github.com/cosmo-workspace/cosmo/internal/cmd/netrule"
	"github.com/cosmo-workspace/cosmo/internal/cmd/user"
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, co *cmdutil.CliOptions) {
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete cosmo resources",
		Long: `
Delete cosmo resources
`,
	}

	o := cmdutil.NewUserNamespacedCliOptions(co)

	deleteCmd.PersistentFlags().StringVarP(&o.User, "user", "u", "", "user name")
	deleteCmd.PersistentFlags().StringVarP(&o.Namespace, "namespace", "n", "", "namespace")

	deleteCmd.AddCommand(workspace.DeleteCmd(&cobra.Command{
		Use:     "workspace WORKSPACE_NAME",
		Aliases: []string{"ws"},
		Short:   "Delete workspace",
	}, o))
	deleteCmd.AddCommand(user.DeleteCmd(&cobra.Command{
		Use:   "user USER_NAME",
		Short: "Delete user",
	}, o.CliOptions))
	deleteCmd.AddCommand(netrule.DeleteCmd(&cobra.Command{
		Use:     "networkrule NETWORK_RULE_NAME --workspace WORKSPACE_NAME --port PORT_NUMBER",
		Short:   "Create or update workspace network rule",
		Aliases: []string{"netrule", "net"},
	}, o))

	cmd.AddCommand(deleteCmd)
}
