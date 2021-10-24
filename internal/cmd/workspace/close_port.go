package workspace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/equality"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
	"github.com/cosmo-workspace/cosmo/pkg/wsnet"
)

type closePortOption struct {
	*cmdutil.UserNamespacedCliOptions

	WorkspaceName string
	PortName      string
}

func closePortCmd(cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &closePortOption{UserNamespacedCliOptions: cliOpt}

	cmd := &cobra.Command{
		Use:               "close-port WORKSPACE_NAME --port-name PORT_NAME",
		Short:             "Remove workspace network port",
		PersistentPreRunE: o.PreRunE,
		RunE:              o.RunE,
	}
	cmd.Flags().StringVar(&o.PortName, "port-name", "", "port name (Required)")
	return cmd
}

func (o *closePortOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *closePortOption) Validate(cmd *cobra.Command, args []string) error {
	if o.AllNamespace {
		return errors.New("--all-namespace is not supported in this command")
	}
	if err := o.UserNamespacedCliOptions.Validate(cmd, args); err != nil {
		return err
	}
	if len(args) < 1 {
		return errors.New("invalid args")
	}
	if o.PortName == "" {
		return errors.New("port name is required")
	}
	return nil
}

func (o *closePortOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.WorkspaceName = args[0]
	return nil
}

func (o *closePortOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client
	if _, err := c.GetUser(ctx, o.User); err != nil {
		return err
	}

	ws, err := c.GetWorkspace(ctx, o.WorkspaceName, o.Namespace)
	if err != nil {
		return err
	}
	o.Logr.DebugAll().Info("GetWorkspace", "ws", ws, "namespace", o.Namespace)

	before := ws.DeepCopy()

	if o.PortName == ws.Status.Config.ServiceMainPortName {
		return errors.New("main port cannot be removed")
	}

	var delRule *wsv1alpha1.NetworkRule
	for _, v := range ws.Spec.Network {
		if v.PortName == o.PortName {
			delRule = v.DeepCopy()
		}
	}
	if delRule == nil {
		return fmt.Errorf("port name %s is not found", o.PortName)
	}

	ws.Spec.Network = wsnet.RemoveNetworkOverrideByName(ws.Spec.Network, *delRule)
	o.Logr.DebugAll().Info("NetworkRule removed", "ws", ws, "namespace", o.Namespace, "portName", o.PortName)

	o.Logr.Debug().PrintObjectDiff(before, ws)
	if equality.Semantic.DeepEqual(before, ws) {
		return errors.New("no change")
	}

	if err = c.Update(ctx, ws); err != nil {
		return err
	}

	cmdutil.PrintfColorInfo(o.Out, "Successfully closed port '%s' for workspace '%s'\n", o.PortName, o.WorkspaceName)
	return nil
}
