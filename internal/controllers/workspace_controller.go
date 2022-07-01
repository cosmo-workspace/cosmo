package controllers

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/wscfg"
	"github.com/cosmo-workspace/cosmo/pkg/wsnet"
)

// WorkspaceReconciler reconciles a Workspace object
type WorkspaceReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

//+kubebuilder:rbac:groups=workspace.cosmo.cosmo-workspace.github.io,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=workspace.cosmo.cosmo-workspace.github.io,resources=workspaces/status,verbs=get;update;patch
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.FromContext(ctx).WithName("WorkspaceReconciler")
	ctx = clog.IntoContext(ctx, log)

	log.Debug().Info("start reconcile")

	var ws wsv1alpha1.Workspace
	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	currentWs := ws.DeepCopy()

	log.DebugAll().DumpObject(r.Scheme, currentWs, "request object")

	// sync workspace config with template
	cfg, err := getWorkspaceConfig(ctx, r.Client, ws.Spec.Template.Name)
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

	now := metav1.Now()
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
		UpdateTimestamp:   &now,
	}

	switch op {
	case controllerutil.OperationResultCreated:
		r.Recorder.Eventf(&ws, corev1.EventTypeNormal, "Created", "successfully instance created")
		ws.Status.Instance.CreationTimestamp = &now

	case controllerutil.OperationResultUpdated:
		r.Recorder.Eventf(&ws, corev1.EventTypeNormal, "Updated", "instance is not desired state, updated")
	}

	// update workspace status
	if !equality.Semantic.DeepEqual(currentWs, ws) {
		if err := r.Status().Update(ctx, &ws); err != nil {
			return ctrl.Result{}, err
		}
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
			instance.InstanceResourceName(ws.Name, ws.Status.Config.ServiceName))
	}

	scaleTargetRef := func(ws wsv1alpha1.Workspace) cosmov1alpha1.ObjectRef {
		tgt := cosmov1alpha1.ObjectRef{}
		tgt.SetName(ws.Status.Config.DeploymentName)
		tgt.SetGroupVersionKind(kubeutil.DeploymentGVK)
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

func getWorkspaceServicesAndIngress(ctx context.Context, c client.Client, ws wsv1alpha1.Workspace) (svc corev1.Service, ing netv1.Ingress, err error) {
	var svcList corev1.ServiceList
	var ingList netv1.IngressList

	ls := labels.NewSelector()
	req, _ := labels.NewRequirement(cosmov1alpha1.LabelKeyInstance, selection.In, []string{ws.GetName()})
	ls = ls.Add(*req)

	opts := &client.ListOptions{
		LabelSelector: ls,
		Namespace:     ws.GetNamespace(),
	}

	if err := c.List(ctx, &svcList, opts); err != nil {
		return svc, ing, err
	}

	if len(svcList.Items) == 0 {
		return svc, ing, errors.New("no services")
	}

	for _, v := range svcList.Items {
		if instance.EqualInstanceResourceName(ws.GetName(), v.Name, ws.Status.Config.ServiceName) {
			svc = v
		}
	}

	if err := c.List(ctx, &ingList, opts); err != nil {
		return svc, ing, err
	}

	for _, v := range ingList.Items {
		if instance.EqualInstanceResourceName(ws.GetName(), v.Name, ws.Status.Config.IngressName) {
			ing = v
		}
	}

	return svc, ing, nil
}

func getWorkspaceConfig(ctx context.Context, c client.Client, tmplName string) (cfg wsv1alpha1.Config, err error) {
	tmpl := &cosmov1alpha1.Template{}
	if err := c.Get(ctx, types.NamespacedName{Name: tmplName}, tmpl); err != nil {
		return cfg, err
	}
	return wscfg.ConfigFromTemplateAnnotations(tmpl)
}
