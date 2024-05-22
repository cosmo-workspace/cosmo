package workspace

import (
	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
)

func AddCommand(cmd *cobra.Command, o *cli.RootOptions) {
	workspaceCmd := &cobra.Command{
		Use:     "workspace",
		Short:   "Manipulate Workspace resource",
		Aliases: []string{"ws"},
		Long: `
Manipulate COSMO Workspace resource.

"Workspace" is a namespaced Kubernetes CRD which represents a instance of workspace.
`,
	}

	workspaceCmd.AddCommand(CreateCmd(&cobra.Command{
		Use:   "create WORKSPACE_NAME --template TEMPLATE_NAME",
		Short: "Create workspace",
	}, o))
	workspaceCmd.AddCommand(GetCmd(&cobra.Command{
		Use:     "get [WORKSPACE_NAME...]",
		Short:   "Get workspaces",
		Aliases: []string{"list"},
	}, o))
	workspaceCmd.AddCommand(GetTemplatesCmd(&cobra.Command{
		Use:     "templates [TEMPLATE_NAME...]",
		Short:   "Get workspace templates in cluster",
		Aliases: []string{"template", "tmpls", "tmpl", "get-templates", "get-template", "get-tmpls", "get-tmpl"},
	}, o))
	workspaceCmd.AddCommand(DeleteCmd(&cobra.Command{
		Use:     "delete WORKSPACE_NAME...",
		Short:   "Delete workspaces",
		Aliases: []string{"rm"},
	}, o))
	workspaceCmd.AddCommand(ResumeCmd(&cobra.Command{
		Use:     "resume WORKSPACE_NAME",
		Short:   "Resume stopped workspace pod",
		Aliases: []string{"start", "run"},
	}, o))
	workspaceCmd.AddCommand(SuspendCmd(&cobra.Command{
		Use:     "suspend WORKSPACE_NAME",
		Short:   "Suspend workspace pod",
		Aliases: []string{"stop"},
	}, o))
	workspaceCmd.AddCommand(GetNetworkCmd(&cobra.Command{
		Use:     "network WORKSPACE_NAME",
		Short:   "Get workspace network",
		Aliases: []string{"net", "get-network", "get-networks", "get-net"},
	}, o))
	workspaceCmd.AddCommand(UpsertNetworkCmd(&cobra.Command{
		Use:     "upsert-network WORKSPACE_NAME --port 8080",
		Short:   "Upsert workspace network",
		Aliases: []string{"add-net"},
	}, o))
	workspaceCmd.AddCommand(RemoveNetworkCmd(&cobra.Command{
		Use:     "remove-network WORKSPACE_NAME --port 8080",
		Short:   "Remove workspace network",
		Aliases: []string{"rm-net", "remove-net", "delete-net", "delete-network"},
	}, o))
	workspaceCmd.AddCommand(UpdateCmd(&cobra.Command{
		Use:   "update WORKSPACE_NAME",
		Short: "Update workspace",
	}, o))

	cmd.AddCommand(workspaceCmd)
}
