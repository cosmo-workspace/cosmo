package dashboard

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"sync"
	"time"

	eventv1 "k8s.io/api/events/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	connect_go "github.com/bufbuild/connect-go"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/apiconv"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1/dashboardv1alpha1connect"
)

func (s *Server) StreamServiceHandler(mux *http.ServeMux) {
	path, handler := dashboardv1alpha1connect.NewStreamServiceHandler(s,
		connect_go.WithInterceptors(authorizationInterceptorFunc(s.verifyAndGetLoginUser)),
	)
	mux.Handle(path, s.contextMiddleware(handler))
}

// StreamingEvents implements dashboardv1alpha1connect.UserServiceHandler.
func (s *Server) StreamingEvents(ctx context.Context, req *connect_go.Request[dashv1alpha1.GetEventsRequest], stream *connect_go.ServerStream[dashv1alpha1.GetEventsResponse]) error {
	log := clog.FromContext(ctx).WithCaller()
	log.Info("request", "req", req)

	if err := userAuthentication(ctx, req.Msg.UserName); err != nil {
		return err
	}
	key := sha256.Sum256([]byte(stream.Conn().RequestHeader().Get("Cookie")))

	ctx, cancel := context.WithTimeout(ctx, time.Second*300)
	defer cancel()

	events, err := s.Klient.ListEvents(ctx, cosmov1alpha1.UserNamespace(req.Msg.UserName))
	if err != nil {
		return ErrResponse(log, err)
	}

	for _, v := range events {
		if req.Msg.From != nil {
			events := make([]eventv1.Event, 0)
			if _, last := apiconv.EventObservedTime(v); last.After(req.Msg.From.AsTime()) {
				events = append(events, v)
			}
			if len(events) > 0 {
				res := &dashv1alpha1.GetEventsResponse{
					Items: apiconv.K2D_Events([]eventv1.Event{v}),
				}
				if err := stream.Send(res); err != nil {
					log.Error(err, "send error")
					return err
				}
			}
		}
	}

	eventCh := s.watcher.subscribe(ctx, fmt.Sprintf("%x", key))
	if eventCh == nil {
		return fmt.Errorf("channel already exists")
	}

	for {
		select {
		case <-ctx.Done():
			log.Debug().Info("ctx done")
			return nil
		case event, ok := <-eventCh:
			if ok {
				log.Debug().Info("delegating event", "user", req.Msg.UserName)
				if event.Namespace != cosmov1alpha1.UserNamespace(req.Msg.UserName) {
					continue
				}
				res := &dashv1alpha1.GetEventsResponse{
					Items: apiconv.K2D_Events([]eventv1.Event{event}),
				}
				log.Info("sending event", "event", event)
				if err := stream.Send(res); err != nil {
					log.Error(err, "send error")
					return err
				}
			} else {
				log.Debug().Info("event channel closed")
				return nil
			}
		}
	}
}

type watcher struct {
	Klient          kosmo.Client
	Log             *clog.Logger
	subscribers     sync.Map
	cancelSubscribe sync.Map
	sendingLock     sync.Mutex
}

func (r *watcher) subscribe(ctx context.Context, key string) <-chan eventv1.Event {
	log := r.Log.WithValues("key", key)
	log.Debug().Info("create new channel")

	ctx, cancel := context.WithCancel(ctx)
	preCancel, ok := r.cancelSubscribe.Load(key)
	if ok {
		preCancel.(context.CancelFunc)()
	}

	ch := make(chan eventv1.Event)
	r.subscribers.Store(key, ch)
	r.cancelSubscribe.Store(key, cancel)

	go func(ch chan eventv1.Event) {
		log.Debug().Info("wait channel closed...")
		<-ctx.Done()
		log.Debug().Info("close channel")
		r.sendingLock.Lock()
		defer r.sendingLock.Unlock()
		r.subscribers.Delete(key)
		r.cancelSubscribe.Delete(key)
		close(ch)
	}(ch)

	return ch
}

func (r *watcher) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := r.Log.WithValues("req", req)
	log.Debug().Info("start reconcile")

	var event eventv1.Event
	if err := r.Klient.Get(ctx, req.NamespacedName, &event); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if err := r.Klient.UpdateEventAnnotations(ctx, &event); err != nil {
		log.Error(err, "failed to set regaining instance on event annotation")
	}

	r.sendingLock.Lock()
	defer r.sendingLock.Unlock()
	r.subscribers.Range(func(key, value any) bool {
		log.Debug().Info("send event to channel", "key", key, "event", event)
		ch := value.(chan eventv1.Event)
		ch <- event
		return true
	})
	log.Debug().Info("finish reconcile")
	return reconcile.Result{}, nil
}

func (r *watcher) SetupWithManager(mgr ctrl.Manager) error {
	_, err := ctrl.NewControllerManagedBy(mgr).
		For(&eventv1.Event{}).
		Build(r)
	if err != nil {
		return err
	}
	return nil
}
