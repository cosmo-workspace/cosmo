package get

import (
	"github.com/cosmo-workspace/cosmo/internal/cmd/template"
	"github.com/cosmo-workspace/cosmo/internal/cmd/user"
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, co *cmdutil.CliOptions) {
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get cosmo resources",
		Long: `
Get cosmo resources
`,
	}

	o := cmdutil.NewUserNamespacedCliOptions(co)

	getCmd.PersistentFlags().StringVarP(&o.User, "user", "u", "", "user name")
	getCmd.PersistentFlags().StringVarP(&o.Namespace, "namespace", "n", "", "namespace")

	getCmd.AddCommand(workspace.GetCmd(&cobra.Command{
		Use:     "workspace WORKSPACE_NAME",
		Aliases: []string{"ws"},
		Short:   "Get workspace",
	}, o))
	getCmd.AddCommand(user.GetCmd(&cobra.Command{
		Use:   "user USER_NAME",
		Short: "Get user",
	}, o.CliOptions))
	getCmd.AddCommand(template.GetCmd(&cobra.Command{
		Use:     "template WORKSPACE_NAME",
		Aliases: []string{"tmpl"},
		Short:   "Get template",
	}, o.CliOptions))

	cmd.AddCommand(getCmd)
}
