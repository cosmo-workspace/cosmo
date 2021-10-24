package template

import (
	cmdutil "github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func AddCommand(cmd *cobra.Command, o *cmdutil.CliOptions) {
	tmplCmd := &cobra.Command{
		Use:   "template",
		Short: "Template utility command",
		Long: `
Template utility command such as template generation. 

For create generated template, just do "kubectl create -f cosmo-template.yaml"
`,
		Aliases: []string{"tmpl"},
	}

	tmplCmd.AddCommand(generateCmd(o))
	tmplCmd.AddCommand(getCmd(o))

	cmd.AddCommand(tmplCmd)
}
