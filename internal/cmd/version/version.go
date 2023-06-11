package version

import (
	"fmt"

	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/spf13/cobra"
)

const Footprint = `cosmoctl - cosmo v0.9.0 cosmo-workspace 2023`

func AddCommand(cmd *cobra.Command, o *cmdutil.CliOptions) {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(o.Out, Footprint)
		},
	}
	cmd.AddCommand(versionCmd)
}
