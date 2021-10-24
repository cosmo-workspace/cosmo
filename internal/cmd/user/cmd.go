package user

import (
	"github.com/spf13/cobra"

	cmdutil "github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

func AddCommand(cmd *cobra.Command, o *cmdutil.CliOptions) {
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manipulate User",
		Long: `
Manipulate Workspaces like COSMO Dashboard UI.

User is actually a Kubernetes Namespace for running Workspaces.

Password is used for the authentication of COSMO Auth Proxy and COSMO Dashboard UI
`,
	}

	userCmd.AddCommand(resetPasswordCmd(o))
	userCmd.AddCommand(createCmd(o))
	userCmd.AddCommand(getCmd(o))
	userCmd.AddCommand(deleteCmd(o))
	userCmd.AddCommand(updateCmd(o))

	cmd.AddCommand(userCmd)
}
