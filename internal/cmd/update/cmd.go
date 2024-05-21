package update

import (
	"github.com/cosmo-workspace/cosmo/internal/cmd/user"
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, o *cli.RootOptions) {
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update cosmo resources",
	}

	updateCmd.AddCommand(user.UpdateCmd(&cobra.Command{
		Use:     "user USER_NAME",
		Short:   "Update user. Alias of 'cosmoctl user update'",
		Aliases: []string{"us"},
	}, o))

	updateCmd.AddCommand(workspace.UpdateCmd(&cobra.Command{
		Use:     "workspace WORKSPACE_NAME",
		Short:   "Update workspace. Alias of 'cosmoctl workspace update'",
		Aliases: []string{"ws"},
	}, o))

	updateCmd.AddCommand(workspace.UpsertNetworkCmd(&cobra.Command{
		Use:     "network WORKSPACE_NAME --port 8080",
		Short:   "Upsert workspace network. Alias of 'cosmoctl workspace upsert-network'",
		Aliases: []string{"net", "workspace-network", "workspace-networks", "ws-net", "wsnet"},
	}, o))

	cmd.AddCommand(updateCmd)
}
