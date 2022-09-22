package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

// ClusterTemplateReconciler reconciles a ClusterTemplate object
type ClusterTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *ClusterTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("ClusterTemplateReconciler").WithValues("req", req)
	ctx = clog.IntoContext(ctx, log)

	log.Debug().Info("start reconcile")

	if err := r.reconcile(ctx, req); err != nil {
		log.Error(err, "reconcile end with warn", "clustertemplate", req.Name)
	}

	log.Debug().Info("finish reconcile")
	return ctrl.Result{}, nil
}

func (r *ClusterTemplateReconciler) reconcile(ctx context.Context, req ctrl.Request) error {
	log := clog.FromContext(ctx)

	var tmpl cosmov1alpha1.ClusterTemplate
	if err := r.Get(ctx, req.NamespacedName, &tmpl); err != nil {
		return client.IgnoreNotFound(err)
	}

	var insts cosmov1alpha1.ClusterInstanceList
	err := r.List(ctx, &insts)
	if err != nil {
		return fmt.Errorf("failed to list clusterinstances for clustertemplate %s: %w", tmpl.Name, err)
	}

	// update instance annotations to notify template updates
	now := time.Now()
	for _, inst := range insts.Items {
		if tmpl.Name != inst.GetSpec().Template.Name {
			continue
		}
		if err := notifyUpdateToInstance(ctx, r.Client, now, &inst); err != nil {
			log.Error(err, "failed to notify template updates", "tmplName", tmpl.Name, "instName", inst.GetName())
		}
	}
	return nil
}

func (r *ClusterTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cosmov1alpha1.ClusterTemplate{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(ce event.CreateEvent) bool { return false },
		}).
		Complete(r)
}
