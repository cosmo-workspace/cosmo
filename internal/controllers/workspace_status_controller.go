package controllers

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
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

	key := r.getWorkspaceNamespacedName(ctx, req.NamespacedName)

	var ws cosmov1alpha1.Workspace
	if err := r.Get(ctx, key, &ws); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log = log.WithValues("UID", ws.UID, "Template", ws.Spec.Template.Name)
	ctx = clog.IntoContext(ctx, log)

	current := ws.DeepCopy()

	log.DumpObject(r.Scheme, &ws, "before workspace")

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
		log.Info("status updated")
	}

	log.Debug().Info("finish reconcile", "requeue", requeue)
	if requeue {
		return ctrl.Result{RequeueAfter: time.Second * 5}, nil
	}
	return ctrl.Result{}, nil
}

func (r *WorkspaceStatusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&cosmov1alpha1.Workspace{}).
		Owns(&cosmov1alpha1.Instance{}).
		Build(r)
	if err != nil {
		return err
	}

	// watch pods which has "cosmo-workspace.github.io/instance" label
	predi, _ := predicate.LabelSelectorPredicate(metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      cosmov1alpha1.LabelKeyInstanceName,
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
		// request is Pod with "cosmo-workspace.github.io/instance" label
		return types.NamespacedName{Name: pod.Labels[cosmov1alpha1.LabelKeyInstanceName], Namespace: pod.GetNamespace()}
	}

	var svc corev1.Service
	if err := r.Get(ctx, req, &svc); err == nil {
		// request is Service with "cosmo-workspace.github.io/instance" label
		return types.NamespacedName{Name: svc.Labels[cosmov1alpha1.LabelKeyInstanceName], Namespace: svc.GetNamespace()}
	}

	// request is Workspace
	return req
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
