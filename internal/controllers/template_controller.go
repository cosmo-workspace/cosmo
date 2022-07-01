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

// TemplateReconciler reconciles a Template object
type TemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *TemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("TemplateReconciler")
	ctx = clog.IntoContext(ctx, log)

	log.Debug().Info("start reconcile")

	if err := r.reconcile(ctx, req); err != nil {
		log.Error(err, "reconcile end with warn", "template", req.Name)
	}

	log.Debug().Info("finish reconcile")
	return ctrl.Result{}, nil
}

func (r *TemplateReconciler) reconcile(ctx context.Context, req ctrl.Request) error {
	log := clog.FromContext(ctx)

	var tmpl cosmov1alpha1.Template
	if err := r.Get(ctx, req.NamespacedName, &tmpl); err != nil {
		return client.IgnoreNotFound(err)
	}

	var insts cosmov1alpha1.InstanceList
	err := r.List(ctx, &insts)
	if err != nil {
		return fmt.Errorf("failed to list instances for template %s: %w", tmpl.Name, err)
	}

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

func (r *TemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cosmov1alpha1.Template{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(ce event.CreateEvent) bool { return false },
		}).
		Complete(r)
}

// update instance annotations to notify template updates
func notifyUpdateToInstance(ctx context.Context, c client.Client, updateTime time.Time, inst cosmov1alpha1.InstanceObject) error {
	ann := inst.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string)
	}
	ann[cosmov1alpha1.InstanceAnnKeyTemplateUpdated] = updateTime.String()
	inst.SetAnnotations(ann)

	return c.Update(ctx, inst)
}
