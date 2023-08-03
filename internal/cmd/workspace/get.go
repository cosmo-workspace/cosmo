package workspace

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/printers"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type GetOption struct {
	*cmdutil.UserNamespacedCliOptions

	WorkspaceName string

	showNetwork bool
}

func GetCmd(cmd *cobra.Command, cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &GetOption{UserNamespacedCliOptions: cliOpt}

	cmd.PersistentPreRunE = o.PreRunE
	cmd.RunE = cmdutil.RunEHandler(o.RunE)

	cmd.Flags().StringVarP(&o.User, "user", "u", "", "user name")
	cmd.Flags().StringVarP(&o.Namespace, "namespace", "n", "", "namespace")
	cmd.Flags().BoolVarP(&o.AllNamespace, "all-namespaces", "A", false, "all namespaces")

	cmd.Flags().BoolVar(&o.showNetwork, "network", false, "show workspace network")
	return cmd
}

func (o *GetOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *GetOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Validate(cmd, args); err != nil {
		return err
	}
	return nil
}

func (o *GetOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Complete(cmd, args); err != nil {
		return err
	}
	if len(args) > 0 {
		o.WorkspaceName = args[0]
		if o.AllNamespace {
			return errors.New("--all-namespaces is not allowed to use if WORKSPACE_NAME specified")
		}
	}
	return nil
}

func (o *GetOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	var wss []cosmov1alpha1.Workspace

	o.Logr.Debug().Info("options", "namespace", o.Namespace, "all-namespaces", o.AllNamespace, "workspaceName", o.WorkspaceName)

	if o.AllNamespace {
		users, err := c.ListUsers(ctx)
		if err != nil {
			return err
		}
		o.Logr.DebugAll().Info("ListUsers", "users", users)

		for _, user := range users {
			ws, err := c.ListWorkspacesByUserName(ctx, user.Name)
			if err != nil {
				return err
			}
			o.Logr.DebugAll().Info("ListWorkspacesByUserName", "user", o.User, "wsCount", len(ws), "wsList", ws)
			wss = append(wss, ws...)
		}

	} else if o.WorkspaceName != "" {
		ws, err := c.GetWorkspaceByUserName(ctx, o.WorkspaceName, o.User)
		if err != nil {
			return err
		}
		wss = []cosmov1alpha1.Workspace{*ws}
		o.Logr.DebugAll().Info("GetWorkspaceByUserName", "user", o.User, "ws", ws)

	} else {
		_, err := c.GetUser(ctx, o.User)
		if err != nil {
			return err
		}

		wss, err = c.ListWorkspacesByUserName(ctx, o.User)
		if err != nil {
			return err
		}
		o.Logr.DebugAll().Info("ListWorkspacesByUserName", "user", o.User, "wsCount", len(wss), "wsList", wss)
	}

	w := printers.GetNewTabWriter(o.Out)
	defer w.Flush()

	if o.showNetwork {
		columnNames := []string{"USER", "NAME", "PORT", "URL", "PUBLIC"}
		fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))

		for _, ws := range wss {
			for _, v := range ws.Spec.Network {
				url := ws.Status.URLs[v.UniqueKey()]
				rowdata := []string{cosmov1alpha1.UserNameByNamespace(ws.Namespace), ws.Name, strconv.Itoa(int(v.PortNumber)), url, strconv.FormatBool(v.Public)}
				fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
			}
		}

	} else {
		columnNames := []string{"USER", "NAME", "TEMPLATE", "PHASE"}
		fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))

		for _, ws := range wss {
			rowdata := []string{cosmov1alpha1.UserNameByNamespace(ws.Namespace), ws.Name, ws.Spec.Template.Name, string(ws.Status.Phase)}
			fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
		}
	}
	return nil
}
