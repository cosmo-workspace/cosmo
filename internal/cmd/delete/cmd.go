package delete

import (
	"github.com/cosmo-workspace/cosmo/internal/cmd/user"
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, o *cli.RootOptions) {
	deleteCmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete cosmo resources",
		Aliases: []string{"rm", "remove"},
	}

	deleteCmd.AddCommand(user.DeleteCmd(&cobra.Command{
		Use:     "user USER_NAME...",
		Short:   "Delete users. Alias of 'cosmoctl user delete'",
		Aliases: []string{"us", "users"},
	}, o))
	deleteCmd.AddCommand(workspace.DeleteCmd(&cobra.Command{
		Use:     "workspace WORKSPACE_NAME...",
		Short:   "Delete workspaces. Alias of 'cosmoctl workspace delete'",
		Aliases: []string{"ws", "workspaces"},
	}, o))
	deleteCmd.AddCommand(workspace.RemoveNetworkCmd(&cobra.Command{
		Use:     "network WORKSPACE_NAME --port 8080",
		Short:   "Remove workspace network. Alias of 'cosmoctl workspace remove-network'",
		Aliases: []string{"net", "workspace-network", "workspace-networks", "ws-net", "wsnet"},
	}, o))
	cmd.AddCommand(deleteCmd)
}
