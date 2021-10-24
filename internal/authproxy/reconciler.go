package authproxy

import (
	"context"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/internal/authproxy/proxy"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

const (
	EnvInstance  = "COSMO_AUTH_PROXY_INSTANCE"
	EnvNamespace = "COSMO_AUTH_PROXY_NAMESPACE"

	authProxyFieldManager = "cosmo-auth-proxy"
)

// NetworkRuleReconciler reconciles the Instance network override for my own Instance.
type NetworkRuleReconciler struct {
	kosmo.Client
	Recorder     record.EventRecorder
	Scheme       *runtime.Scheme
	ProxyManager *proxy.Manager

	WorkspaceName string

	lock sync.Mutex
}

func (r *NetworkRuleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	log := clog.FromContext(ctx).WithName("NetworkRuleReconciler")

	var ws wsv1alpha1.Workspace
	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, ignoreNotFound(err)
	}

	if ws.Name != r.WorkspaceName {
		// not myself
		return ctrl.Result{}, nil
	}

	before := ws.DeepCopy()

	if len(ws.Spec.Network) == 0 {
		// no network
		return ctrl.Result{}, nil
	}

	usingProxyList := make([]string, 0)

	for i, netRule := range ws.Spec.Network {
		if netRule.Public {
			continue
		}

		// if service port == service target port, create auth proxy and update target port as proxy port
		usingProxyList = append(usingProxyList, netRule.PortName)

		existingProxyPort, existingProxyTargetPort, exist := r.ProxyManager.GetRunningProxy(netRule.PortName)
		if exist {
			if existingProxyTargetPort == netRule.PortNumber {
				// if target port is expected, update netSpec properly
				log.Debug().Info("proxy already running",
					"name", netRule.PortName, "portNumber", netRule.PortNumber, "existingProxyPort", existingProxyPort, "existingProxyTargetPort", existingProxyTargetPort)

				ws.Spec.Network[i].TargetPortNumber = pointer.Int32(int32(existingProxyPort))
				continue

			}

			// target port is different, recreate proxy
			log.Debug().Info("proxy listening different port, shutdown",
				"name", netRule.PortName, "portNumber", netRule.PortNumber, "existingProxyPort", existingProxyPort, "existingProxyTargetPort", existingProxyTargetPort)

			err := r.ProxyManager.ShutdownProxy(ctx, netRule.PortName)
			if err != nil {
				log.Error(err, "error in shotdown proxy",
					"name", netRule.PortName, "portNumber", netRule.PortNumber, "existingProxyPort", existingProxyPort, "existingProxyTargetPort", existingProxyTargetPort)
			} else {
				r.Recorder.Eventf(&ws, corev1.EventTypeNormal, "Proxy removed", "successfully shotdown proxy: name=%s portNumber=%d proxyPort=%d",
					netRule.PortName, netRule.PortNumber, existingProxyPort, existingProxyTargetPort)
			}
		}

		log.Info("creating new proxy", "name", netRule.PortName, "targetPort", netRule.TargetPortNumber)

		proxyCreateCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		proxyPort, err := r.ProxyManager.CreateNewProxy(proxyCreateCtx, netRule.PortName, netRule.PortNumber)
		cancel()
		if err != nil {
			r.Recorder.Eventf(&ws, corev1.EventTypeWarning, "Proxy create failed", "failed to create new proxy: %s proxyPort: %d targetPort: %d %v", netRule.PortName, proxyPort, netRule.PortNumber, err.Error())
			continue
		}

		r.Recorder.Eventf(&ws, corev1.EventTypeNormal, "Proxy created", "successfully created new proxy: name=%s portNumber=%d proxyPort=%d",
			netRule.PortName, netRule.PortNumber, proxyPort)
	}

	// Update Workspace
	if !equality.Semantic.DeepEqual(*before, ws) {
		err := r.Update(ctx, &ws)
		if err != nil {
			log.Error(err, "failed to apply Instance")
			return ctrl.Result{}, err
		}

		log.Info("updated")
		log.PrintObjectDiff(*before, ws)
	}

	// Shutdown unused proxy
	r.ProxyManager.GC(ctx, usingProxyList)

	return ctrl.Result{}, nil
}

func (r *NetworkRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// only watch my workspace
	predi := predicate.NewPredicateFuncs(func(object client.Object) bool {
		return object.GetName() == r.WorkspaceName
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(&wsv1alpha1.Workspace{}).
		WithEventFilter(predi).
		Complete(r)
}

// ignoreNotFound return nil if the given err is NotFoundErr.
func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}
