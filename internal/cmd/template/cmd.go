package template

import (
	cmdutil "github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, o *cmdutil.CliOptions) {
	tmplCmd := &cobra.Command{
		Use:   "template",
		Short: "Manipulate Template resource",
		Long: `
Template utility command.
`,
		Aliases: []string{"tmpl"},
	}

	tmplCmd.AddCommand(generateCmd(&cobra.Command{
		Use:     "generate --name TEMPLATE_NAME [< Input via Stdin or pipe]",
		Aliases: []string{"gen"},
		Short:   "Generate Template",
		Long: `Generate Template

For create generated template, just do "kubectl create -f cosmo-template.yaml"

Example:
  * Pipe from kustomize build and apply to your cluster in a single line 
	
      kustomize build ./kubernetes/ | cosmoctl template generate --name TEMPLATE_NAME | kubectl apply -f -

  * Input merged config file (kustomize build ... or helm template ... etc.) and save it to file

      cosmoctl template generate --name TEMPLATE_NAME -o cosmo-template.yaml < merged.yaml
`,
	}, o))
	tmplCmd.AddCommand(GetCmd(&cobra.Command{
		Use:   "get",
		Short: "Get templates",
		Long: `Get Templates

Basically it is similar to "kubectl get template"

For type workspace template, use with --workspace flag to see more information. 
`,
	}, o))
	tmplCmd.AddCommand(validateCmd(&cobra.Command{
		Use:     "validate --file FILE",
		Aliases: []string{"valid", "check"},
		Short:   "Validate Template",
		Long: `Validate Template by dry-run

Usage:
  * Dry-run on server-side
	
      cosmoctl template validate -f cosmo-template.yaml

  * Dry-run on client-side using kubectl
	
      cosmoctl template validate -f cosmo-template.yaml --client

  * Input from stdin not file.

      cat cosmo-template.yaml | cosmoctl template validate -f -
`,
	}, o))

	cmd.AddCommand(tmplCmd)
}
