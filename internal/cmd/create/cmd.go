package create

import (
	"github.com/cosmo-workspace/cosmo/internal/cmd/netrule"
	"github.com/cosmo-workspace/cosmo/internal/cmd/user"
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, co *cmdutil.CliOptions) {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create cosmo resources",
		Long: `
Create cosmo resources
`,
	}

	o := cmdutil.NewUserNamespacedCliOptions(co)

	createCmd.AddCommand(workspace.CreateCmd(&cobra.Command{
		Use:     "workspace WORKSPACE_NAME --template TEMPLATE_NAME",
		Short:   "Create workspace",
		Example: "workspace my-code-server --user example-user --template code-server --vars PVC_SIZE_Gi:10",
		Aliases: []string{"ws"},
	}, o))
	createCmd.AddCommand(user.CreateCmd(&cobra.Command{
		Use:   "user USER_NAME --role cosmo-admin",
		Short: "Create user",
	}, o.CliOptions))
	createCmd.AddCommand(netrule.CreateCmd(&cobra.Command{
		Use:     "networkrule NETWORK_RULE_NAME --workspace WORKSPACE_NAME --port PORT_NUMBER",
		Short:   "Create or update workspace network rule",
		Aliases: []string{"netrule", "net"},
	}, o))
	cmd.AddCommand(createCmd)
}
