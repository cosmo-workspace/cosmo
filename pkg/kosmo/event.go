package kosmo

import (
	"context"
	"fmt"
	"slices"
	"time"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
)

func (c *Client) ListEvents(ctx context.Context, namespace string) ([]eventsv1.Event, error) {
	log := clog.FromContext(ctx).WithCaller()
	var events eventsv1.EventList
	if err := c.List(ctx, &events, &client.ListOptions{Namespace: namespace}); err != nil {
		log.Error(err, "failed to list Event", "namespace", namespace)
		return nil, apierrs.NewInternalError(fmt.Errorf("failed to list Event: %w", err))
	}
	for _, v := range events.Items {
		err := c.UpdateEventAnnotations(ctx, v.DeepCopy())
		if err != nil {
			log.Debug().Info("failed to cache instance name on event annotation", "error", err, "namespace", namespace, "event", v.Name)
		}
	}

	sorted(events.Items)
	return events.Items, nil
}

func (c *Client) UpdateEventAnnotations(ctx context.Context, event *eventsv1.Event) error {
	log := clog.FromContext(ctx).WithCaller()

	ann := event.GetAnnotations()
	if ann != nil {
		instName := ann[cosmov1alpha1.EventAnnKeyInstanceName]
		userName := ann[cosmov1alpha1.EventAnnKeyUserName]
		if instName != "" && userName != "" {
			// annotation is expected
			return nil
		}
	} else {
		ann = map[string]string{}
	}

	// fetch regarding object to get instance name on label
	obj := unstructured.Unstructured{}
	obj.SetAPIVersion(event.Regarding.APIVersion)
	obj.SetKind(event.Regarding.Kind)
	obj.SetName(event.Regarding.Name)
	obj.SetNamespace(event.Regarding.Namespace)

	err := c.Get(ctx, client.ObjectKeyFromObject(&obj), &obj)
	if err != nil {
		return fmt.Errorf("failed to fetch regarding object: %w", err)
	}

	ann[cosmov1alpha1.EventAnnKeyInstanceName] = kubeutil.GetLabel(&obj, cosmov1alpha1.LabelKeyInstanceName)
	ann[cosmov1alpha1.EventAnnKeyUserName] = cosmov1alpha1.UserNameByNamespace(event.Namespace)
	event.SetAnnotations(ann)
	if err := c.Update(ctx, event); err != nil {
		return fmt.Errorf("failed to cache on event annotation: %w", err)
	}

	log.Info("cached on event annotation", "event", event)
	return nil
}

// sorted sort given events
// rules:
//  1. event time
//  2. last event time
//  3. regardings:
//     - 1. user
//     - 2. workspace
//     - 3. instance
//     - 4. deployment
//     - 5. ingressroute
//     - 6. service
func sorted(events []eventsv1.Event) {
	slices.SortStableFunc(events, func(a, b eventsv1.Event) int {
		if res := sortFuncEventTime(a, b); res != 0 {
			return res
		}
		if res := sortFuncLastEventTime(a, b); res != 0 {
			return res
		}
		if res := sortFuncRegardings(a, b); res != 0 {
			return res
		}
		return 0
	})
}

func regardingPriority(ref corev1.ObjectReference) int {
	switch ref.APIVersion {
	case "cosmo-workspace.github.io/v1alpha1", "cosmo-workspace.github.io/v1":
		switch ref.Kind {
		case "User":
			return 0
		case "Workspace":
			return 1
		case "Instance":
			return 2
		default:
			return 3
		}

	case "traefik.io/v1alpha1", "traefik.io/v1":
		return 10
	case "apps/v1":
		if ref.Kind == "Deployment" {
			return 20
		}
		return 30

	default:
		if ref.Kind == "Service" {
			return 90
		}
		return 100
	}
}

func sortFuncEventTime(a, b eventsv1.Event) int {
	if eventTime(a).Before(eventTime(b)) {
		return -1
	} else if eventTime(a).After(eventTime(b)) {
		return 1
	} else {
		return 0
	}
}

func sortFuncLastEventTime(a, b eventsv1.Event) int {
	if lastTime(a).Before(lastTime(b)) {
		return -1
	} else if lastTime(a).After(lastTime(b)) {
		return 1
	} else {
		return 0
	}
}

func sortFuncRegardings(a, b eventsv1.Event) int {
	return regardingPriority(a.Regarding) - regardingPriority(b.Regarding)
}

func eventTime(v eventsv1.Event) time.Time {
	if v.EventTime.Year() != 1 {
		return v.EventTime.Time
	} else {
		return v.DeprecatedLastTimestamp.Time
	}
}

func lastTime(v eventsv1.Event) time.Time {
	if v.Series != nil {
		return v.Series.LastObservedTime.Time
	} else {
		return v.DeprecatedLastTimestamp.Time
	}
}

func UserEventf(rec record.EventRecorder, user *cosmov1alpha1.User, eventType, reason, messageFmt string, args ...interface{}) {
	user.SetNamespace(cosmov1alpha1.UserNamespace(user.Name))
	ann := map[string]string{
		cosmov1alpha1.EventAnnKeyUserName: user.Name,
	}
	rec.AnnotatedEventf(user, ann, eventType, reason, messageFmt, args...)
}

func WorkspaceEventf(rec record.EventRecorder, ws *cosmov1alpha1.Workspace, eventType, reason, messageFmt string, args ...interface{}) {
	ann := map[string]string{
		cosmov1alpha1.EventAnnKeyUserName:     cosmov1alpha1.UserNameByNamespace(ws.Namespace),
		cosmov1alpha1.EventAnnKeyInstanceName: ws.Name,
	}
	rec.AnnotatedEventf(ws, ann, eventType, reason, messageFmt, args...)
}

func InstanceEventf(rec record.EventRecorder, inst cosmov1alpha1.InstanceObject, eventType, reason, messageFmt string, args ...interface{}) {
	ann := map[string]string{
		cosmov1alpha1.EventAnnKeyInstanceName: inst.GetName(),
	}

	namespace := inst.GetNamespace()
	if namespace != "" {
		ann[cosmov1alpha1.EventAnnKeyUserName] = cosmov1alpha1.UserNameByNamespace(namespace)
	} else {
		user := userFromOwnerReferences(inst.GetOwnerReferences())
		if user != "" {
			inst.SetNamespace(cosmov1alpha1.UserNamespace(user))
			ann[cosmov1alpha1.EventAnnKeyUserName] = user
		}
	}

	rec.AnnotatedEventf(inst, ann, eventType, reason, messageFmt, args...)
}

func userFromOwnerReferences(refs []metav1.OwnerReference) string {
	for _, ref := range refs {
		if ref.Kind == "User" {
			return ref.Name
		}
	}
	return ""
}
