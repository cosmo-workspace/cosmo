package controllers

import (
	"context"
	"reflect"
	"sort"
	"testing"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
)

var _ = Describe("Instance controller", func() {
	tmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-nginx-tmpl1",
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ing
  namespace: '{{NAMESPACE}}'
spec:
  rules:
  - host: '{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}'
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: '{{INSTANCE}}-nginx'
            port: 
              number: 80
---
apiVersion: v1
kind: Service
metadata:
  name: svc
  namespace: '{{NAMESPACE}}'
spec:
  ports:
  - name: main
    port: 80
    protocol: TCP
  selector:
    cosmo/instance: '{{INSTANCE}}'
    cosmo/template: nginx
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
  namespace: '{{NAMESPACE}}'
spec:
  replicas: 1
  selector:
    matchLabels:
      cosmo/instance: '{{INSTANCE}}'
  template:
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
    spec:
      containers:
      - image: 'nginx:{{IMAGE_TAG}}'
        name: nginx
        ports:
        - containerPort: 80
          name: main
          protocol: TCP
`,
			// RequiredVars: []string{"{{DOMAIN}}", "{{IMAGE_TAG}}"},
			RequiredVars: []cosmov1alpha1.RequiredVarSpec{
				{
					Var: "{{DOMAIN}}",
				},
				{
					Var: "{{IMAGE_TAG}}",
				},
			},
		},
	}

	inst := cosmov1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-inst1",
			Namespace: "default",
		},
		Spec: cosmov1alpha1.InstanceSpec{
			Template: cosmov1alpha1.TemplateRef{
				Name: tmpl.GetName(),
			},
			Override: cosmov1alpha1.OverrideSpec{},
			Vars: map[string]string{
				"{{DOMAIN}}":    "example.com",
				"{{IMAGE_TAG}}": "latest",
			},
		},
	}

	Context("when creating Template resource on new cluster", func() {
		It("should do nothing", func() {
			ctx := context.Background()

			By("creating template before instance")

			err := k8sClient.Create(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(3 * time.Second)

			var createdTmpl cosmov1alpha1.Template
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: tmpl.GetName()}, &createdTmpl)
			}, time.Second*10).Should(Succeed())
		})
	})

	Context("when creating a Instance resource", func() {
		It("should do reconcile once and create child resources", func() {
			ctx := context.Background()

			err := k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())

			By("checking if child deployment is as expected")
			var deploy appsv1.Deployment
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(inst.Name, "deploy"),
					Namespace: inst.Namespace,
				}
				return k8sClient.Get(ctx, key, &deploy)
			}, time.Second*10).Should(Succeed())
			Ω(objectSnapshot(&deploy)).To(MatchSnapShot())

			By("checking if child service is as expected")
			var svc corev1.Service
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(inst.Name, "svc"),
					Namespace: inst.Namespace,
				}
				return k8sClient.Get(ctx, key, &svc)
			}, time.Second*10).Should(Succeed())
			Ω(serviceSnapshot(&svc)).To(MatchSnapShot())

			By("checking if child ingress is as expected")
			var ing netv1.Ingress
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(inst.Name, "ing"),
					Namespace: inst.Namespace,
				}
				return k8sClient.Get(ctx, key, &ing)
			}, time.Second*10).Should(Succeed())
			Ω(objectSnapshot(&ing)).To(MatchSnapShot())

			By("fetching instance resource and checking if last applied resources added in instance status")

			var createdInst cosmov1alpha1.Instance
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      inst.Name,
					Namespace: inst.Namespace,
				}
				return k8sClient.Get(ctx, key, &createdInst)
			}, time.Second*10).Should(Succeed())
			Ω(instanceSnapshot(&createdInst)).To(MatchSnapShot())
		})
	})

	Context("when updating Instance resource", func() {
		It("should do reconcile again and update child resources", func() {
			ctx := context.Background()

			// fetch current instance
			var curInst cosmov1alpha1.Instance
			Eventually(func() error {
				key := types.NamespacedName{
					Name:      inst.Name,
					Namespace: inst.Namespace,
				}
				err := k8sClient.Get(ctx, key, &curInst)
				Expect(err).NotTo(HaveOccurred())

				// update instance override spec
				prefix := netv1.PathTypePrefix
				curInst.Spec.Override = cosmov1alpha1.OverrideSpec{
					Scale: []cosmov1alpha1.ScalingOverrideSpec{
						{
							Target: cosmov1alpha1.ObjectRef{
								ObjectReference: corev1.ObjectReference{
									APIVersion: metav1.GroupVersion{
										Group:   "apps",
										Version: "v1",
									}.String(),
									Kind: "Deployment",
									Name: "deploy",
								},
							},
							Replicas: 3,
						},
					},
					Network: &cosmov1alpha1.NetworkOverrideSpec{
						Service: []cosmov1alpha1.ServiceOverrideSpec{
							{
								TargetName: "svc",
								Ports: []corev1.ServicePort{
									{
										Name:     "add",
										Port:     9090,
										Protocol: corev1.ProtocolTCP,
									},
								},
							},
						},
						Ingress: []cosmov1alpha1.IngressOverrideSpec{
							{
								TargetName: "ing",
								Rules: []netv1.IngressRule{
									{
										Host: "add.example.com",
										IngressRuleValue: netv1.IngressRuleValue{
											HTTP: &netv1.HTTPIngressRuleValue{
												Paths: []netv1.HTTPIngressPath{
													{
														Path:     "/add",
														PathType: &prefix,
														Backend: netv1.IngressBackend{
															Service: &netv1.IngressServiceBackend{
																Name: "svc",
																Port: netv1.ServiceBackendPort{
																	Number: 9090,
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					PatchesJson6902: []cosmov1alpha1.Json6902{
						{
							Target: cosmov1alpha1.ObjectRef{
								ObjectReference: corev1.ObjectReference{
									APIVersion: metav1.GroupVersion{
										Group:   "",
										Version: "v1",
									}.String(),
									Kind: "Service",
									Name: "svc",
								},
							},
							Patch: `
[
  {
    "op": "replace",
    "path": "/spec/type",
    "value": "LoadBalancer"
  }
]
						`,
						},
					},
				}
				return k8sClient.Update(ctx, &curInst)
			}, time.Second*60).Should(Succeed())
			Ω(instanceSnapshot(&curInst)).To(MatchSnapShot())

			By("checking if child deployment is as expected")
			var deploy appsv1.Deployment
			Eventually(func() int32 {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(inst.Name, "deploy"),
					Namespace: inst.Namespace,
				}
				err := k8sClient.Get(ctx, key, &deploy)
				Expect(err).ShouldNot(HaveOccurred())

				return *deploy.Spec.Replicas
			}, time.Second*10).Should(Equal(int32(3)))
			Ω(objectSnapshot(&deploy)).To(MatchSnapShot())

			By("checking if child service is as expected")
			var svc corev1.Service
			Eventually(func() corev1.ServiceType {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(inst.Name, "svc"),
					Namespace: inst.Namespace,
				}
				err := k8sClient.Get(ctx, key, &svc)
				Expect(err).ShouldNot(HaveOccurred())

				return svc.Spec.Type
			}, time.Second*10).Should(Equal(corev1.ServiceTypeLoadBalancer))
			Ω(serviceSnapshot(&svc)).To(MatchSnapShot())

			By("checking if child ingress is as expected")
			var ing netv1.Ingress
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(inst.Name, "ing"),
					Namespace: inst.Namespace,
				}
				return k8sClient.Get(ctx, key, &ing)
			}, time.Second*10).Should(Succeed())
			Ω(objectSnapshot(&ing)).To(MatchSnapShot())

			var updatedInst cosmov1alpha1.Instance
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      inst.Name,
					Namespace: inst.Namespace,
				}
				return k8sClient.Get(ctx, key, &updatedInst)
			}, time.Second*10).Should(Succeed())
			Ω(instanceSnapshot(&updatedInst)).To(MatchSnapShot())
		})
	})

	Context("when creating clusterrole", func() {
		It("should not create cluster-scope resource", func() {
			ctx := context.Background()

			clusterLevelTmpl := cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-level-tmpl",
				},
				Spec: cosmov1alpha1.TemplateSpec{
					RawYaml: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: privileged
  namespace: {{NAMESPACE}}
rules:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - '*'
- nonResourceURLs:
  - '*'
  verbs:
  - '*'
`,
				},
			}
			err := k8sClient.Create(ctx, &clusterLevelTmpl)
			Expect(err).ShouldNot(HaveOccurred())

			inst := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cluster-level-inst",
					Namespace: "default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: clusterLevelTmpl.Name,
					},
				},
			}
			err = k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(time.Second * 3)

			var createdInst cosmov1alpha1.Instance
			key := client.ObjectKey{
				Name:      inst.Name,
				Namespace: inst.Namespace,
			}
			err = k8sClient.Get(ctx, key, &createdInst)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(createdInst.Status.LastAppliedObjectsCount).Should(BeZero())

			var cr rbacv1.ClusterRole
			key = client.ObjectKey{
				Name: instance.InstanceResourceName(inst.Name, "privileged"),
			}
			err = k8sClient.Get(ctx, key, &cr)
			Expect(apierrs.IsNotFound(err)).Should(BeTrue())
		})
	})
})

func Test_unstToObjectRef(t *testing.T) {
	creationTimestamp := "2021-07-13T01:50:08Z"

	creationTime, _ := time.Parse("2006-01-02T03:04:05Z", creationTimestamp)
	creationTime = creationTime.Local()
	metaCreationTime := metav1.NewTime(creationTime)

	type args struct {
		obj             *unstructured.Unstructured
		updateTimestamp metav1.Time
	}
	tests := []struct {
		name string
		args args
		want cosmov1alpha1.ObjectRef
	}{
		{
			name: "OK",
			args: args{
				obj: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "networking.k8s.io/v1",
						"kind":       "Ingress",
						"metadata": map[string]interface{}{
							"name":              "test",
							"namespace":         "default",
							"creationTimestamp": "2021-07-13T01:50:08Z",
						},
					},
				},
			},
			want: cosmov1alpha1.ObjectRef{
				ObjectReference: corev1.ObjectReference{
					APIVersion: "networking.k8s.io/v1",
					Kind:       "Ingress",
					Name:       "test",
					Namespace:  "default",
				},
				CreationTimestamp: &metaCreationTime,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unstToObjectRef(tt.args.obj); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unstToObjectRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func instanceSnapshot(in cosmov1alpha1.InstanceObject) cosmov1alpha1.InstanceObject {
	o := in.DeepCopyObject()
	obj := o.(cosmov1alpha1.InstanceObject)
	removeDynamicFields(obj)

	for i, v := range obj.GetStatus().LastApplied {
		v.CreationTimestamp = nil
		v.UID = ""
		v.ResourceVersion = ""
		obj.GetStatus().LastApplied[i] = v
	}
	sort.Slice(obj.GetStatus().LastApplied, func(i, j int) bool {
		return obj.GetStatus().LastApplied[i].Kind < obj.GetStatus().LastApplied[j].Kind
	})
	sort.Slice(obj.GetStatus().LastApplied, func(i, j int) bool {
		return obj.GetStatus().LastApplied[i].Name < obj.GetStatus().LastApplied[j].Name
	})
	obj.GetStatus().TemplateResourceVersion = ""

	return obj
}

func serviceSnapshot(in *corev1.Service) *corev1.Service {
	obj := in.DeepCopy()
	removeDynamicFields(obj)

	obj.Spec.ClusterIP = ""
	obj.Spec.ClusterIPs = nil

	for i, p := range obj.Spec.Ports {
		if p.NodePort >= 30000 {
			obj.Spec.Ports[i].NodePort = 30000
		}
	}

	return obj
}

func objectSnapshot(obj client.Object) client.Object {
	t := obj.DeepCopyObject()
	o := t.(client.Object)
	removeDynamicFields(o)
	return o
}

func removeDynamicFields(o client.Object) {
	o.SetCreationTimestamp(metav1.Time{})
	o.SetResourceVersion("")
	o.SetGeneration(0)
	o.SetUID(types.UID(""))
	o.SetManagedFields(nil)

	ownerRefs := make([]metav1.OwnerReference, len(o.GetOwnerReferences()))
	for i, v := range o.GetOwnerReferences() {
		v.UID = ""
		ownerRefs[i] = v
	}
	o.SetOwnerReferences(ownerRefs)
}
