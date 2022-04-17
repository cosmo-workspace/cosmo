package controllers

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

const (
	UserControllerFieldManager string = "cosmo-user-controller"
)

// UserReconciler reconciles a Template object
type UserReconciler struct {
	kosmo.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

func (r *UserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("UserReconciler")
	ctx = clog.IntoContext(ctx, log)

	log.Debug().Info("start reconcile")

	var user wsv1alpha1.User
	if err := r.Get(ctx, req.NamespacedName, &user); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	currentUser := user.DeepCopy()

	ns := corev1.Namespace{}
	ns.SetName(wsv1alpha1.UserNamespace(user.Name))

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, &ns, func() error {
		return r.patchNamespaceToUserDesired(&ns, user)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	now := metav1.Now()
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
		UpdateTimestamp:   &now,
	}

	switch op {
	case controllerutil.OperationResultCreated:
		r.Recorder.Eventf(&user, corev1.EventTypeNormal, "Created", "successfully namespace created")
		user.Status.Namespace.CreationTimestamp = &now

		log.Info("initializing password secret")
		if err := r.ResetPassword(ctx, user.Name); err != nil {
			r.Recorder.Eventf(&user, corev1.EventTypeWarning, "InitFailed", "failed to reset password: %v", err)
			log.Error(err, "failed to reset password")
			return ctrl.Result{}, err
		}

		if addons := r.userAddonInstances(ctx, user); len(addons) > 0 {
			errs := make([]error, 0)

			for _, addon := range addons {
				log.Info("creating user addon", "addon", addon.Spec.Template.Name)

				if err := r.Create(ctx, &addon); err != nil {
					errs = append(errs, fmt.Errorf("failed to create addon %s :%w", addon.Spec.Template.Name, err))
				}
			}

			if len(errs) > 0 {
				for _, e := range errs {
					r.Recorder.Eventf(&user, corev1.EventTypeWarning, "AddonFailed", "failed to create user addon: %v", e)
					log.Error(e, "failed to create user addon")
				}
				return ctrl.Result{}, errs[0]
			}
		}

	case controllerutil.OperationResultUpdated:
		r.Recorder.Eventf(&user, corev1.EventTypeNormal, "Updated", "namespace is not desired state, updated")
	}

	user.Status.Phase = ns.Status.Phase

	// update workspace status
	if !equality.Semantic.DeepEqual(currentUser, user) {
		if err := r.Status().Update(ctx, &user); err != nil {
			return ctrl.Result{}, err
		}
	}

	log.Debug().Info("finish reconcile")
	return ctrl.Result{}, nil
}

func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&wsv1alpha1.User{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}

func (r *UserReconciler) patchNamespaceToUserDesired(ns *corev1.Namespace, user wsv1alpha1.User) error {
	if ns == nil {
		return errors.New("namespace is nil")
	}

	label := ns.GetLabels()
	if label == nil {
		label = make(map[string]string)
	}
	label[wsv1alpha1.NamespaceLabelKeyUserID] = user.Name
	ns.SetLabels(label)

	err := ctrl.SetControllerReference(&user, ns, r.Scheme)
	if err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	return nil
}

func (r *UserReconciler) userAddonInstances(ctx context.Context, u wsv1alpha1.User) []cosmov1alpha1.Instance {
	if len(u.Spec.Addons) == 0 {
		return nil
	}
	log := clog.FromContext(ctx)

	tmpls, err := r.ListTemplatesByType(ctx, []string{wsv1alpha1.TemplateTypeUserAddon})
	if err != nil {
		log.Error(err, "failed to list templates")
		return nil
	}

	tmplNamesInSysNs := make(map[string]string)
	for _, v := range tmpls {
		if ann := v.GetAnnotations(); ann != nil {
			if sysNs, ok := ann[wsv1alpha1.TemplateAnnKeySysNsUserAddon]; ok {
				tmplNamesInSysNs[v.Name] = sysNs
			}
		}
	}

	insts := make([]cosmov1alpha1.Instance, len(u.Spec.Addons))
	for i, addon := range u.Spec.Addons {
		inst := cosmov1alpha1.Instance{}

		if addon.Vars == nil {
			addon.Vars = make(map[string]string)
		}
		addon.Vars[wsv1alpha1.TemplateVarUserNamespace] = u.Status.Namespace.Name

		inst.Spec = cosmov1alpha1.InstanceSpec{
			Template: addon.Template,
			Vars:     addon.Vars,
		}

		if sysNs, ok := tmplNamesInSysNs[addon.Template.Name]; ok {
			// system namespace
			inst.Name = fmt.Sprintf("useraddon-%s-%s", addon.Template.Name, u.GetName())
			inst.SetNamespace(sysNs)

		} else {
			// user namespace
			inst.Name = fmt.Sprintf("useraddon-%s", addon.Template.Name)
			inst.SetNamespace(u.Status.Namespace.Name)
		}

		err := ctrl.SetControllerReference(&u, &inst, r.Scheme)
		if err != nil {
			log.Error(err, "failed to set controller reference")
		}

		insts[i] = inst
	}

	return insts
}
