package controllers

import (
	"context"
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	"github.com/cosmo-workspace/cosmo/pkg/transformer"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
	Domain   string

	impl instanceReconciler
}

//+kubebuilder:rbac:groups=cosmo-workspace.github.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cosmo-workspace.github.io,resources=instances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=*,resources=*,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cosmo-workspace.github.io,resources=instances/finalizers,verbs=update

func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("InstanceReconciler").WithValues("req", req)

	log.Debug().Info("start reconcile")

	var inst cosmov1alpha1.Instance
	if err := r.Get(ctx, req.NamespacedName, &inst); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log = log.WithValues("UID", inst.UID, "Template", inst.Spec.Template.Name)
	ctx = clog.IntoContext(ctx, log)

	before := inst.DeepCopy()
	log.DebugAll().DumpObject(r.Scheme, before, "request object")

	tmpl := &cosmov1alpha1.Template{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: inst.GetSpec().Template.Name}, tmpl)
	if err != nil {
		log.Error(err, "failed to get template", "tmplName", inst.Spec.Template.Name)
		return ctrl.Result{}, err
	}
	inst.Status.TemplateName = tmpl.Name
	inst.Status.TemplateResourceVersion = tmpl.ResourceVersion

	// 1. Build Unstructured objects
	objects, err := template.BuildObjects(tmpl.Spec, &inst, r.Domain)
	if err != nil {
		kosmo.InstanceEventf(r.Recorder, &inst, corev1.EventTypeWarning, "BuildFailed", "Failed to build manifests from Template: %v", err)
		return ctrl.Result{}, err
	}

	// 2. Transform the objects
	objects, err = transformer.ApplyTransformers(ctx, transformer.AllTransformers(&inst, r.Scheme, tmpl), objects)
	if err != nil {
		kosmo.InstanceEventf(r.Recorder, &inst, corev1.EventTypeWarning, "BuildFailed", "Failed to build resources: %v", err)
		return ctrl.Result{}, err
	}

	// 3. Reconcile objects
	if errs := r.impl.reconcileObjects(ctx, &inst, objects); len(errs) != 0 {
		for _, err := range errs {
			kosmo.InstanceEventf(r.Recorder, &inst, corev1.EventTypeWarning, "SyncFailed", "Failed to sync objects: %v", err)
		}
		// requeue
		return ctrl.Result{}, fmt.Errorf("apply child objects failed: %w", errs[0])
	}

	// 4. Update status
	if !equality.Semantic.DeepEqual(before, &inst) {
		log.Debug().PrintObjectDiff(before, &inst)
		if err := r.Status().Update(ctx, &inst); err != nil {
			log.Error(err, "failed to update InstanceStatus")
			return ctrl.Result{}, err
		}
		log.Info("status updated")
	}

	log.Debug().Info("finish reconcile")
	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager, fieldManager string) error {
	r.impl = instanceReconciler{Client: r.Client, Recorder: r.Recorder, Scheme: r.Scheme, FieldManager: fieldManager}
	return ctrl.NewControllerManagedBy(mgr).
		For(&cosmov1alpha1.Instance{}).
		Complete(r)
}

type instanceReconciler struct {
	client.Client
	Recorder     record.EventRecorder
	Scheme       *runtime.Scheme
	FieldManager string
}

func (r *instanceReconciler) reconcileObjects(ctx context.Context, inst cosmov1alpha1.InstanceObject, objects []unstructured.Unstructured) []error {
	log := clog.FromContext(ctx).WithCaller()
	errs := make([]error, 0)

	lastApplied := make([]cosmov1alpha1.ObjectRef, len(inst.GetStatus().LastApplied))
	copy(lastApplied, inst.GetStatus().LastApplied)

	currAppliedMap := make(map[types.UID]cosmov1alpha1.ObjectRef)

	// check dry-run apply on first reconciliation
	if len(inst.GetStatus().LastApplied) == 0 {
		for _, built := range objects {
			if _, err := r.dryrunApply(ctx, &built, r.FieldManager); err != nil {
				// ignore NotFound in case the template contains a dependency resource that was not found.
				if !apierrs.IsNotFound(err) {
					errs = append(errs, fmt.Errorf("dryrun failed: kind=%s name=%s: %w", built.GetKind(), built.GetName(), err))
				}
			}
		}
		if len(errs) != 0 {
			return errs
		}
	}

	inst.GetStatus().TemplateObjectsCount = len(objects)

	for _, built := range objects {
		mapping, err := r.RESTMapper().RESTMapping(built.GroupVersionKind().GroupKind(), built.GroupVersionKind().Version)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get rest mapping: kind=%s name=%s: %w", built.GetKind(), built.GetName(), err))
			continue
		}

		// namespaced scope instance cannot create cluster scope resources
		if inst.GetScope() == meta.RESTScopeNamespace && mapping.Scope != inst.GetScope() {
			errs = append(errs, fmt.Errorf("kind %s is not scope %s: scope=%s name=%s", built.GetKind(), inst.GetScope(), mapping.Scope.Name(), built.GetName()))
			continue
		}

		current, err := kubeutil.GetUnstructured(ctx, r.Client, built.GroupVersionKind(), built.GetName(), built.GetNamespace())
		if err != nil {
			// if not found, create resource
			if apierrs.IsNotFound(err) {
				log.Info("creating new built resource", "kind", built.GetKind(), "name", built.GetName())
				log.Debug().DumpObject(r.Scheme, &built, "built object")

				created, err := r.apply(ctx, &built, r.FieldManager)
				if err != nil {
					errs = append(errs, fmt.Errorf("failed to create resource: kind = %s name = %s: %w", built.GetKind(), built.GetName(), err))
					continue
				}
				kosmo.InstanceEventf(r.Recorder, inst, corev1.EventTypeNormal, "Synced", "%s %s is created", built.GetKind(), built.GetName())
				currAppliedMap[created.GetUID()] = unstToObjectRef(created)

			} else {
				errs = append(errs, fmt.Errorf("failed to get resource: kind = %s name = %s: %w", built.GetKind(), built.GetName(), err))
				continue
			}

		} else {
			// get desired state
			desired, err := r.dryrunApply(ctx, &built, r.FieldManager)
			if err != nil {
				errs = append(errs, fmt.Errorf("dryrun failed: kind=%s name=%s: %w", built.GetKind(), built.GetName(), err))
				continue
			}
			currAppliedMap[desired.GetUID()] = unstToObjectRef(desired)

			// compare current with the desired state
			if !kubeutil.LooseDeepEqual(current, desired) {
				log.Info("current is not desired state, synced", "kind", desired.GetKind(), "name", desired.GetName())
				log.Debug().PrintObjectDiff(current, desired)

				// apply
				log.DumpObject(r.Scheme, &built, "applying object")
				if _, err := r.apply(ctx, &built, r.FieldManager); err != nil {
					errs = append(errs, fmt.Errorf("failed to apply resource %s %s: %w", built.GetKind(), built.GetName(), err))
				} else {
					kosmo.InstanceEventf(r.Recorder, inst, corev1.EventTypeNormal, "Synced", "%s %s is not desired state, synced", built.GetKind(), built.GetName())
				}
			}
		}
	}

	// garbage collection
	if len(errs) == 0 && !cosmov1alpha1.KeepResourceDeletePolicy(inst) {
		log.Debug().Info("checking garbage collection")
		shouldDeletes := objectRefNotExistsInMap(lastApplied, currAppliedMap)
		for _, d := range shouldDeletes {
			if skip, err := prune(ctx, r.Client, d); err != nil {
				log.Error(err, "failed to delete unused obj", "pruneAPIVersion", d.APIVersion, "pruneKind", d.Kind, "pruneName", d.Name, "pruneNamespace", d.Namespace)
				kosmo.InstanceEventf(r.Recorder, inst, corev1.EventTypeWarning, "GCFailed", "Failed to delete unused obj: kind=%s name=%s namespace=%s", d.Kind, d.Name, d.Namespace)
			} else if !skip {
				log.Info("deleted unmanaged object", "apiVersion", d.APIVersion, "kind", d.Kind, "name", d.Name, "namespace", d.Namespace)
				kosmo.InstanceEventf(r.Recorder, inst, corev1.EventTypeNormal, "GC", "Deleted unmanaged object: kind=%s name=%s namespace=%s", d.Kind, d.Name, d.Namespace)
			}
		}
	}

	inst.GetStatus().LastApplied = objectRefMapToSlice(currAppliedMap)
	inst.GetStatus().LastAppliedObjectsCount = len(inst.GetStatus().LastApplied)

	return errs
}

func (r *instanceReconciler) dryrunApply(ctx context.Context, obj *unstructured.Unstructured, fieldManager string) (patched *unstructured.Unstructured, err error) {
	return kubeutil.Apply(ctx, r.Client, obj, fieldManager, true, true)
}

func (r *instanceReconciler) apply(ctx context.Context, obj *unstructured.Unstructured, fieldManager string) (patched *unstructured.Unstructured, err error) {
	return kubeutil.Apply(ctx, r.Client, obj, fieldManager, false, true)
}

// unstToObjectRef generate ObjectRef by Unstructured object
func unstToObjectRef(obj *unstructured.Unstructured) cosmov1alpha1.ObjectRef {
	ref := cosmov1alpha1.ObjectRef{}
	ref.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
	ref.Name = obj.GetName()
	ref.Namespace = obj.GetNamespace()
	ref.UID = obj.GetUID()
	ref.ResourceVersion = obj.GetResourceVersion()

	create := obj.GetCreationTimestamp()
	ref.CreationTimestamp = &create
	return ref
}

func sliceToObjectMap(s []cosmov1alpha1.ObjectRef) map[types.UID]cosmov1alpha1.ObjectRef {
	m := make(map[types.UID]cosmov1alpha1.ObjectRef)
	for _, objRef := range s {
		m[objRef.UID] = objRef
	}
	return m
}

func objectRefMapToSlice(m map[types.UID]cosmov1alpha1.ObjectRef) []cosmov1alpha1.ObjectRef {
	if len(m) == 0 {
		return nil
	}
	s := make([]cosmov1alpha1.ObjectRef, len(m))
	var i int
	for _, v := range m {
		s[i] = v
		i++
	}
	sort.SliceStable(s, func(i, j int) bool {
		x, y := s[i], s[j]
		if x.APIVersion != y.APIVersion {
			return x.APIVersion < y.APIVersion
		} else if x.Kind != y.Kind {
			return x.Kind < y.Kind
		} else {
			return x.Name < y.Name
		}
	})
	return s
}

func objectRefNotExistsInMap(s []cosmov1alpha1.ObjectRef, m map[types.UID]cosmov1alpha1.ObjectRef) []cosmov1alpha1.ObjectRef {
	notExists := make([]cosmov1alpha1.ObjectRef, 0)
	for _, ss := range s {
		_, exist := m[ss.UID]
		if !exist {
			notExists = append(notExists, ss)
		}
	}
	return notExists
}

func prune(ctx context.Context, r client.Client, d cosmov1alpha1.ObjectRef) (skip bool, err error) {
	log := clog.FromContext(ctx).WithValues("pruneAPIVersion", d.APIVersion, "pruneKind", d.Kind, "pruneName", d.Name, "pruneNamespace", d.Namespace)

	var obj unstructured.Unstructured
	obj.SetAPIVersion(d.APIVersion)
	obj.SetKind(d.Kind)
	err = r.Get(ctx, types.NamespacedName{Name: d.Name, Namespace: d.Namespace}, &obj)
	if err != nil {
		log.Error(err, "failed to get object to be deleted")
		return true, nil
	}

	if obj.GetUID() != d.UID {
		log.Error(err, "target object UID is changed. skip pruning", "desiredUID", d.UID, "currentUID", obj.GetUID())
		return true, nil
	}
	if cosmov1alpha1.KeepResourceDeletePolicy(&obj) {
		log.Debug().Info("skip pruning by annotation", "apiVersion")
		return true, nil
	}

	return false, r.Delete(ctx, &obj)
}
