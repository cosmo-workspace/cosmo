package user

import (
	"github.com/spf13/cobra"

	cmdutil "github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

func AddCommand(cmd *cobra.Command, o *cmdutil.CliOptions) {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manipulate User resource",
		Long: `
Manipulate Users like COSMO Dashboard UI.

User is actually a Kubernetes Namespace for running Workspaces.
`,
	}

	userCmd.AddCommand(resetPasswordCmd(&cobra.Command{
		Use:   "reset-password USER_NAME",
		Short: "Reset user password",
	}, o))
	userCmd.AddCommand(CreateCmd(&cobra.Command{
		Use:   "create USER_NAME --role cosmo-admin",
		Short: "Create user",
	}, o))
	userCmd.AddCommand(GetCmd(&cobra.Command{
		Use:   "get",
		Short: "Get users",
		Long: `
Get Users. This command is similar to "kubectl get namespace"
`,
	}, o))
	userCmd.AddCommand(DeleteCmd(&cobra.Command{
		Use:     "delete USER_NAME",
		Aliases: []string{"del"},
		Short:   "Delete user",
	}, o))
	userCmd.AddCommand(updateCmd(&cobra.Command{
		Use:   "update USER_NAME --role ROLE --name NAME",
		Short: "Update user",
	}, o))

	cmd.AddCommand(userCmd)
}
