package controllers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/wsnet"
)

// WorkspaceStatusReconciler reconciles a Workspace object
type WorkspaceStatusReconciler struct {
	client.Client
	Recorder       record.EventRecorder
	Scheme         *runtime.Scheme
	DefaultURLBase string
}

// +kubebuilder:rbac:groups=workspace.cosmo.cosmo-workspace.github.io,resources=workspaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=workspace.cosmo.cosmo-workspace.github.io,resources=workspaces/status,verbs=get;update;patch
func (r *WorkspaceStatusReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("WorkspaceStatusReconciler").WithValues("req", req)

	log.Debug().Info("start reconcile")

	key := r.getWorkspaceNamespacedName(ctx, req.NamespacedName)

	var ws wsv1alpha1.Workspace
	if err := r.Get(ctx, key, &ws); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log = log.WithValues("UID", ws.UID, "Template", ws.Spec.Template.Name)
	ctx = clog.IntoContext(ctx, log)

	current := ws.DeepCopy()

	log.DumpObject(r.Scheme, &ws, "before workspace")

	// sync workspace config with template
	cfg, err := getWorkspaceConfig(ctx, r.Client, ws.Spec.Template.Name)
	if err != nil {
		return ctrl.Result{}, err
	}
	if cfg.URLBase == "" {
		cfg.URLBase = r.DefaultURLBase
	}
	ws.Status.Config = cfg

	// fetch child pod status
	if urlMap, err := r.GenWorkspaceURLMap(ctx, ws); err == nil {
		log.Debug().Info(fmt.Sprintf("workspace urlmap: %s", urlMap))
		ws.Status.URLs = urlMap

	} else {
		log.Info("failed to gen urlmap", "error", err, "ws", ws.Name, "urlbase", cfg.URLBase, "logLevel", "warn")
	}

	// set workspace phase
	requeue := false
	pods, err := listWorkspacePods(ctx, r.Client, ws)
	if err != nil {
		log.Info("failed to list instance pods", "error", err, "ws", ws.Name, "logLevel", "warn")
		ws.Status.Phase = "NotRunning"

	} else {
		if len(pods) > 0 {
			if *ws.Spec.Replicas == 0 {
				ws.Status.Phase = "Stopping"
				requeue = true
			} else {
				requeue = true
				for _, pod := range pods {
					ws.Status.Phase = kubeutil.PodStatusReason(pod)
					if ws.Status.Phase == "Running" {
						requeue = false
						break
					}
				}
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
	if !equality.Semantic.DeepEqual(current, &ws) {
		log.Debug().PrintObjectDiff(current, &ws)

		if err := r.Status().Update(ctx, &ws); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("workspace status updated", "ws", ws.Name)
	}

	log.Debug().Info("finish reconcile")
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

	svc, _, err := getWorkspaceServicesAndIngress(ctx, r.Client, ws)
	if err != nil {
		return nil, err
	}

	urlvarsMap := make(map[string]wsnet.URLVars)
	for _, netRule := range ws.Spec.Network {
		urlvars := wsnet.URLVars{}
		urlvars.NetworkRuleName = netRule.Name
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
			if p.Name == netRule.Name {
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

		urlvarsMap[netRule.Name] = urlvars
	}

	urlMap := make(map[string]string)
	for name, urlvars := range urlvarsMap {
		log.DebugAll().Info("urlvar map", urlvars.Dump()...)
		urlMap[name] = urlbase.GenURL(urlvars)
	}

	return urlMap, nil
}

func listWorkspacePods(ctx context.Context, c client.Client, ws wsv1alpha1.Workspace) ([]corev1.Pod, error) {
	var podList corev1.PodList

	ls := labels.NewSelector()
	req, _ := labels.NewRequirement(cosmov1alpha1.LabelKeyInstance, selection.Equals, []string{ws.GetName()})
	ls = ls.Add(*req)

	opts := &client.ListOptions{
		LabelSelector: ls,
		Namespace:     ws.GetNamespace(),
	}
	if err := c.List(ctx, &podList, opts); err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func getWorkspaceServicesAndIngress(ctx context.Context, c client.Client, ws wsv1alpha1.Workspace) (svc corev1.Service, ing netv1.Ingress, err error) {
	var svcList corev1.ServiceList
	var ingList netv1.IngressList

	ls := labels.NewSelector()
	req, _ := labels.NewRequirement(cosmov1alpha1.LabelKeyInstance, selection.In, []string{ws.GetName()})
	ls = ls.Add(*req)

	opts := &client.ListOptions{
		LabelSelector: ls,
		Namespace:     ws.GetNamespace(),
	}

	if err := c.List(ctx, &svcList, opts); err != nil {
		return svc, ing, err
	}

	if len(svcList.Items) == 0 {
		return svc, ing, errors.New("no services")
	}

	for _, v := range svcList.Items {
		if instance.EqualInstanceResourceName(ws.GetName(), v.Name, ws.Status.Config.ServiceName) {
			svc = v
		}
	}

	if err := c.List(ctx, &ingList, opts); err != nil {
		return svc, ing, err
	}

	for _, v := range ingList.Items {
		if instance.EqualInstanceResourceName(ws.GetName(), v.Name, ws.Status.Config.IngressName) {
			ing = v
		}
	}

	return svc, ing, nil
}
