package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

// TemplateReconciler reconciles a Template object
type TemplateReconciler struct {
	client.Client
	Scheme       *runtime.Scheme
	FieldManager string
}

func (r *TemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("TemplateReconciler").WithValues("req", req)

	log.Debug().Info("start reconcile")

	var tmpl cosmov1alpha1.Template
	if err := r.Get(ctx, req.NamespacedName, &tmpl); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log = log.WithValues("UID", tmpl.UID)
	ctx = clog.IntoContext(ctx, log)

	if err := r.reconcile(ctx, &tmpl); err != nil {
		return ctrl.Result{}, err
	}

	log.Debug().Info("finish reconcile")
	return ctrl.Result{}, nil
}

func (r *TemplateReconciler) reconcile(ctx context.Context, tmpl *cosmov1alpha1.Template) error {
	log := clog.FromContext(ctx)

	var insts cosmov1alpha1.InstanceList
	err := r.List(ctx, &insts)
	if err != nil {
		return fmt.Errorf("failed to list instances for template %s: %w", tmpl.Name, err)
	}

	if errs := notifyUpdateToInstances(ctx, r.Client, tmpl, insts.InstanceObjects()); len(errs) > 0 {
		for _, e := range errs {
			log.Error(e, "failed to notify the update of template")
		}
		return fmt.Errorf("failed to notify the update of template %s", tmpl.Name)
	}

	return nil
}

func (r *TemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cosmov1alpha1.Template{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(ce event.CreateEvent) bool { return false },
		}).
		Complete(r)
}

func notifyUpdateToInstances(ctx context.Context, c client.Client, tmpl cosmov1alpha1.TemplateObject, insts []cosmov1alpha1.InstanceObject) []error {
	log := clog.FromContext(ctx)
	errs := make([]error, 0)
	for _, inst := range insts {
		if tmpl.GetName() != inst.GetSpec().Template.Name {
			continue
		}

		before := inst.DeepCopyObject()
		inst.GetStatus().TemplateResourceVersion = tmpl.GetResourceVersion()
		if equality.Semantic.DeepEqual(before, inst) {
			// log
			continue
		}

		log.Info("notify template update to reconcile instance again", "template", tmpl.GetName(), "templateResourceVersion", tmpl.GetResourceVersion(), "instance", inst.GetName())

		if err := c.Status().Update(ctx, inst); err != nil {
			errs = append(errs, fmt.Errorf("failed to update instance status: %s: %w", inst.GetName(), err))
		}
	}
	return errs
}
