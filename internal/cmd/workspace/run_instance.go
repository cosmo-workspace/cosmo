package workspace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/api/equality"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type runInstanceOption struct {
	*cmdutil.UserNamespacedCliOptions

	InstanceName string
}

func runInstanceCmd(cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &runInstanceOption{UserNamespacedCliOptions: cliOpt}

	cmd := &cobra.Command{
		Use:               "run-instance WORKSPACE_NAME",
		Aliases:           []string{"run"},
		Short:             "Run workspace instance",
		PersistentPreRunE: o.PreRunE,
		RunE:              o.RunE,
	}
	return cmd
}

func (o *runInstanceOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *runInstanceOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	return nil
}

func (o *runInstanceOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.InstanceName = args[0]
	return nil
}

func (o *runInstanceOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	if _, err := c.GetUser(ctx, o.User); err != nil {
		return err
	}

	ws, err := c.GetWorkspace(ctx, o.InstanceName, o.Namespace)
	if err != nil {
		return err
	}
	o.Logr.DebugAll().Info("GetWorkspace", "ws", ws, "namespace", o.Namespace)

	before := ws.DeepCopy()

	var rep int64 = 1
	ws.Spec.Replicas = &rep

	o.Logr.Debug().PrintObjectDiff(before, ws)
	if equality.Semantic.DeepEqual(before, ws) {
		return errors.New("no change")
	}

	if err = c.Update(ctx, ws); err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully run workspace %s\n", o.InstanceName)
	return nil
}
