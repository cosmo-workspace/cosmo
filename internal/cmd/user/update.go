package user

import (
	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
)

func UpdateCmd(cmd *cobra.Command, o *cli.RootOptions) *cobra.Command {
	cmd.AddCommand(UpdateDisplayNameCmd(&cobra.Command{
		Use:     "display-name USER_NAME",
		Aliases: []string{"displayname", "name"},
		Short:   "Update display name",
	}, o))
	cmd.AddCommand(UpdateRoleCmd(&cobra.Command{
		Use:   "role USER_NAME",
		Short: "Update role",
	}, o))
	cmd.AddCommand(UpdateAddonCmd(&cobra.Command{
		Use:     "addon USER_NAME",
		Aliases: []string{"addon", "useraddon", "user-addon"},
		Short:   "Update addon",
	}, o))
	return cmd
}
