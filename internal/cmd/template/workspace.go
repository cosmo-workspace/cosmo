package template

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"

	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/internal/authproxy"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

const (
	AuthProxyPatchFile = "cosmo-auth-proxy-patch.yaml"
	AuthProxyRoleBFile = "cosmo-auth-proxy-roleb.yaml"
)

func completeWorkspaceConfig(wsConfig *wsv1alpha1.Config, unst []unstructured.Unstructured) error {
	if wsConfig == nil || len(unst) == 0 {
		return errors.New("invalid args")
	}

	dps := make([]unstructured.Unstructured, 0)
	svcs := make([]unstructured.Unstructured, 0)
	ings := make([]unstructured.Unstructured, 0)
	for _, u := range unst {
		if kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.DeploymentGVK) {
			dps = append(dps, u)
		} else if kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.ServiceGVK) {
			svcs = append(svcs, u)
		} else if kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.IngressGVK) {
			ings = append(ings, u)
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
	var validDep, validSvc, validIng bool
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

	if len(ings) > 0 {
		// complete ingress name
		if wsConfig.IngressName == "" {
			if len(ings) == 1 {
				wsConfig.IngressName = ings[0].GetName()

			} else {
				return errors.New("failed to specify the ingress")
			}
		}

		// validate ingress
		for _, v := range ings {
			if wsConfig.IngressName == v.GetName() {
				var ing netv1.Ingress
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(v.Object, &ing)
				if err != nil {
					return err
				}

				for _, rule := range ing.Spec.Rules {
					for _, path := range rule.HTTP.Paths {
						if path.Backend.Service == nil {
							continue
						}
						if !instance.EqualInstanceResourceName(template.DefaultVarsInstance,
							path.Backend.Service.Name, wsConfig.ServiceName) {
							continue
						}

						if path.Backend.Service.Port.Name != "" {
							if path.Backend.Service.Port.Name == wsConfig.ServiceMainPortName {
								validIng = true
								break
							}
						} else {
							if path.Backend.Service.Port.Number == mainServicePort {
								validIng = true
								break
							}
						}
					}
				}
				break
			}
		}
		if !validIng {
			return fmt.Errorf("ingress '%s' is not found", wsConfig.IngressName)
		}
	}
	return nil
}

func deploymentAuthProxyPatch(injectDeploymentName string, authProxyImage string, tlsSecretName string) *appsv1apply.DeploymentApplyConfiguration {
	applydeploy := (&appsv1apply.DeploymentApplyConfiguration{}).
		WithAPIVersion("apps/v1").
		WithKind("Deployment").
		WithName(injectDeploymentName).
		WithSpec(appsv1apply.DeploymentSpec().
			WithTemplate(corev1apply.PodTemplateSpec().
				WithSpec(corev1apply.PodSpec().
					WithContainers(corev1apply.Container().
						WithName("cosmo-auth-proxy").
						WithImage(authProxyImage).
						WithEnv(
							corev1apply.EnvVar().
								WithName(authproxy.EnvInstance).
								WithValue(template.DefaultVarsInstance),
							corev1apply.EnvVar().
								WithName(authproxy.EnvNamespace).
								WithValue(template.DefaultVarsNamespace))))))
	if tlsSecretName == "" {
		applydeploy.Spec.Template.Spec.Containers[0].WithArgs("--insecure")

	} else {
		applydeploy.Spec.Template.Spec.Containers[0].WithArgs(
			"--tls-cert=/app/cert/tls.crt",
			"--tls-key=/app/cert/tls.key")

		applydeploy.Spec.Template.Spec.Containers[0].WithVolumeMounts(corev1apply.VolumeMount().
			WithMountPath("/app/cert").
			WithName("cert").
			WithReadOnly(true))
		applydeploy.Spec.Template.Spec.WithVolumes(corev1apply.Volume().
			WithName("cert").
			WithSecret(corev1apply.SecretVolumeSource().
				WithDefaultMode(420).
				WithSecretName(tlsSecretName)))
	}

	return applydeploy
}
