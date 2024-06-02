package user

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

type GetEventsOption struct {
	*cli.RootOptions
	UserName string
}

func GetEventsCmd(cmd *cobra.Command, opt *cli.RootOptions) *cobra.Command {
	o := &GetEventsOption{RootOptions: opt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	return cmd
}

func (o *GetEventsOption) Validate(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	if o.UseKubeAPI && len(args) < 1 {
		return fmt.Errorf("user name is required")
	}
	return nil
}

func (o *GetEventsOption) Complete(cmd *cobra.Command, args []string) error {
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}
	if len(args) > 0 {
		o.UserName = args[0]
	}
	if !o.UseKubeAPI && o.UserName == "" {
		o.UserName = o.CliConfig.User
	}

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return nil
}

func (o *GetEventsOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*30)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	var (
		events []*dashv1alpha1.Event
		err    error
	)
	if o.UseKubeAPI {
		events, err = o.GetEventsByKubeClient(ctx)
		if err != nil {
			return err
		}
	} else {
		events, err = o.GetEventsWithDashClient(ctx)
		if err != nil {
			return err
		}
	}
	o.Logr.Debug().Info("Events", "events", events)

	o.OutputTable(cmd.OutOrStdout(), events)
	return nil
}

func (o *GetEventsOption) GetEventsWithDashClient(ctx context.Context) ([]*dashv1alpha1.Event, error) {
	req := &dashv1alpha1.GetEventsRequest{
		UserName: o.UserName,
	}
	c := o.CosmoDashClient
	res, err := c.UserServiceClient.GetEvents(ctx, cli.NewRequestWithToken(req, o.CliConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	o.Logr.DebugAll().Info("UserServiceClient.GetUser", "res", res)
	return res.Msg.Items, nil
}

func (o *GetEventsOption) OutputTable(w io.Writer, events []*dashv1alpha1.Event) {
	data := [][]string{}

	for _, v := range events {
		data = append(data, []string{lastSeen(v.Series), v.Type, v.Reason, regarding(v.Regarding), v.ReportingController, v.Note})
	}
	cli.OutputTable(w,
		[]string{"LAST SEEN", "TYPE", "REASON", "OBJECT", "REPORTER", "MESSAGE"},
		data)
}

func lastSeen(series *dashv1alpha1.EventSeries) string {
	if series.Count > 0 {
		return fmt.Sprintf("%s (%vx)", time.Since(series.LastObservedTime.AsTime()).Round(time.Second), series.Count)
	}
	return time.Since(series.LastObservedTime.AsTime()).Round(time.Second).String()
}

func regarding(v *dashv1alpha1.ObjectReference) string {
	return fmt.Sprintf("%s/%s", v.Kind, v.Name)
}

func (o *GetEventsOption) GetEventsByKubeClient(ctx context.Context) ([]*dashv1alpha1.Event, error) {
	c := o.KosmoClient
	events, err := c.ListEvents(ctx, v1alpha1.UserNamespace(o.UserName))
	if err != nil {
		return nil, err
	}
	o.Logr.Debug().Info("ListEvents", "events", events)
	return apiconv.K2D_Events(events), nil
}
