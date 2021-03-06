package controllers

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	"github.com/cosmo-workspace/cosmo/pkg/transformer"
)

// ClusterInstanceReconciler reconciles a ClusterInstance object
type ClusterInstanceReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	impl instanceReconciler
}

func (r *ClusterInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("ClusterInstanceReconciler")
	ctx = clog.IntoContext(ctx, log)

	log.Debug().Info("start reconcile")

	var inst cosmov1alpha1.ClusterInstance
	if err := r.Get(ctx, req.NamespacedName, &inst); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	before := inst.DeepCopy()
	log.DebugAll().DumpObject(r.Scheme, before, "request object")

	tmpl := &cosmov1alpha1.ClusterTemplate{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: inst.GetSpec().Template.Name}, tmpl)
	if err != nil {
		log.Error(err, "failed to get cluster template", "tmplName", inst.Spec.Template.Name)
		return ctrl.Result{}, err
	}

	// 1. Build Unstructured objects
	objects, err := template.BuildObjects(tmpl.Spec, &inst)
	if err != nil {
		r.Recorder.Event(&inst, corev1.EventTypeWarning, "BuildFailed", err.Error())
		return ctrl.Result{}, err
	}

	// 2. Transform the objects
	objects, err = transformer.ApplyTransformers(ctx, transformer.AllTransformers(&inst, r.Scheme, tmpl), objects)
	if err != nil {
		r.Recorder.Event(&inst, corev1.EventTypeWarning, "BuildFailed", err.Error())
		return ctrl.Result{}, err
	}

	// 3. Reconcile objects
	if errs := r.impl.reconcileObjects(ctx, &inst, objects); len(errs) != 0 {
		for _, err := range errs {
			r.Recorder.Event(&inst, corev1.EventTypeWarning, "SyncFailed", err.Error())
		}
		// requeue
		return ctrl.Result{}, errors.New("apply child objects failed")
	}

	// 4. Update status
	if !equality.Semantic.DeepEqual(before, &inst) {
		log.DebugAll().PrintObjectDiff(before, &inst)
		// Update status
		if err := r.Status().Update(ctx, &inst); err != nil {
			log.Error(err, "failed to update InstanceStatus")
			return ctrl.Result{}, err
		}
	}

	log.Info("finish reconcile")
	return ctrl.Result{}, nil
}

func (r *ClusterInstanceReconciler) SetupWithManager(mgr ctrl.Manager, fieldManager string) error {
	r.impl = instanceReconciler{Client: r.Client, Recorder: r.Recorder, Scheme: r.Scheme, FieldManager: fieldManager}
	return ctrl.NewControllerManagedBy(mgr).
		For(&cosmov1alpha1.ClusterInstance{}).
		Complete(r)
}
