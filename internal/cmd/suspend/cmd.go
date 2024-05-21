package suspend

import (
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, o *cli.RootOptions) {
	suspendCmd := &cobra.Command{
		Use:     "suspend",
		Short:   "Suspend workspaces",
		Aliases: []string{"stop"},
	}

	suspendCmd.AddCommand(workspace.SuspendCmd(&cobra.Command{
		Use:     "workspace WORKSPACE_NAME...",
		Short:   "Suspend workspaces. Alias of 'cosmoctl workspace suspend'",
		Aliases: []string{"ws", "workspaces"},
	}, o))
	cmd.AddCommand(suspendCmd)
}
