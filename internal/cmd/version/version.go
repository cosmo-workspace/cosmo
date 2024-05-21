package version

import (
	"fmt"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, o *cli.RootOptions) {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "cosmoctl - cosmo-workspace %s commit=%s build=%s\n",
				o.Versions.Version, o.Versions.Commit, o.Versions.Date)
		},
	}
	cmd.AddCommand(versionCmd)
}
