package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/workspace"
	"github.com/cosmo-workspace/cosmo/pkg/wscfg"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=workspace.cosmo.cosmo-workspace.github.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=workspace.cosmo.cosmo-workspace.github.io,resources=workspaces/status,verbs=get;update;patch
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("WorkspaceReconciler").WithValues("req", req)
	ctx = clog.IntoContext(ctx, log)

	log.Debug().Info("start reconcile")

	var ws wsv1alpha1.Workspace
	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	currentWs := ws.DeepCopy()

	log.DumpObject(r.Scheme, currentWs, "request object")

	// sync workspace config with template
	cfg, err := getWorkspaceConfig(ctx, r.Client, ws.Spec.Template.Name)
	if err != nil {
		return ctrl.Result{}, err
	}
	ws.Status.Config = cfg

	inst := &cosmov1alpha1.Instance{}
	inst.SetName(ws.Name)
	inst.SetNamespace(ws.Namespace)

	op, err := kubeutil.CreateOrUpdate(ctx, r.Client, inst, func() error {
		return workspace.PatchWorkspaceInstanceAsDesired(inst, ws, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	now := metav1.Now()
	gvk, _ := apiutil.GVKForObject(inst, r.Scheme)
	ws.Status.Instance = cosmov1alpha1.ObjectRef{
		ObjectReference: corev1.ObjectReference{
			APIVersion:      gvk.GroupVersion().String(),
			Kind:            gvk.Kind,
			Name:            inst.GetName(),
			Namespace:       inst.GetNamespace(),
			ResourceVersion: inst.GetResourceVersion(),
			UID:             inst.GetUID(),
		},
		CreationTimestamp: &inst.CreationTimestamp,
		UpdateTimestamp:   &now,
	}

	switch op {
	case controllerutil.OperationResultCreated:
		r.Recorder.Eventf(&ws, corev1.EventTypeNormal, "Created", "successfully instance created")
		ws.Status.Instance.CreationTimestamp = &now

	case controllerutil.OperationResultUpdated:
		r.Recorder.Eventf(&ws, corev1.EventTypeNormal, "Updated", "instance is not desired state, updated")
	}

	// update workspace status
	if !equality.Semantic.DeepEqual(currentWs, ws) {
		if err := r.Status().Update(ctx, &ws); err != nil {
			return ctrl.Result{}, err
		}
	}

	log.Info("finish reconcile")
	return ctrl.Result{}, nil
}

func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&wsv1alpha1.Workspace{}).
		Owns(&cosmov1alpha1.Instance{}).
		Complete(r)
}

func getWorkspaceConfig(ctx context.Context, c client.Client, tmplName string) (cfg wsv1alpha1.Config, err error) {
	tmpl := &cosmov1alpha1.Template{}
	if err := c.Get(ctx, types.NamespacedName{Name: tmplName}, tmpl); err != nil {
		return cfg, err
	}
	return wscfg.ConfigFromTemplateAnnotations(tmpl)
}
