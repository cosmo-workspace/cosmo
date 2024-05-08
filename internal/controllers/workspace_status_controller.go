package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
)

// WorkspaceStatusReconciler reconciles a Workspace object
type WorkspaceStatusReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=cosmo-workspace.github.io,resources=workspaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=cosmo-workspace.github.io,resources=workspaces/status,verbs=get;update;patch
func (r *WorkspaceStatusReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("WorkspaceStatusReconciler").WithValues("req", req)

	log.Debug().Info("start reconcile")

	var ws cosmov1alpha1.Workspace
	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log = log.WithValues("UID", ws.UID, "Template", ws.Spec.Template.Name)
	ctx = clog.IntoContext(ctx, log)

	current := ws.DeepCopy()

	log.DebugAll().DumpObject(r.Scheme, &ws, "before workspace")

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
		log.Info("status phase updated", "before", current.Status.Phase, "now", ws.Status.Phase)
	}

	// Requeue: true makes exponential backoff by default
	// https://github.com/kubernetes-sigs/controller-runtime/issues/808
	log.Debug().Info("finish reconcile", "requeue", requeue)
	return ctrl.Result{Requeue: requeue}, nil
}

func (r *WorkspaceStatusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&cosmov1alpha1.Workspace{}).
		Owns(&cosmov1alpha1.Instance{}).
		Build(r)
	if err != nil {
		return err
	}

	err = c.Watch(source.Kind(mgr.GetCache(), &corev1.Pod{}, handler.TypedEnqueueRequestsFromMapFunc[*corev1.Pod](r.findWorkspaceByPod)))
	if err != nil {
		return err
	}
	err = c.Watch(source.Kind(mgr.GetCache(), &corev1.Service{}, handler.TypedEnqueueRequestsFromMapFunc[*corev1.Service](r.findWorkspaceByService)))
	if err != nil {
		return err
	}
	return nil
}

func (r *WorkspaceStatusReconciler) findWorkspaceByPod(ctx context.Context, obj *corev1.Pod) []reconcile.Request {
	return findWorkspace(ctx, obj, r.Client)
}

func (r *WorkspaceStatusReconciler) findWorkspaceByService(ctx context.Context, obj *corev1.Service) []reconcile.Request {
	return findWorkspace(ctx, obj, r.Client)
}

func findWorkspace[T client.Object](ctx context.Context, obj T, c client.Client) []reconcile.Request {
	var ws cosmov1alpha1.Workspace
	if err := c.Get(ctx, types.NamespacedName{Name: obj.GetLabels()[cosmov1alpha1.LabelKeyInstanceName]}, &ws); err == nil {
		// request is Pod with "cosmo-workspace.github.io/instance" label
		return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: ws.Name, Namespace: ws.Namespace}}}
	}
	return nil
}

func listWorkspacePods(ctx context.Context, c client.Client, ws cosmov1alpha1.Workspace) ([]corev1.Pod, error) {
	var podList corev1.PodList

	ls := labels.NewSelector()
	req, _ := labels.NewRequirement(cosmov1alpha1.LabelKeyInstanceName, selection.Equals, []string{ws.GetName()})
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
