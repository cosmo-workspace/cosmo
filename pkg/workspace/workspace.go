package workspace

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
)

func JSONPatch(op, path string, value any) (string, error) {
	v, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json patch value")
	}
	return fmt.Sprintf(`[{"op": "%s","path": "%s","value": %s}]`, op, path, string(v)), nil
}

func PatchWorkspaceInstanceAsDesired(inst *cosmov1alpha1.Instance, ws cosmov1alpha1.Workspace, scheme *runtime.Scheme) error {
	svcTargetRef := cosmov1alpha1.ObjectRef{}
	svcTargetRef.SetName(ws.Status.Config.ServiceName)
	svcTargetRef.SetGroupVersionKind(kubeutil.ServiceGVK)

	svcPorts := svcPorts(ws.Spec.Network)
	svcPortsPatch, err := JSONPatch("replace", "/spec/ports", svcPorts)
	if err != nil {
		return err
	}

	inst.Spec = cosmov1alpha1.InstanceSpec{
		Template: ws.Spec.Template,
		Vars:     addWorkspaceDefaultVars(ws.Spec.Vars, ws),
		Override: cosmov1alpha1.OverrideSpec{
			PatchesJson6902: []cosmov1alpha1.Json6902{
				{
					Target: svcTargetRef,
					Patch:  svcPortsPatch,
				},
			},
		},
	}

	if ws.Spec.Replicas != nil {
		scaleTargetRef := cosmov1alpha1.ObjectRef{}
		scaleTargetRef.SetName(ws.Status.Config.DeploymentName)
		scaleTargetRef.SetGroupVersionKind(kubeutil.DeploymentGVK)

		scalePatch, err := JSONPatch("replace", "/spec/replicas", ws.Spec.Replicas)
		if err != nil {
			return err
		}

		inst.Spec.Override.PatchesJson6902 = append(inst.Spec.Override.PatchesJson6902, cosmov1alpha1.Json6902{
			Target: scaleTargetRef,
			Patch:  scalePatch,
		})
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
	portMap := make(map[int32]corev1.ServicePort, len(netRules))

	for _, netRule := range netRules {
		port := netRule.ServicePort()
		if _, ok := portMap[port.Port]; ok {
			continue
		}
		portMap[port.Port] = port
		ports = append(ports, port)
	}
	return ports
}

func addWorkspaceDefaultVars(vars map[string]string, ws cosmov1alpha1.Workspace) map[string]string {
	user := cosmov1alpha1.UserNameByNamespace(ws.GetNamespace())

	if vars == nil {
		vars = make(map[string]string)
	}
	// urlvar
	vars[cosmov1alpha1.URLVarWorkspaceName] = ws.GetName()
	vars[cosmov1alpha1.URLVarUserName] = user

	// workspace config
	vars[cosmov1alpha1.WorkspaceTemplateVarDeploymentName] = ws.Status.Config.DeploymentName
	vars[cosmov1alpha1.WorkspaceTemplateVarServiceName] = ws.Status.Config.ServiceName

	return vars
}
