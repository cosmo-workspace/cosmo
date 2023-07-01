package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
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
		r.Recorder.Eventf(&user, corev1.EventTypeWarning, "Sync Failed", "failed to sync namespace: %v", err)
		return ctrl.Result{}, fmt.Errorf("failed to sync namespace: %w", err)
	}
	if op != controllerutil.OperationResultNone {
		r.Recorder.Eventf(&user, corev1.EventTypeNormal, string(op), "successfully reconciled. namespace synced")
	}

	user.Status.Phase = ns.Status.Phase

	gvk, err := apiutil.GVKForObject(&ns, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	user.Status.Namespace = cosmov1alpha1.ObjectRef{
		ObjectReference: corev1.ObjectReference{
			APIVersion:      gvk.GroupVersion().String(),
			Kind:            gvk.Kind,
			Name:            ns.GetName(),
			UID:             ns.GetUID(),
			ResourceVersion: ns.GetResourceVersion(),
		},
		CreationTimestamp: &ns.CreationTimestamp,
	}

	if user.Spec.AuthType == cosmov1alpha1.UserAuthTypePasswordSecert {
		// generate default password if password secret is not found
		if _, err := password.GetDefaultPassword(ctx, r.Client, user.Name); err != nil && apierrs.IsNotFound(err) {
			if err := password.ResetPassword(ctx, r.Client, user.Name); err != nil {
				r.Recorder.Eventf(&user, corev1.EventTypeWarning, "InitFailed", "failed to reset password: %v", err)
				log.Error(err, "failed to reset password")
				return ctrl.Result{}, err
			}
			r.Recorder.Eventf(&user, corev1.EventTypeNormal, "PasswordSecret Initialized", "successfully reset password secret")
		}
	}

	// reconcile user addon
	addonErrs := make([]error, 0)

	lastAddons := make([]cosmov1alpha1.ObjectRef, len(user.Status.Addons))
	copy(lastAddons, user.Status.Addons)

	currAddonsMap := make(map[types.UID]cosmov1alpha1.ObjectRef)
	for _, addon := range user.Spec.Addons {
		log.Info("syncing user addon", "addon", addon)

		inst := useraddon.EmptyInstanceObject(addon, user.GetName())
		if inst == nil {
			log.Info("WARNING: addon has no Template or ClusterTemplate", "addon", addon)
			continue
		}

		op, err := kubeutil.CreateOrUpdate(ctx, r.Client, inst, func() error {
			return useraddon.PatchUserAddonInstanceAsDesired(inst, addon, user, r.Scheme)
		})
		if err != nil {
			addonErrs = append(addonErrs, fmt.Errorf("failed to create or update addon %s :%w", inst.GetSpec().Template.Name, err))
			continue
		}

		if op != controllerutil.OperationResultNone {
			r.Recorder.Eventf(&user, corev1.EventTypeNormal, "Addon Synced", fmt.Sprintf("addon %s is %s", addon.Template.Name, op))
		}

		ct := inst.GetCreationTimestamp()
		gvk, err := apiutil.GVKForObject(inst, r.Scheme)
		if err != nil {
			addonErrs = append(addonErrs, fmt.Errorf("failed to recognize addon instance GVK %s :%w", inst.GetSpec().Template.Name, err))
			continue
		}
		currAddonsMap[inst.GetUID()] = cosmov1alpha1.ObjectRef{
			ObjectReference: corev1.ObjectReference{
				APIVersion:      gvk.GroupVersion().String(),
				Kind:            gvk.Kind,
				Name:            inst.GetName(),
				Namespace:       inst.GetNamespace(),
				UID:             inst.GetUID(),
				ResourceVersion: inst.GetResourceVersion(),
			},
			CreationTimestamp: &ct,
		}
	}
	user.Status.Addons = objectRefMapToSlice(currAddonsMap)

	if len(addonErrs) > 0 {
		for _, e := range addonErrs {
			r.Recorder.Eventf(&user, corev1.EventTypeWarning, "AddonFailed", "failed to create or update user addon: %v", e)
			log.Error(e, "failed to create or update user addon")
		}
		user.Status.Phase = "AddonFailed"
	}

	// update user status
	if !equality.Semantic.DeepEqual(currentUser, user) {
		log.Debug().PrintObjectDiff(currentUser, &user)
		if err := r.Status().Update(ctx, &user); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("status updated")
	}

	// garbage collection
	shouldDeletes := objectRefNotExistsInMap(lastAddons, currAddonsMap)
	for _, d := range shouldDeletes {
		log.Info("start garbage collection", "apiVersion", d.APIVersion, "kind", d.Kind, "name", d.Name)

		var obj unstructured.Unstructured
		obj.SetAPIVersion(d.APIVersion)
		obj.SetKind(d.Kind)
		err := r.Get(ctx, types.NamespacedName{Name: d.GetName(), Namespace: d.Namespace}, &obj)
		if err != nil {
			if !apierrs.IsNotFound(err) {
				log.Error(err, "failed to get object to be deleted", "apiVersion", d.APIVersion, "kind", d.Kind, "name", d.Name)
			}
			continue
		}

		if err := r.Delete(ctx, &obj); err != nil {
			r.Recorder.Eventf(&user, corev1.EventTypeWarning, "GCFailed", "failed to delete unused addon: %s %s", obj.GetKind(), obj.GetName())
		}
		r.Recorder.Eventf(&user, corev1.EventTypeNormal, "GC", "deleted unmanaged addon: %s %s", obj.GetKind(), obj.GetName())
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
