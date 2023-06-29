package template

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
)

func completeWorkspaceConfig(wsConfig *cosmov1alpha1.Config, unst []unstructured.Unstructured) error {
	if wsConfig == nil || len(unst) == 0 {
		return errors.New("invalid args")
	}

	dps := make([]unstructured.Unstructured, 0)
	svcs := make([]unstructured.Unstructured, 0)

	for _, u := range unst {
		if kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.DeploymentGVK) {
			dps = append(dps, u)
		} else if kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.ServiceGVK) {
			svcs = append(svcs, u)
		}
	}

	// complete deployment name
	if wsConfig.DeploymentName == "" {
		if len(dps) != 1 {
			return errors.New("no deployment")
		}
		wsConfig.DeploymentName = dps[0].GetName()
	}

	// validate deployment
	var validDep, validSvc bool
	for _, v := range dps {
		if wsConfig.DeploymentName == v.GetName() {
			validDep = true
		}
	}
	if !validDep {
		return fmt.Errorf("deployment '%s' is not found", wsConfig.DeploymentName)
	}

	// complete service name
	if wsConfig.ServiceName == "" {
		if len(svcs) != 1 {
			return errors.New("no service")
		}
		wsConfig.ServiceName = svcs[0].GetName()
	}

	// validate service
	var svc corev1.Service
	for _, v := range svcs {
		if wsConfig.ServiceName == v.GetName() {
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(v.Object, &svc)
			if err != nil {
				return err
			}
			validSvc = true
		}
	}
	if !validSvc {
		return fmt.Errorf("service '%s' is not found", wsConfig.ServiceName)
	}

	// complete service main port
	if wsConfig.ServiceMainPortName == "" {
		if len(svc.Spec.Ports) != 1 {
			return errors.New("failed to specify the service port")
		}
		wsConfig.ServiceMainPortName = svc.Spec.Ports[0].Name
	}

	// validate service main port
	var mainServicePort int32
	for _, port := range svc.Spec.Ports {
		if port.Name == wsConfig.ServiceMainPortName {
			mainServicePort = port.Port
		}
	}
	if mainServicePort == 0 {
		return fmt.Errorf("service '%s' is not found", wsConfig.ServiceName)
	}

	return nil
}
