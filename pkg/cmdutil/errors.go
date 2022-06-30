package cmdutil

import (
	"fmt"

	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/spf13/cobra"
)

func UnwrapKosmoError(err error) error {
	if ke, ok := err.(*kosmo.KosmoError); ok {
		e := ke.Unwrap()
		if e != nil {
			return fmt.Errorf("%s: %s", err.Error(), e.Error())
		}
	}
	return err
}

func RunEHandler(runE func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := runE(cmd, args)
		return UnwrapKosmoError(err)
	}
}
