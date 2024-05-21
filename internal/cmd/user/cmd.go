package user

import (
	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
)

func AddCommand(cmd *cobra.Command, o *cli.RootOptions) {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manipulate User resource",
		Long: `
Manipulate COSMO User resource.

"User" is a cluster-scoped Kubernetes CRD which represents a developer or user who use Workspace.

Once you create User, Kubernetes Namespace is created and bound to the User.
`,
	}

	userCmd.AddCommand(resetPasswordCmd(&cobra.Command{
		Use:   "reset-password USER_NAME",
		Short: "Reset password",
	}, o))
	userCmd.AddCommand(changePasswordCmd(&cobra.Command{
		Use:   "change-password [USER_NAME]",
		Short: "Change password",
	}, o))
	userCmd.AddCommand(CreateCmd(&cobra.Command{
		Use:   "create USER_NAME",
		Short: "Create user",
	}, o))
	userCmd.AddCommand(GetCmd(&cobra.Command{
		Use:     "get [USER_NAME...]",
		Short:   "Get users",
		Aliases: []string{"list"},
	}, o))
	userCmd.AddCommand(GetAddonsCmd(&cobra.Command{
		Use:     "get-addons [ADDON_NAME...]",
		Short:   "Get addons",
		Aliases: []string{"get-addon", "get-addons", "addons", "addon"},
	}, o))
	userCmd.AddCommand(GetEventsCmd(&cobra.Command{
		Use:     "get-events [USER_NAME]",
		Short:   "Get events for user",
		Aliases: []string{"get-events", "get-event", "events", "event"},
	}, o))
	userCmd.AddCommand(DeleteCmd(&cobra.Command{
		Use:     "delete USER_NAME...",
		Aliases: []string{"rm"},
		Short:   "Delete users",
	}, o))
	userCmd.AddCommand(UpdateCmd(&cobra.Command{
		Use:   "update",
		Short: "Update user",
	}, o))
	cmd.AddCommand(userCmd)
}
