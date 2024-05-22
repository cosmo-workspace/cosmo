/*
Copyright © 2024 cosmo-workspace
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/internal/cmd/create"
	del "github.com/cosmo-workspace/cosmo/internal/cmd/delete"
	"github.com/cosmo-workspace/cosmo/internal/cmd/get"
	"github.com/cosmo-workspace/cosmo/internal/cmd/login"
	"github.com/cosmo-workspace/cosmo/internal/cmd/resume"
	"github.com/cosmo-workspace/cosmo/internal/cmd/suspend"
	"github.com/cosmo-workspace/cosmo/internal/cmd/template"
	"github.com/cosmo-workspace/cosmo/internal/cmd/user"
	"github.com/cosmo-workspace/cosmo/internal/cmd/version"
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
)

func NewRootCmd(o *cli.RootOptions) *cobra.Command {

	rootCmd := &cobra.Command{
		Use:   "cosmoctl",
		Short: "Command line tool for cosmo API",
		Long: `
Command line tool for cosmo API
Complete documentation is available at http://github.com/cosmo-workspace/cosmo

MIT 2024 cosmo-workspace/cosmo
`,
	}
	o.AddFlags(rootCmd)

	version.AddCommand(rootCmd, o)
	login.AddCommand(rootCmd, o)

	create.AddCommand(rootCmd, o)
	get.AddCommand(rootCmd, o)
	del.AddCommand(rootCmd, o)
	resume.AddCommand(rootCmd, o)
	suspend.AddCommand(rootCmd, o)

	user.AddCommand(rootCmd, o)
	workspace.AddCommand(rootCmd, o)
	template.AddCommand(rootCmd, o)

	return rootCmd
}

func Execute(v cli.VersionInfo) {
	o := cli.NewRootOptions()
	o.Versions = v
	rootCmd := NewRootCmd(o)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(rootCmd.ErrOrStderr(), color.RedString("Error: %s", err))
		os.Exit(1)
	}

}
