package controllers

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/wsnet"
)

const (
	WsControllerFieldManager string = "cosmo-workspace-controller"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	kosmo.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=workspace.cosmo.cosmo-workspace.github.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=workspace.cosmo.cosmo-workspace.github.io,resources=workspaces/status,verbs=get;update;patch
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("WorkspaceReconciler").WithValues("req", req)
	ctx = clog.IntoContext(ctx, log)

	log.Debug().Info("start reconcile")

	var ws wsv1alpha1.Workspace
	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, ignoreNotFound(err)
	}
	currentWs := ws.DeepCopy()

	log.DebugAll().DumpObject(r.Scheme, currentWs, "request object")

	// sync workspace config with template
	cfg, err := r.GetWorkspaceConfig(ctx, ws.Spec.Template.Name)
	if err != nil {
		return ctrl.Result{}, err
	}
	ws.Status.Config = cfg

	inst := &cosmov1alpha1.Instance{}
	inst.SetName(ws.Name)
	inst.SetNamespace(ws.Namespace)

	op, err := ctrl.CreateOrUpdate(ctx, r.Client, inst, func() error {
		return r.patchInstanceToWorkspaceDesired(inst, ws)
	})
	if err != nil {
		return ctrl.Result{}, err
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

	if op != controllerutil.OperationResultNone {
		r.Recorder.Eventf(&ws, corev1.EventTypeNormal, string(op), "successfully reconciled. instance synced")
	}

	// update workspace status
	if !equality.Semantic.DeepEqual(currentWs, ws) {
		if err := r.Status().Update(ctx, &ws); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("status updated")
	}

	log.Info("finish reconcile")
	return ctrl.Result{}, nil
}

func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&wsv1alpha1.Workspace{}).
		Owns(&cosmov1alpha1.Instance{}).
		Complete(r)
}

func (r *WorkspaceReconciler) patchInstanceToWorkspaceDesired(inst *cosmov1alpha1.Instance, ws wsv1alpha1.Workspace) error {
	if inst == nil {
		return errors.New("instance is nil")
	}

	svcPorts := make([]corev1.ServicePort, len(ws.Spec.Network))
	ingRules := make([]netv1.IngressRule, len(ws.Spec.Network))

	for i, netRule := range ws.Spec.Network {
		svcPorts[i] = netRule.ServicePort()
		ingRules[i] = netRule.IngressRule(
			cosmov1alpha1.InstanceResourceName(ws.Name, ws.Status.Config.ServiceName))
	}

	scaleTargetRef := func(ws wsv1alpha1.Workspace) cosmov1alpha1.ObjectRef {
		tgt := cosmov1alpha1.ObjectRef{}
		tgt.SetName(ws.Status.Config.DeploymentName)
		tgt.SetGroupVersionKind(kosmo.DeploymentGVK)
		return tgt
	}

	inst.Spec = cosmov1alpha1.InstanceSpec{
		Template: ws.Spec.Template,
		Vars:     addWorkspaceVars(ws.Spec.Vars, ws),
		Override: cosmov1alpha1.OverrideSpec{
			Scale: []cosmov1alpha1.ScalingOverrideSpec{
				{
					Target:   scaleTargetRef(ws),
					Replicas: *ws.Spec.Replicas,
				},
			},
			Network: &cosmov1alpha1.NetworkOverrideSpec{
				Service: []cosmov1alpha1.ServiceOverrideSpec{
					{
						TargetName: ws.Status.Config.ServiceName,
						Ports:      svcPorts,
					},
				},
				Ingress: []cosmov1alpha1.IngressOverrideSpec{
					{
						TargetName: ws.Status.Config.IngressName,
						Rules:      ingRules,
					},
				},
			},
		},
	}

	err := ctrl.SetControllerReference(&ws, inst, r.Scheme)
	if err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	return nil
}

func addWorkspaceVars(vars map[string]string, ws wsv1alpha1.Workspace) map[string]string {
	user := wsv1alpha1.UserIDByNamespace(ws.GetNamespace())

	if vars == nil {
		vars = make(map[string]string)
	}
	// urlvar
	vars[wsnet.URLVarWorkspaceName] = ws.GetName()
	vars[wsnet.URLVarUserID] = user

	// workspace config
	vars[wsv1alpha1.TemplateVarDeploymentName] = ws.Status.Config.DeploymentName
	vars[wsv1alpha1.TemplateVarServiceName] = ws.Status.Config.ServiceName
	vars[wsv1alpha1.TemplateVarIngressName] = ws.Status.Config.IngressName
	vars[wsv1alpha1.TemplateVarServiceMainPortName] = ws.Status.Config.ServiceMainPortName

	return vars
}
