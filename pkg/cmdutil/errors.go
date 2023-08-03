package cmdutil

import (
	"github.com/spf13/cobra"
)

func RunEHandler(runE func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := runE(cmd, args)
		return err
	}
}
