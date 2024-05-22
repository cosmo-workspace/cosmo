package resume

import (
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, o *cli.RootOptions) {
	resumeCmd := &cobra.Command{
		Use:     "resume",
		Short:   "Start stopped workspaces",
		Aliases: []string{"start", "run"},
	}

	resumeCmd.AddCommand(workspace.ResumeCmd(&cobra.Command{
		Use:     "workspace WORKSPACE_NAME...",
		Short:   "Resume workspaces. Alias of 'cosmoctl workspace resume'",
		Aliases: []string{"ws", "workspaces"},
	}, o))
	cmd.AddCommand(resumeCmd)
}
