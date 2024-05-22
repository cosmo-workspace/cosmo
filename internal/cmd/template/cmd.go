package template

import (
	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/internal/cmd/user"
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
)

func AddCommand(cmd *cobra.Command, o *cli.RootOptions) {
	templateCmd := &cobra.Command{
		Use:     "template",
		Short:   "Manipulate Template resource",
		Aliases: []string{"tmpl"},
		Long: `
Manipulate COSMO Workspace Template resource.

"Template" is a set of Kubernetes resources for Workspace.
`,
	}
	generateCmd := &cobra.Command{
		Use:     "generate [< Input via Stdin or pipe]",
		Short:   "Generate Template",
		Aliases: []string{"gen"},
	}
	generateCmd.AddCommand(generateWorkspaceCmd(&cobra.Command{
		Use:   "workspace [< Input via Stdin or pipe]",
		Short: "Generate WorkspaceTemplate",
		Long: `Generate WorkspaceTemplate

For create generated Workspace Template, just do kubectl apply
`,
		Example: `
  * Pipe from kustomize build and apply to your cluster in a single line 
	
      kustomize build ./kubernetes/ | cosmoctl gen tmpl --name TEMPLATE_NAME | kubectl apply -f -

  * Input merged config file (kustomize build ... or helm template ... etc.) and save it to file

      cosmoctl gen tmpl --name TEMPLATE_NAME -o cosmo-template.yaml < merged.yaml
`,
		Aliases: []string{"workspace", "ws", "workspace-template"},
	}, o))
	generateCmd.AddCommand(generateUserAddonCmd(&cobra.Command{
		Use:   "useraddon [< Input via Stdin or pipe]",
		Short: "Generate UserAddon",
		Long: `Generate UserAddon

For create generated UserAddon Template, just do kubectl apply
`,
		Example: `
  * Pipe from kustomize build and apply to your cluster in a single line 
	
      kustomize build ./kubernetes/ | cosmoctl gen addon --name TEMPLATE_NAME | kubectl apply -f -

  * Input merged config file (kustomize build ... or helm template ... etc.) and save it to file

      cosmoctl gen addon --name TEMPLATE_NAME -o cosmo-template.yaml < merged.yaml
`,
		Aliases: []string{"addon", "useraddon", "user-addon"},
	}, o))

	templateCmd.AddCommand(validateCmd(&cobra.Command{
		Use:     "validate --file FILE",
		Aliases: []string{"valid", "check"},
		Short:   "Validate Template by dry-run",
		Example: `
  * Dry-run on server-side
	
      cosmoctl template validate -f cosmo-template.yaml

  * Dry-run on client-side using kubectl
	
      cosmoctl template validate -f cosmo-template.yaml --client

  * Input from stdin not file.

      cat cosmo-template.yaml | cosmoctl template validate -f -
`,
	}, o))

	getCmd := &cobra.Command{
		Use:     "get",
		Short:   "Get Templates",
		Aliases: []string{"list"},
	}
	getCmd.AddCommand(workspace.GetTemplatesCmd(&cobra.Command{
		Use:     "workspace [TEMPLATE_NAME...]",
		Short:   "Get workspace templates in cluster",
		Aliases: []string{"workspaces", "workspace", "ws"},
	}, o))
	getCmd.AddCommand(user.GetAddonsCmd(&cobra.Command{
		Use:     "useraddons [ADDON_NAME...]",
		Short:   "Get addons",
		Aliases: []string{"useraddon", "addons", "addon", "user-addon"},
	}, o))

	templateCmd.AddCommand(getCmd)
	templateCmd.AddCommand(generateCmd)
	cmd.AddCommand(templateCmd)
}
