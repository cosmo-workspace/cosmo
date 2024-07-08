package workspace

import (
	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
)

func UpdateCmd(cmd *cobra.Command, o *cli.RootOptions) *cobra.Command {
	cmd.AddCommand(UpdateVarsCmd(&cobra.Command{
		Use:   "vars WORKSPACE_NAME",
		Short: "Update workspace vars",
	}, o))
	cmd.AddCommand(UpdateDeletePolicyCmd(&cobra.Command{
		Use:     "deletepolicy WORKSPACE_NAME [delete|keep]",
		Aliases: []string{"delete-policy"},
		Short:   "Update delete polocy",
	}, o))
	return cmd
}
