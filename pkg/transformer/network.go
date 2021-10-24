package transformer

import (
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

type NetworkTransformer struct {
	instName string
	netSpec  *cosmov1alpha1.NetworkOverrideSpec
}

func NewNetworkTransformer(netSpec *cosmov1alpha1.NetworkOverrideSpec, instName string) *NetworkTransformer {
	return &NetworkTransformer{netSpec: netSpec, instName: instName}
}

func (t *NetworkTransformer) Transform(src *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	obj := src.DeepCopy()

	if t.netSpec == nil {
		return obj, nil
	}

	for _, ingSpec := range t.netSpec.Ingress {
		if cosmov1alpha1.IsGVKEqual(obj.GroupVersionKind(), kosmo.IngressGVK) && cosmov1alpha1.EqualInstanceResourceName(t.instName, obj.GetName(), ingSpec.TargetName) {
			// Append ingress rules
			overrideIngressRules(obj, ingSpec.Rules)

			// Append or override annotations
			overrideAnnotations(obj, ingSpec.Annotations)
		}
	}

	for _, svcSpec := range t.netSpec.Service {
		if cosmov1alpha1.IsGVKEqual(obj.GroupVersionKind(), kosmo.ServiceGVK) && cosmov1alpha1.EqualInstanceResourceName(t.instName, obj.GetName(), svcSpec.TargetName) {
			// Append service ports
			overrideServicePort(obj, svcSpec.Ports)
		}
	}

	return obj, nil
}

func overrideAnnotations(obj *unstructured.Unstructured, ann map[string]string) {
	if obj == nil {
		return
	}

	objAnn := obj.GetAnnotations()
	if objAnn == nil {
		objAnn = make(map[string]string)
	}
	for key, val := range ann {
		objAnn[key] = val
	}
	obj.SetAnnotations(objAnn)
}

func overrideIngressRules(ingress *unstructured.Unstructured, ingRules []netv1.IngressRule) {
	if ingress == nil {
		return
	}

	if len(ingRules) == 0 {
		return
	}

	var uing netv1.Ingress
	if err := ToObject(ingress.DeepCopy().Object, &uing); err != nil {
		return
	}

	modified := uing.DeepCopy()
	for _, rule := range ingRules {
		ruleFound := false
		for ri, uRule := range uing.Spec.Rules {
			if uRule.Host == rule.Host {
				ruleFound = true

				for _, path := range rule.HTTP.Paths {
					pathFound := false
					for pi, uPath := range uRule.HTTP.Paths {
						if uPath.Path == path.Path && *uPath.PathType == *path.PathType {
							pathFound = true
							modified.Spec.Rules[ri].HTTP.Paths[pi] = path
						}
					}
					if !pathFound {
						modified.Spec.Rules[ri].HTTP.Paths = append(modified.Spec.Rules[ri].HTTP.Paths, path)
					}
				}
			}
		}
		if !ruleFound {
			modified.Spec.Rules = append(modified.Spec.Rules, rule)
		}
	}
	newObj, err := ToUnstructured(modified)
	if err == nil {
		NestedMapDelete(newObj, "metadata.creationTimestamp")
		NestedMapDelete(newObj, "status")
		ingress.Object = newObj
	}
}

func overrideServicePort(svc *unstructured.Unstructured, svcPorts []corev1.ServicePort) {
	if svc == nil {
		return
	}

	if len(svcPorts) == 0 {
		return
	}

	if spec, ok := NestedMap(svc.Object, "spec"); ok {
		if ports, ok := NestedSlice(spec, "ports"); ok {
			for _, v := range svcPorts {
				obj, err := ToUnstructured(&v)
				if err == nil {
					found := false
					for i, p := range ports {
						pm, ok := p.(map[string]interface{})
						if ok && pm["name"] == obj["name"] {
							ports[i] = obj
							found = true
						}
					}
					if !found {
						ports = append(ports, obj)
					}
				}
			}
			spec["ports"] = ports
		}
	}
}
