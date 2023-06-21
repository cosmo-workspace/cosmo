package controllers

import (
	"context"
	"reflect"
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

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
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
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: nginx
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
      cosmo-workspace.github.io/instance: '{{INSTANCE}}'
  template:
    metadata:
      labels:
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
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
			Ω(ObjectSnapshot(&deploy)).To(MatchSnapShot())

			By("checking if child service is as expected")
			var svc corev1.Service
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(inst.Name, "svc"),
					Namespace: inst.Namespace,
				}
				return k8sClient.Get(ctx, key, &svc)
			}, time.Second*10).Should(Succeed())
			Ω(ServiceSnapshot(&svc)).To(MatchSnapShot())

			By("checking if child ingress is as expected")
			var ing netv1.Ingress
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(inst.Name, "ing"),
					Namespace: inst.Namespace,
				}
				return k8sClient.Get(ctx, key, &ing)
			}, time.Second*10).Should(Succeed())
			Ω(ObjectSnapshot(&ing)).To(MatchSnapShot())

			By("fetching instance resource and checking if last applied resources added in instance status")

			var createdInst cosmov1alpha1.Instance
			Eventually(func() int {
				key := client.ObjectKey{
					Name:      inst.Name,
					Namespace: inst.Namespace,
				}
				err = k8sClient.Get(ctx, key, &createdInst)
				Expect(err).ShouldNot(HaveOccurred())

				return createdInst.Status.LastAppliedObjectsCount
			}, time.Second*10).ShouldNot(BeZero())
			Ω(InstanceSnapshot(&createdInst)).To(MatchSnapShot())
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
				curInst.Spec.Override = cosmov1alpha1.OverrideSpec{
					PatchesJson6902: []cosmov1alpha1.Json6902{
						{
							Target: cosmov1alpha1.ObjectRef{
								ObjectReference: corev1.ObjectReference{
									APIVersion: "apps/v1",
									Kind:       "Deployment",
									Name:       "deploy",
								},
							},
							Patch: `[{"op": "replace", "path": "/spec/replicas", "value": 3}]`,
						},
						{
							Target: cosmov1alpha1.ObjectRef{
								ObjectReference: corev1.ObjectReference{
									APIVersion: "v1",
									Kind:       "Service",
									Name:       "svc",
								},
							},
							Patch: `[{"op": "replace", "path": "/spec/ports", "value": [{"name": "add", "port": 9090, "protocol": "TCP"},{"name": "add2", "port": 9091, "protocol": "TCP"}]}]`,
						},
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
							Patch: `[{"op": "replace", "path": "/spec/type", "value": "LoadBalancer"}]`,
						},
					},
				}
				return k8sClient.Update(ctx, &curInst)
			}, time.Second*60).Should(Succeed())
			Ω(InstanceSnapshot(&curInst)).To(MatchSnapShot())

			var updatedInst cosmov1alpha1.Instance
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      inst.Name,
					Namespace: inst.Namespace,
				}
				return k8sClient.Get(ctx, key, &updatedInst)
			}, time.Second*10).Should(Succeed())
			Ω(InstanceSnapshot(&updatedInst)).To(MatchSnapShot())

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
			Ω(ObjectSnapshot(&deploy)).To(MatchSnapShot())

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
			Ω(ServiceSnapshot(&svc)).To(MatchSnapShot())

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
		obj *unstructured.Unstructured
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
