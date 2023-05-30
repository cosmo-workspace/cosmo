/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/internal/cmd/create"
	del "github.com/cosmo-workspace/cosmo/internal/cmd/delete"
	"github.com/cosmo-workspace/cosmo/internal/cmd/get"
	"github.com/cosmo-workspace/cosmo/internal/cmd/netrule"
	"github.com/cosmo-workspace/cosmo/internal/cmd/run"
	"github.com/cosmo-workspace/cosmo/internal/cmd/stop"
	"github.com/cosmo-workspace/cosmo/internal/cmd/template"
	"github.com/cosmo-workspace/cosmo/internal/cmd/user"
	"github.com/cosmo-workspace/cosmo/internal/cmd/version"
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

func NewRootCmd(o *cmdutil.CliOptions) *cobra.Command {

	rootCmd := &cobra.Command{
		Use:   "cosmoctl",
		Short: "Command line tool to manipulate comso",
		Long: `
Command line tool to manipulate comso
Complete documentation is available at http://github.com/cosmo-workspace/cosmo

MIT 2022 cosmo-workspace/cosmo
`,
	}

	rootCmd.SetIn(o.In)
	rootCmd.SetOut(o.Out)
	rootCmd.SetErr(o.ErrOut)
	rootCmd.PersistentFlags().StringVar(&o.KubeConfigPath, "kubeconfig", "", "kubeconfig file path (default: $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringVar(&o.KubeContext, "context", "", "kube-context (default: current context)")
	rootCmd.PersistentFlags().IntVarP(&o.LogLevel, "verbose", "v", -1, "log level. -1:DISABLED, 0:INFO, 1:DEBUG, 2:ALL")

	version.AddCommand(rootCmd, o)
	template.AddCommand(rootCmd, o)
	user.AddCommand(rootCmd, o)
	workspace.AddCommand(rootCmd, o)
	netrule.AddCommand(rootCmd, o)

	create.AddCommand(rootCmd, o)
	get.AddCommand(rootCmd, o)
	del.AddCommand(rootCmd, o)
	run.AddCommand(rootCmd, o)
	stop.AddCommand(rootCmd, o)

	return rootCmd
}

func Execute() {
	o := cmdutil.NewCliOptions()
	o.In = os.Stdin
	o.Out = os.Stdout
	o.ErrOut = os.Stderr
	rootCmd := NewRootCmd(o)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(o.Out, err)
		os.Exit(1)
	}

}
