package controllers

import (
	"context"
	"fmt"
	"slices"

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

	traefikv1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/workspace"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme

	TraefikIngressRouteCfg *workspace.TraefikIngressRouteConfig
	URLBaseProtocol        string
}

// +kubebuilder:rbac:groups=cosmo-workspace.github.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cosmo-workspace.github.io,resources=workspaces/status,verbs=get;update;patch
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("WorkspaceReconciler").WithValues("req", req)

	log.Debug().Info("start reconcile")

	var ws cosmov1alpha1.Workspace
	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log = log.WithValues("UID", ws.UID, "Template", ws.Spec.Template.Name)
	ctx = clog.IntoContext(ctx, log)

	currentWs := ws.DeepCopy()

	log.DumpObject(r.Scheme, currentWs, "request object")

	// sync workspace config with template
	cfg, err := getWorkspaceConfig(ctx, r.Client, ws.Spec.Template.Name)
	if err != nil {
		return ctrl.Result{}, err
	}
	ws.Status.Config = cfg

	inst := &cosmov1alpha1.Instance{}
	inst.SetName(ws.Name)
	inst.SetNamespace(ws.Namespace)

	tmpl := &cosmov1alpha1.Template{}
	if err := r.Get(ctx, types.NamespacedName{Name: ws.Spec.Template.Name}, tmpl); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to fetch template %s: %w", ws.Spec.Template.Name, err)
	}

	// sync
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, inst, func() error {
		if err := workspace.PatchWorkspaceInstanceAsDesired(inst, &ws, r.Scheme); err != nil {
			return err
		}
		instance.Mutate(inst, tmpl)
		return nil
	})
	if err != nil {
		if apierrs.IsConflict(err) {
			// if conflict, retry
			return ctrl.Result{Requeue: true}, nil
		} else {
			kosmo.WorkspaceEventf(r.Recorder, &ws, corev1.EventTypeWarning, "SyncFailed", "Failed to sync instance %s: %v", inst.Name, err)
			return ctrl.Result{}, fmt.Errorf("failed to sync instance: %w", err)
		}
	}
	if op != controllerutil.OperationResultNone {
		log.Info("instance synced", "instance", inst.Name)
		kosmo.WorkspaceEventf(r.Recorder, &ws, corev1.EventTypeNormal, "Synced", "Successfully reconciled. Instance %s is %s", inst.Name, op)
	} else {
		log.Debug().Info("the result of update workspace instance operation is None", "instance", inst.Name)
	}

	gvk, _ := apiutil.GVKForObject(inst, r.Scheme)
	ws.Status.Instance = cosmov1alpha1.ObjectRef{
		ObjectReference: corev1.ObjectReference{
			APIVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
			Name:       inst.GetName(),
			Namespace:  inst.GetNamespace(),
			UID:        inst.GetUID(),
		},
		CreationTimestamp: &inst.CreationTimestamp,
	}

	// sync ingress route
	ir := traefikv1.IngressRoute{}
	ir.SetName(ws.Name)
	ir.SetNamespace(ws.Namespace)
	op, err = controllerutil.CreateOrUpdate(ctx, r.Client, &ir, func() error {
		return r.TraefikIngressRouteCfg.PatchTraefikIngressRouteAsDesired(&ir, ws, r.Scheme)
	})
	if err != nil {
		if apierrs.IsConflict(err) {
			// if conflict, retry
			return ctrl.Result{Requeue: true}, nil
		} else {
			kosmo.WorkspaceEventf(r.Recorder, &ws, corev1.EventTypeWarning, "SyncFailed", "Failed to sync traefik ingress route %s: %v", ir.Name, err)
			return ctrl.Result{}, fmt.Errorf("failed to sync traefik ingress route: %w", err)
		}
	}
	if op != controllerutil.OperationResultNone {
		log.Info("traefik ingress route synced", "ingressroute", ir.Name)
		kosmo.WorkspaceEventf(r.Recorder, &ws, corev1.EventTypeNormal, "Synced", "Successfully reconciled. Traefik ingress route %s is %s", ir.Name, op)
	}

	// generate URL and set to status
	urlMap := r.GenWorkspaceURLMap(ctx, ws)
	log.DebugAll().Info(fmt.Sprintf("workspace urlmap: %s", urlMap))
	ws.Status.URLs = urlMap

	// update workspace status
	if !equality.Semantic.DeepEqual(currentWs, &ws) {
		log.Debug().PrintObjectDiff(currentWs, &ws)
		if err := r.Status().Update(ctx, &ws); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("status updated")
	}

	// update shared workspace user status
	for _, netRule := range ws.Spec.Network {
		for _, AllowedUsers := range netRule.AllowedUsers {
			var user cosmov1alpha1.User
			err := r.Get(ctx, client.ObjectKey{Name: AllowedUsers}, &user)
			if err != nil {
				kosmo.WorkspaceEventf(r.Recorder, &ws, corev1.EventTypeWarning, "UpdateAllowedUserStatusFailed", "Failed to update share user status: %v", err)
				continue
			}

			wsRef := cosmov1alpha1.ObjectRef{
				ObjectReference: corev1.ObjectReference{
					Name:      ws.Name,
					Namespace: ws.Namespace,
				},
			}
			if len(user.Status.SharedWorkspaces) > 0 {
				if !slices.ContainsFunc(user.Status.SharedWorkspaces, func(v cosmov1alpha1.ObjectRef) bool { return v == wsRef }) {
					user.Status.SharedWorkspaces = append(user.Status.SharedWorkspaces, wsRef)
				}
			} else {
				user.Status.SharedWorkspaces = []cosmov1alpha1.ObjectRef{wsRef}
			}

			if err := r.Status().Update(ctx, &user); err != nil {
				if !apierrs.IsConflict(err) {
					kosmo.UserEventf(r.Recorder, &user, corev1.EventTypeWarning, "UpdateUserStatusFailed", "Failed to update user status: %v", err)
				}
				return ctrl.Result{}, err
			}
			log.Info("user status updated", "user", user.Name, "add", wsRef, "result", user.Status.SharedWorkspaces)
		}
	}

	log.Debug().Info("finish reconcile")
	return ctrl.Result{}, nil
}

func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cosmov1alpha1.Workspace{}).
		Owns(&cosmov1alpha1.Instance{}).
		Complete(r)
}

func getWorkspaceConfig(ctx context.Context, c client.Client, tmplName string) (cfg cosmov1alpha1.Config, err error) {
	tmpl := &cosmov1alpha1.Template{}
	if err := c.Get(ctx, types.NamespacedName{Name: tmplName}, tmpl); err != nil {
		return cfg, err
	}
	return workspace.ConfigFromTemplateAnnotations(tmpl)
}

func (r *WorkspaceReconciler) GenWorkspaceURLMap(ctx context.Context, ws cosmov1alpha1.Workspace) map[string]string {
	urlMap := make(map[string]string)
	for _, netRule := range ws.Spec.Network {
		host := cosmov1alpha1.GenHost(r.TraefikIngressRouteCfg.HostBase, r.TraefikIngressRouteCfg.Domain, netRule.HostPrefix(), ws)
		url := cosmov1alpha1.GenURL(r.URLBaseProtocol, host, netRule.HTTPPath)
		urlMap[netRule.UniqueKey()] = url
	}
	return urlMap
}
