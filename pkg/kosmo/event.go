package kosmo

import (
	"context"
	"fmt"

	eventsv1 "k8s.io/api/events/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

func (c *Client) ListEvents(ctx context.Context, namespace string) ([]eventsv1.Event, error) {
	log := clog.FromContext(ctx).WithCaller()
	var events eventsv1.EventList
	if err := c.List(ctx, &events, &client.ListOptions{Namespace: namespace}); err != nil {
		log.Error(err, "failed to list Event", "namespace", namespace)
		return nil, apierrs.NewInternalError(fmt.Errorf("failed to list Event: %w", err))
	} else {
		return events.Items, nil
	}
}
