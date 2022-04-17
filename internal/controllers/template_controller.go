package controllers

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

const (
	TmplControllerFieldManager string = "cosmo-template-controller"
)

// TemplateReconciler reconciles a Template object
type TemplateReconciler struct {
	kosmo.Client
	Scheme *runtime.Scheme
}

func (r *TemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("TemplateReconciler")
	ctx = clog.IntoContext(ctx, log)

	log.Debug().Info("start reconcile")

	var tmpl cosmov1alpha1.Template
	if err := r.Get(ctx, req.NamespacedName, &tmpl); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var insts cosmov1alpha1.InstanceList
	err := r.List(ctx, &insts)
	if err != nil {
		log.Error(err, "failed to list template", "tmplName", tmpl.Name)
		return ctrl.Result{}, err
	}

	// update instance annotations to notify template updates
	now := time.Now()
	for _, inst := range insts.Items {
		if tmpl.Name != inst.Spec.Template.Name {
			continue
		}
		ann := inst.GetAnnotations()
		if ann == nil {
			ann = make(map[string]string)
		}
		ann[cosmov1alpha1.InstanceAnnKeyTemplateUpdated] = now.String()
		inst.SetAnnotations(ann)

		if err := r.Update(ctx, &inst); err != nil {
			log.Error(err, "failed to notify template updates", "tmplName", tmpl.Name, "instName", inst.Name)
		}
	}

	log.Debug().Info("finish reconcile")
	return ctrl.Result{}, nil
}

func (r *TemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cosmov1alpha1.Template{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(ce event.CreateEvent) bool { return false },
		}).
		Complete(r)
}
