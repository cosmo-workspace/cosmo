package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/internal/cmd/template"
	"github.com/cosmo-workspace/cosmo/internal/cmd/user"
	"github.com/cosmo-workspace/cosmo/internal/cmd/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "cosmoctl",
		Short: "Command line tool to manipulate comso",
		Long: `
Command line tool to manipulate comso
Complete documentation is available at http://github.com/cosmo-workspace/cosmo

MIT 2021 cosmo-workspace/cosmo
`,
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("cosmoctl - cosmo v0.2.1 cosmo-workspace 2021")
		},
	}
	rootCmd.AddCommand(versionCmd)

	o := cmdutil.NewCliOptions()
	o.Out = os.Stdout
	o.ErrOut = os.Stderr

	rootCmd.PersistentFlags().StringVar(&o.KubeConfigPath, "kubeconfig", "", "kubeconfig file path (default: $HOME/.kube/config)")
	rootCmd.PersistentFlags().StringVar(&o.KubeContext, "context", "", "kube-context (default: current context)")
	rootCmd.PersistentFlags().IntVar(&o.LogLevel, "v", -1, "log level. -1:DISABLED, 0:INFO, 1:DEBUG, 2:ALL")

	template.AddCommand(rootCmd, o)
	user.AddCommand(rootCmd, o)
	workspace.AddCommand(rootCmd, o)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
