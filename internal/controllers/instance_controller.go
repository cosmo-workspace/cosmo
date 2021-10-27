package controllers

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	"github.com/cosmo-workspace/cosmo/pkg/transformer"
)

const (
	InstControllerFieldManager string = "cosmo-instance-controller"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	kosmo.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=cosmo.cosmo-workspace.github.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cosmo.cosmo-workspace.github.io,resources=instances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=*,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cosmo.cosmo-workspace.github.io,resources=instances/finalizers,verbs=update

func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("InstanceReconciler")
	ctx = clog.IntoContext(ctx, log)

	log.Debug().Info("start reconcile")

	var inst cosmov1alpha1.Instance
	if err := r.Get(ctx, req.NamespacedName, &inst); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, ignoreNotFound(err)
	}

	log.DebugAll().DumpObject(r.Scheme, &inst, "request object")

	before := inst.DeepCopy()

	// Fetch template
	tmpl, err := r.GetTemplate(ctx, inst.Spec.Template.Name)
	if err != nil {
		log.Error(err, "failed to get template", "tmplName", inst.Spec.Template.Name)
		return ctrl.Result{}, err
	}

	// Build child resource config by template
	builts, err := r.buildTemplateObjects(ctx, tmpl, &inst)
	if err != nil {
		r.Recorder.Event(&inst, corev1.EventTypeWarning, "Build failed", err.Error())
		return ctrl.Result{}, err
	}

	// Reconcile child resources
	if errs := r.applyChildObjects(ctx, &inst, builts); len(errs) != 0 {
		for _, err := range errs {
			r.Recorder.Event(&inst, corev1.EventTypeWarning, "Sync failed", err.Error())
		}
		return ctrl.Result{}, errors.New("apply child objects failed")
	}

	if !equality.Semantic.DeepEqual(*before, inst) {
		log.DebugAll().PrintObjectDiff(*before, inst)
		// Update status
		if err := r.Status().Update(ctx, &inst); err != nil {
			log.Error(err, "failed to update InstanceStatus")
			return ctrl.Result{}, err
		}
	}

	log.Info("finish reconcile")
	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) buildTemplateObjects(ctx context.Context, tmpl *cosmov1alpha1.Template, inst *cosmov1alpha1.Instance) ([]unstructured.Unstructured, error) {
	log := clog.FromContext(ctx).WithCaller()

	inst.Status.TemplateName = tmpl.GetName()

	builts, err := template.NewTemplateBuilder(tmpl.Spec.RawYaml, inst).
		ReplaceDefaultVars().
		ReplaceCustomVars().
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build template: %w", err)
	}

	transformers := []transformer.Transformer{
		// MetadataTransformer perform update each object's metadata
		transformer.NewMetadataTransformer(inst, tmpl, r.Scheme),
		// NetworkTransformer perform update ingresses and services by netSpec
		transformer.NewNetworkTransformer(inst.Spec.Override.Network, inst.Name),
		// JSONPatchTransformer perform JSONPatch
		transformer.NewJSONPatchTransformer(inst.Spec.Override.PatchesJson6902, inst.Name),
		// ScalingTransformer perform override replicas
		transformer.NewScalingTransformer(inst.Spec.Override.Scale, inst.Name),
	}

	for i := 0; i < len(builts); i++ {
		// Perform each transformers
		for _, trans := range transformers {
			transName := transformer.Name(trans)
			log.DebugAll().Info(fmt.Sprintf("transforming %s", transName), "transformer", transName, "kind", builts[i].GetKind(), "name", builts[i].GetName())
			before := builts[i].DeepCopy()

			transformed, err := trans.Transform(&builts[i])
			if err != nil {
				return nil, fmt.Errorf("failed to transform: %w", err)
			}
			builts[i] = *transformed

			if !equality.Semantic.DeepEqual(before, transformed) {
				log.DebugAll().PrintObjectDiff(before, transformed)
				log.DebugAll().Info("transformed", "transformer", transName, "kind", builts[i].GetKind(), "name", builts[i].GetName())
			} else {
				log.DebugAll().Info("not transformed", "transformer", transName, "kind", builts[i].GetKind(), "name", builts[i].GetName())
			}
		}
	}
	return builts, nil
}

func (r *InstanceReconciler) applyChildObjects(ctx context.Context, inst *cosmov1alpha1.Instance, builts []unstructured.Unstructured) []error {
	log := clog.FromContext(ctx).WithCaller()
	errs := make([]error, 0)

	refs := inst.Status.LastApplied

	if len(refs) == 0 {
		// First reconcile
		for _, built := range builts {
			if _, err := r.dryrunApply(ctx, &built); err != nil {
				// Ignore NotFound in case the template contains a dependency resource that was not found.
				if !apierrs.IsNotFound(err) {
					errs = append(errs, fmt.Errorf("dryrun failed: kind=%s name=%s: %w", built.GetKind(), built.GetName(), err))
				}
			}
		}
		if len(errs) != 0 {
			return errs
		}

		refs = make([]cosmov1alpha1.ObjectRef, 0, len(builts))
	}

	for _, built := range builts {
		mapping, err := r.RESTMapper().RESTMapping(built.GroupVersionKind().GroupKind(), built.GroupVersionKind().Version)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get rest mapping: kind=%s name=%s: %w", built.GetKind(), built.GetName(), err))
			continue
		}
		if mapping.Scope != meta.RESTScopeNamespace {
			errs = append(errs, fmt.Errorf("kind %s is not namespaced scope: scope=%s name=%s", built.GetKind(), mapping.Scope.Name(), built.GetName()))
			continue
		}

		current, err := r.GetUnstructured(ctx, built.GroupVersionKind(), built.GetName(), built.GetNamespace())
		if err != nil {
			// If not found, create resource
			if apierrs.IsNotFound(err) {
				log.Info("creating new built resource", "kind", built.GetKind(), "name", built.GetName())
				log.DebugAll().DumpObject(r.Scheme, &built, "built object")

				created, err := r.apply(ctx, &built)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to create resource: kind = %s name = %s: %w", built.GetKind(), built.GetName(), err))
					continue
				}

				r.Recorder.Eventf(inst, corev1.EventTypeNormal, "Synced", "%s %s created", built.GetKind(), built.GetName())

				refs = upsertLastAppliedResourceRef(refs, cosmov1alpha1.UnstructuredToResourceRef(*created, metav1.Now()), inst.Name)
			} else {
				errs = append(errs, fmt.Errorf("failed to get resource: kind = %s name = %s: %w", built.GetKind(), built.GetName(), err))
				continue
			}

		} else {
			// get desired state
			desired, err := r.dryrunApply(ctx, &built)
			if err != nil {
				errs = append(errs, fmt.Errorf("dryrun failed: kind=%s name=%s: %w", built.GetKind(), built.GetName(), err))
				continue
			}

			// compare current with the desired state
			if !kosmo.LooseDeepEqual(current.DeepCopy(), desired.DeepCopy()) {
				log.Info("current is not desired state, synced", "kind", desired.GetKind(), "name", desired.GetName())
				log.PrintObjectDiff(current, desired)

				// apply
				log.DebugAll().DumpObject(r.Scheme, &built, "applying object")
				if _, err := r.apply(ctx, &built); err != nil {
					errs = append(errs, fmt.Errorf("failed to apply resource %s %s: %w", built.GetKind(), built.GetName(), err))
					continue
				}

				r.Recorder.Eventf(inst, corev1.EventTypeNormal, "Synced", "%s %s is not desired state, synced", built.GetKind(), built.GetName())

				refs = upsertLastAppliedResourceRef(refs, cosmov1alpha1.UnstructuredToResourceRef(*desired, metav1.Now()), inst.Name)
			}
		}
	}
	inst.Status.LastApplied = refs
	return errs

}

func (r *InstanceReconciler) dryrunApply(ctx context.Context, obj *unstructured.Unstructured) (patched *unstructured.Unstructured, err error) {
	return r.Client.Apply(ctx, obj, InstControllerFieldManager, true, true)
}

func (r *InstanceReconciler) apply(ctx context.Context, obj *unstructured.Unstructured) (patched *unstructured.Unstructured, err error) {
	return r.Client.Apply(ctx, obj, InstControllerFieldManager, false, true)
}

func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cosmov1alpha1.Instance{}).
		Complete(r)
}

func upsertLastAppliedResourceRef(refs []cosmov1alpha1.ObjectRef, update cosmov1alpha1.ObjectRef, instName string) []cosmov1alpha1.ObjectRef {
	for i, v := range refs {
		if v.IsTarget(instName, update) {
			refs[i] = update
			return refs
		}
	}

	// not found
	refs = append(refs, update)
	return refs
}
