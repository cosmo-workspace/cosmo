package controllers

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/auth/password"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	"github.com/cosmo-workspace/cosmo/pkg/useraddon"
)

// UserReconciler reconciles a Template object
type UserReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("UserReconciler").WithValues("req", req)

	log.Debug().Info("start reconcile")

	var user cosmov1alpha1.User
	if err := r.Get(ctx, req.NamespacedName, &user); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log = log.WithValues("UID", user.UID)
	ctx = clog.IntoContext(ctx, log)
	currentUser := user.DeepCopy()

	// reconcile namespace
	ns := corev1.Namespace{}
	ns.SetName(cosmov1alpha1.UserNamespace(user.Name))

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, &ns, func() error {
		return r.patchNamespaceToUserDesired(&ns, user)
	})
	if err != nil {
		r.Recorder.Eventf(&user, corev1.EventTypeWarning, "SyncFailed", "Failed to sync namespace %s: %v", ns.Name, err)
		return ctrl.Result{}, fmt.Errorf("failed to sync namespace: %w", err)
	}
	if op != controllerutil.OperationResultNone {
		log.Info("namespace synced", "namespace", ns.Name)
		r.Recorder.Eventf(&user, corev1.EventTypeNormal, "Synced", "Successfully reconciled. Namespace %s is %s", ns.Name, op)
	}

	user.Status.Phase = ns.Status.Phase

	gvk, err := apiutil.GVKForObject(&ns, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	user.Status.Namespace = cosmov1alpha1.ObjectRef{
		ObjectReference: corev1.ObjectReference{
			APIVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
			Name:       ns.GetName(),
			UID:        ns.GetUID(),
		},
		CreationTimestamp: &ns.CreationTimestamp,
	}

	if user.Spec.AuthType == cosmov1alpha1.UserAuthTypePasswordSecert {
		// generate default password if password secret is not found
		if _, err := password.GetDefaultPassword(ctx, r.Client, user.Name); err != nil && apierrs.IsNotFound(err) {
			if err := password.ResetPassword(ctx, r.Client, user.Name); err != nil {
				r.Recorder.Eventf(&user, corev1.EventTypeWarning, "PasswordInitFailed", "Failed to reset password: %v", err)
				log.Error(err, "failed to reset password")
				return ctrl.Result{}, err
			}
			log.Info("password secret initialized")
			r.Recorder.Eventf(&user, corev1.EventTypeNormal, "PasswordInitialized", "Successfully reset password secret")
		}
	}

	// reconcile user addon
	addonErrs := make([]error, 0)

	lastAddons := make([]cosmov1alpha1.ObjectRef, len(user.Status.Addons))
	copy(lastAddons, user.Status.Addons)

	currAddonsMap := make(map[types.UID]cosmov1alpha1.ObjectRef)
	for _, addon := range user.Spec.Addons {
		log.Debug().Info("syncing user addon", "addon", addon)

		inst := useraddon.EmptyInstanceObject(addon, user.GetName())
		if inst == nil {
			log.Error(errors.New("instance is nil"), "addon has no Template or ClusterTemplate", "addon", addon)
			continue
		}
		tmpl := useraddon.EmptyTemplateObject(addon)
		if err := r.Get(ctx, types.NamespacedName{Name: tmpl.GetName()}, tmpl); err != nil {
			addonErrs = append(addonErrs, fmt.Errorf("failed to create or update addon %s: failed to fetch template: %w", tmpl.GetName(), err))
			continue
		}

		op, err := controllerutil.CreateOrUpdate(ctx, r.Client, inst, func() error {
			if err := useraddon.PatchUserAddonInstanceAsDesired(inst, addon, user, r.Scheme); err != nil {
				return err
			}
			instance.Mutate(inst, tmpl)
			return nil
		})
		if err != nil {
			addonErrs = append(addonErrs, fmt.Errorf("failed to create or update addon %s :%w", inst.GetSpec().Template.Name, err))
			continue
		}

		if op != controllerutil.OperationResultNone {
			log.Info("addon synced", "addon", addon)
			r.Recorder.Eventf(&user, corev1.EventTypeNormal, "AddonSynced", "Addon %s is %s", addon.Template.Name, op)
		} else {
			log.Debug().Info("the result of update addon instance operation is None", "addon", addon)
		}

		ct := inst.GetCreationTimestamp()
		gvk, err := apiutil.GVKForObject(inst, r.Scheme)
		if err != nil {
			addonErrs = append(addonErrs, fmt.Errorf("failed to recognize addon instance GVK %s :%w", inst.GetSpec().Template.Name, err))
			continue
		}
		currAddonsMap[inst.GetUID()] = cosmov1alpha1.ObjectRef{
			ObjectReference: corev1.ObjectReference{
				APIVersion: gvk.GroupVersion().String(),
				Kind:       gvk.Kind,
				Name:       inst.GetName(),
				Namespace:  inst.GetNamespace(),
				UID:        inst.GetUID(),
			},
			CreationTimestamp: &ct,
		}
	}
	user.Status.Addons = objectRefMapToSlice(currAddonsMap)

	if len(addonErrs) > 0 {
		for _, e := range addonErrs {
			r.Recorder.Eventf(&user, corev1.EventTypeWarning, "AddonFailed", "Failed to sync addon: %v", e)
			log.Error(e, "failed to sync user addon")
		}
		user.Status.Phase = "AddonFailed"
		err = addonErrs[0]
	}

	// update user status
	if !equality.Semantic.DeepEqual(currentUser, &user) {
		log.Debug().PrintObjectDiff(currentUser, &user)
		if err := r.Status().Update(ctx, &user); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("status updated")
	}

	if user.Status.Phase != "AddonFailed" && !cosmov1alpha1.IsPruneDisabled(&user) {
		log.Debug().Info("checking garbage collection")
		shouldDeletes := objectRefNotExistsInMap(lastAddons, currAddonsMap)
		for _, d := range shouldDeletes {
			if skip, err := prune(ctx, r.Client, d); err != nil {
				log.Error(err, "failed to delete unused addon", "pruneAPIVersion", d.APIVersion, "pruneKind", d.Kind, "pruneName", d.Name, "pruneNamespace", d.Namespace)
				r.Recorder.Eventf(&user, corev1.EventTypeWarning, "GCFailed", "Failed to delete unused addon: kind=%s name=%s namespace=%s", d.Kind, d.Name, d.Namespace)
			} else if !skip {
				log.Info("deleted unmanaged addon", "apiVersion", d.APIVersion, "kind", d.Kind, "name", d.Name, "namespace", d.Namespace)
				r.Recorder.Eventf(&user, corev1.EventTypeNormal, "GC", "Deleted unmanaged addon: kind=%s name=%s namespace=%s", d.Kind, d.Name, d.Namespace)
			}
		}
	}

	log.Debug().Info("finish reconcile")
	return ctrl.Result{}, err
}

func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cosmov1alpha1.User{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}

func (r *UserReconciler) patchNamespaceToUserDesired(ns *corev1.Namespace, user cosmov1alpha1.User) error {
	label := ns.GetLabels()
	if label == nil {
		label = make(map[string]string)
	}
	label[cosmov1alpha1.NamespaceLabelKeyUserName] = user.Name
	ns.SetLabels(label)

	err := ctrl.SetControllerReference(&user, ns, r.Scheme)
	if err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	return nil
}
