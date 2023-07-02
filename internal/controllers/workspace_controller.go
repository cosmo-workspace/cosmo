package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	traefikv1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
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

	// sync instance
	inst := &cosmov1alpha1.Instance{}
	inst.SetName(ws.Name)
	inst.SetNamespace(ws.Namespace)
	op, err := kubeutil.CreateOrUpdate(ctx, r.Client, inst, func() error {
		return workspace.PatchWorkspaceInstanceAsDesired(inst, ws, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		r.Recorder.Eventf(&ws, corev1.EventTypeNormal, string(op), "successfully reconciled. instance synced")
	}
	gvk, _ := apiutil.GVKForObject(inst, r.Scheme)
	ws.Status.Instance = cosmov1alpha1.ObjectRef{
		ObjectReference: corev1.ObjectReference{
			APIVersion:      gvk.GroupVersion().String(),
			Kind:            gvk.Kind,
			Name:            inst.GetName(),
			Namespace:       inst.GetNamespace(),
			ResourceVersion: inst.GetResourceVersion(),
			UID:             inst.GetUID(),
		},
		CreationTimestamp: &inst.CreationTimestamp,
	}

	// sync ingress route
	ir := traefikv1.IngressRoute{}
	ir.SetName(ws.Name)
	ir.SetNamespace(ws.Namespace)
	op, err = kubeutil.CreateOrUpdate(ctx, r.Client, &ir, func() error {
		return r.TraefikIngressRouteCfg.PatchTraefikIngressRouteAsDesired(&ir, ws, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	if op != controllerutil.OperationResultNone {
		r.Recorder.Eventf(&ws, corev1.EventTypeNormal, string(op), "successfully reconciled. traefik ingress route synced")
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
