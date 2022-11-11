package controllers

import (
	"context"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/wsnet"
)

const (
	WsStatControllerFieldManager string = "cosmo-workspace-status-controller"
)

// WorkspaceStatusReconciler reconciles a Workspace object
type WorkspaceStatusReconciler struct {
	kosmo.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=workspace.cosmo.cosmo-workspace.github.io,resources=workspaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=workspace.cosmo.cosmo-workspace.github.io,resources=workspaces/status,verbs=get;update;patch
func (r *WorkspaceStatusReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("WorkspaceStatusReconciler")
	ctx = clog.IntoContext(ctx, log)

	log.Debug().Info("start reconcile")
	requeue := false

	key := r.getWorkspaceNamespacedName(ctx, req.NamespacedName)

	var ws wsv1alpha1.Workspace
	if err := r.Get(ctx, key, &ws); err != nil {
		return ctrl.Result{}, ignoreNotFound(err)
	}

	current := ws.DeepCopy()

	log.DebugAll().DumpObject(r.Scheme, &ws, "before workspace")

	// sync workspace config with template
	cfg, err := r.GetWorkspaceConfig(ctx, ws.Spec.Template.Name)
	if err != nil {
		return ctrl.Result{}, err
	}
	ws.Status.Config = cfg

	// fetch child pod status
	if urlMap, err := r.GenWorkspaceURLMap(ctx, ws); err == nil {
		log.DebugAll().Info("workspace urlmap", "urlmap", urlMap)
		ws.Status.URLs = urlMap

	} else {
		log.Info("failed to gen urlmap", "error", err, "ws", ws.Name, "urlbase", cfg.URLBase, "logLevel", "warn")
	}

	// set workspace phase
	pods, err := r.ListWorkspacePods(ctx, ws)
	if err != nil {
		log.Info("failed to list instance pods", "error", err, "ws", ws.Name, "logLevel", "warn")
		ws.Status.Phase = "NotRunning"

	} else {
		if len(pods) > 0 {
			if *ws.Spec.Replicas == 0 {
				ws.Status.Phase = "Stopping"
				requeue = true
			} else {
				ws.Status.Phase = kosmo.PodStatusReason(pods[0])
			}

		} else {
			if *ws.Spec.Replicas > 0 {
				ws.Status.Phase = "Starting"
				requeue = true
			} else {
				ws.Status.Phase = "Stopped"
			}
		}
	}

	// update workspace status
	if !equality.Semantic.DeepEqual(*current, ws) {
		log.Debug().PrintObjectDiff(*current, ws)

		if err := r.Status().Update(ctx, &ws); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("workspace status updated", "ws", ws.Name)
	}

	log.Info("finish reconcile", "requeue", requeue)
	if requeue {
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}
	return ctrl.Result{}, nil
}

func (r *WorkspaceStatusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&wsv1alpha1.Workspace{}).
		Owns(&cosmov1alpha1.Instance{}).
		Build(r)
	if err != nil {
		return err
	}

	// watch pods which has "cosmo/instance" label
	predi, _ := predicate.LabelSelectorPredicate(metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      cosmov1alpha1.LabelKeyInstance,
				Operator: metav1.LabelSelectorOpExists,
			},
		},
	})
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{}, predi)
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{}, predi)
	if err != nil {
		return err
	}
	return nil
}

func (r *WorkspaceStatusReconciler) getWorkspaceNamespacedName(ctx context.Context, req types.NamespacedName) types.NamespacedName {
	var pod corev1.Pod
	if err := r.Get(ctx, req, &pod); err == nil {
		// request is Pod with "cosmo/instance" label
		return types.NamespacedName{Name: pod.Labels[cosmov1alpha1.LabelKeyInstance], Namespace: pod.GetNamespace()}
	}

	var svc corev1.Service
	if err := r.Get(ctx, req, &svc); err == nil {
		// request is Service with "cosmo/instance" label
		return types.NamespacedName{Name: svc.Labels[cosmov1alpha1.LabelKeyInstance], Namespace: svc.GetNamespace()}
	}

	// request is Workspace
	return req
}

func (r *WorkspaceStatusReconciler) GenWorkspaceURLMap(ctx context.Context, ws wsv1alpha1.Workspace) (map[string]string, error) {
	log := clog.FromContext(ctx).WithCaller()
	urlbase := wsnet.URLBase(ws.Status.Config.URLBase)

	svc, _, err := r.GetWorkspaceServicesAndIngress(ctx, ws)
	if err != nil {
		return nil, err
	}

	urlvarsMap := make(map[string]wsnet.URLVars)
	for _, netRule := range ws.Spec.Network {
		urlvars := wsnet.URLVars{}
		urlvars.PortName = netRule.PortName
		urlvars.PortNumber = strconv.Itoa(netRule.PortNumber)
		if netRule.Group != nil {
			urlvars.NetRuleGroup = *netRule.Group
		}

		urlvars.InstanceName = ws.Name
		urlvars.WorkspaceName = ws.Name
		urlvars.Namespace = ws.Namespace
		urlvars.UserID = wsv1alpha1.UserIDByNamespace(ws.Namespace)

		urlvars.IngressPath = netRule.HTTPPath

		// node port
		for _, p := range svc.Spec.Ports {
			if p.Name == netRule.PortName {
				urlvars.NodePortNumber = strconv.Itoa(int(p.NodePort))
			}
		}

		// load balancer
		if svc.Status.LoadBalancer.Size() > 0 {
			lb := svc.Status.LoadBalancer.Ingress[0]
			if lb.Hostname != "" {
				urlvars.LoadBalancer = lb.Hostname
			} else {
				urlvars.LoadBalancer = lb.IP
			}
		}

		urlvarsMap[netRule.PortName] = urlvars
	}

	urlMap := make(map[string]string)
	for name, urlvars := range urlvarsMap {
		if log.DebugAll().Enabled() {
			urlvars.Dump(log)
		}
		urlMap[name] = urlbase.GenURL(urlvars)
	}

	return urlMap, nil
}
