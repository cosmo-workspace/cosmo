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
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/yaml"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/cmdutil"
)

type getOption struct {
	*cmdutil.UserNamespacedCliOptions

	WorkspaceName string

	outputFormat string
	showNetwork  bool
}

func getCmd(cliOpt *cmdutil.UserNamespacedCliOptions) *cobra.Command {
	o := &getOption{UserNamespacedCliOptions: cliOpt}
	cmd := &cobra.Command{
		Use:   "get [WORKSPACE_NAME]",
		Short: "Get workspaces",
		Long: `
Get workspaces

This command is like "kubectl get workspace" but show more information.

But for Workspace detailed status or trouble shooting, 
use "kubectl describe workspace" or "kubectl describe instance" and see controller's events.
`,
		PersistentPreRunE: o.PreRunE,
		RunE:              cmdutil.RunEHandler(o.RunE),
	}
	cmd.Flags().StringVarP(&o.outputFormat, "output", "o", "", "output format. available: 'wide', 'yaml'")
	cmd.Flags().BoolVar(&o.showNetwork, "network", false, "show workspace network")
	return cmd
}

func (o *getOption) PreRunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}
	return nil
}

func (o *getOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.UserNamespacedCliOptions.Validate(cmd, args); err != nil {
		return err
	}
	switch o.outputFormat {
	case "wide", "yaml":
	case "":
	default:
		return fmt.Errorf("invalid output format: available formats is ['wide', 'yaml']")
	}
	return nil
}

func (o *getOption) Complete(cmd *cobra.Command, args []string) error {
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

func (o *getOption) RunE(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.Client

	var wss []wsv1alpha1.Workspace

	o.Logr.Debug().Info("options", "namespace", o.Namespace, "all-namespaces", o.AllNamespace, "workspaceName", o.WorkspaceName)

	if o.AllNamespace {
		users, err := c.ListUsers(ctx)
		if err != nil {
			// todo:
			err = cmdutil.UnwrapKosmoError(err)
			return fmt.Errorf("failed to list users: %v", err)
		}
		o.Logr.DebugAll().Info("ListUsers", "users", users)

		for _, user := range users {
			ws, err := c.ListWorkspacesByUserID(ctx, user.Name)
			if err != nil {
				// todo:
				err = cmdutil.UnwrapKosmoError(err)
				return fmt.Errorf("failed to list workspaces: %v", err)
			}
			o.Logr.DebugAll().Info("ListWorkspacesByUserID", "user", o.User, "wsCount", len(ws), "wsList", ws)
			wss = append(wss, ws...)
		}

	} else if o.WorkspaceName != "" {
		ws, err := c.GetWorkspaceByUserID(ctx, o.WorkspaceName, o.User)
		if err != nil {
			// todo:
			err = cmdutil.UnwrapKosmoError(err)
			return fmt.Errorf("failed to get workspace: %v", err)
		}
		wss = []wsv1alpha1.Workspace{*ws}
		o.Logr.DebugAll().Info("GetWorkspaceByUserID", "user", o.User, "ws", ws)

	} else {
		_, err := c.GetUser(ctx, o.User)
		if err != nil {
			// todo:
			err = cmdutil.UnwrapKosmoError(err)
			return fmt.Errorf("failed to get user: %v", err)
		}

		wss, err = c.ListWorkspacesByUserID(ctx, o.User)
		if err != nil {
			// todo:
			err = cmdutil.UnwrapKosmoError(err)
			return fmt.Errorf("failed to list workspaces: %v", err)
		}
		o.Logr.DebugAll().Info("ListWorkspacesByUserID", "user", o.User, "wsCount", len(wss), "wsList", wss)
	}

	if o.outputFormat == "yaml" {
		raw := make([]byte, 0, len(wss))
		for _, ws := range wss {
			v := ws.DeepCopy()
			gvk, err := apiutil.GVKForObject(v, o.Scheme)
			if err != nil {
				return err
			}
			v.SetGroupVersionKind(gvk)
			rawObj, err := yaml.Marshal(v)
			if err != nil {
				o.Logr.Error(err, "failed to marshal yaml", "workspace", v.Name)
				continue
			}
			raw = append(raw, rawObj...)
			raw = append(raw, []byte("---\n")...)
		}
		fmt.Fprintln(o.Out, string(raw))
		return nil
	}

	w := printers.GetNewTabWriter(o.Out)
	defer w.Flush()

	if o.showNetwork {
		columnNames := []string{"USER-NAMESPACE", "WORKSPACE-NAME", "PORT-NAME", "PORT-NUMBER", "GROUP", "HTTP-PATH", "URL"}
		fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))

		for _, ws := range wss {
			for _, v := range ws.Spec.Network {
				url := ws.Status.URLs[v.PortName]
				rowdata := []string{ws.Namespace, ws.Name, v.PortName, strconv.Itoa(int(v.PortNumber)), *v.Group, v.HTTPPath, url}
				fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
			}
		}

	} else {
		columnNames := []string{"USER-NAMESPACE", "NAME", "TEMPLATE", "POD-PHASE"}
		if o.outputFormat == "wide" {
			columnNames = append(columnNames, "URLS")
		}
		fmt.Fprintf(w, "%s\n", strings.Join(columnNames, "\t"))

		for _, ws := range wss {
			rowdata := []string{ws.Namespace, ws.Name, ws.Spec.Template.Name, string(ws.Status.Phase)}
			if o.outputFormat == "wide" {
				rowdata = append(rowdata, fmt.Sprintf("%s", ws.Status.URLs))
			}
			fmt.Fprintf(w, "%s\n", strings.Join(rowdata, "\t"))
		}
	}
	return nil
}
