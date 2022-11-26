package workspace

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/wsnet"
	netv1 "k8s.io/api/networking/v1"
)

func PatchWorkspaceInstanceAsDesired(inst *cosmov1alpha1.Instance, ws cosmov1alpha1.Workspace, scheme *runtime.Scheme) error {
	backendSvcName := instance.InstanceResourceName(ws.Name, ws.Status.Config.ServiceName)
	svcPorts := svcPorts(ws.Spec.Network)
	ingRules := ingressRules(ws.Spec.Network, backendSvcName)

	scaleTargetRef := func(ws cosmov1alpha1.Workspace) cosmov1alpha1.ObjectRef {
		tgt := cosmov1alpha1.ObjectRef{}
		tgt.SetName(ws.Status.Config.DeploymentName)
		tgt.SetGroupVersionKind(kubeutil.DeploymentGVK)
		return tgt
	}

	inst.Spec = cosmov1alpha1.InstanceSpec{
		Template: ws.Spec.Template,
		Vars:     addWorkspaceDefaultVars(ws.Spec.Vars, ws),
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

	if scheme != nil {
		err := ctrl.SetControllerReference(&ws, inst, scheme)
		if err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}
	}

	return nil
}

func svcPorts(netRules []cosmov1alpha1.NetworkRule) []corev1.ServicePort {
	ports := make([]corev1.ServicePort, 0, len(netRules))
	portMap := make(map[string]corev1.ServicePort, len(netRules))

	for _, netRule := range netRules {
		if !netRule.TargetPortNumberIsValid() {
			continue
		}
		port := netRule.ServicePort()
		if _, ok := portMap[port.Name]; ok {
			continue
		}
		portMap[port.Name] = port
		ports = append(ports, port)
	}
	return ports
}

func ingressRules(netRules []cosmov1alpha1.NetworkRule, backendSvcName string) []netv1.IngressRule {
	ingRules := make([]netv1.IngressRule, 0, len(netRules))
	ingRuleMap := make(map[string]netv1.IngressRule, len(netRules))

	for _, netRule := range netRules {
		if !netRule.TargetPortNumberIsValid() {
			continue
		}
		ingRule := netRule.IngressRule(backendSvcName)
		// Merge rules for the same host
		if r, ok := ingRuleMap[ingRule.Host]; ok {
			r.IngressRuleValue.HTTP.Paths = append(r.IngressRuleValue.HTTP.Paths, ingRule.HTTP.Paths[0])
			continue
		}
		ingRuleMap[ingRule.Host] = ingRule
		ingRules = append(ingRules, ingRule)
	}
	return ingRules
}

func addWorkspaceDefaultVars(vars map[string]string, ws cosmov1alpha1.Workspace) map[string]string {
	user := cosmov1alpha1.UserNameByNamespace(ws.GetNamespace())

	if vars == nil {
		vars = make(map[string]string)
	}
	// urlvar
	vars[wsnet.URLVarWorkspaceName] = ws.GetName()
	vars[wsnet.URLVarUserName] = user

	// workspace config
	vars[cosmov1alpha1.WorkspaceTemplateVarDeploymentName] = ws.Status.Config.DeploymentName
	vars[cosmov1alpha1.WorkspaceTemplateVarServiceName] = ws.Status.Config.ServiceName
	vars[cosmov1alpha1.WorkspaceTemplateVarIngressName] = ws.Status.Config.IngressName
	vars[cosmov1alpha1.WorkspaceTemplateVarServiceMainPortName] = ws.Status.Config.ServiceMainPortName

	return vars
}
