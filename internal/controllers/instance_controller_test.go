package controllers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"

	appsv1apply "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1apply "k8s.io/client-go/applyconfigurations/core/v1"
	metav1apply "k8s.io/client-go/applyconfigurations/meta/v1"
	netv1apply "k8s.io/client-go/applyconfigurations/networking/v1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
)

var _ = Describe("Instance controller", func() {
	const tmplName string = "nginx-test"
	const instName string = "inst-test"
	const nsName string = "default"

	tmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: tmplName,
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  labels:
    cosmo/instance: '{{INSTANCE}}'
    cosmo/template: nginx
  name: nginx
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
  labels:
    cosmo/instance: '{{INSTANCE}}'
    cosmo/template: nginx
  name: nginx
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
  labels:
    cosmo/instance: '{{INSTANCE}}'
    cosmo/template: nginx
  name: nginx
  namespace: '{{NAMESPACE}}'
spec:
  replicas: 1
  selector:
    matchLabels:
      cosmo/instance: '{{INSTANCE}}'
      cosmo/template: nginx
  template:
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: nginx
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

	expectedDeployApply := func(instName, namespace, imageTag string, ownerRef metav1.OwnerReference) *appsv1apply.DeploymentApplyConfiguration {
		return appsv1apply.Deployment(instance.InstanceResourceName(instName, "nginx"), namespace).
			WithAPIVersion("apps/v1").
			WithKind("Deployment").
			WithLabels(map[string]string{
				cosmov1alpha1.LabelKeyInstance: instName,
				cosmov1alpha1.LabelKeyTemplate: tmplName,
			}).
			WithOwnerReferences(
				metav1apply.OwnerReference().
					WithAPIVersion(ownerRef.APIVersion).
					WithBlockOwnerDeletion(*ownerRef.BlockOwnerDeletion).
					WithController(*ownerRef.Controller).
					WithKind(ownerRef.Kind).
					WithName(ownerRef.Name).
					WithUID(ownerRef.UID),
			).
			WithSpec(appsv1apply.DeploymentSpec().
				WithReplicas(1).
				WithSelector(metav1apply.LabelSelector().
					WithMatchLabels(map[string]string{
						cosmov1alpha1.LabelKeyInstance: instName,
						cosmov1alpha1.LabelKeyTemplate: "nginx",
					})).
				WithTemplate(corev1apply.PodTemplateSpec().
					WithLabels(map[string]string{
						cosmov1alpha1.LabelKeyInstance: instName,
						cosmov1alpha1.LabelKeyTemplate: "nginx",
					}).
					WithSpec(corev1apply.PodSpec().
						WithContainers(corev1apply.Container().
							WithName("nginx").
							WithImage("nginx:" + imageTag).
							WithPorts(
								corev1apply.ContainerPort().
									WithName("main").
									WithProtocol(corev1.ProtocolTCP).
									WithContainerPort(80))))))
	}

	expectedServiceApply := func(instName, namespace string, ownerRef metav1.OwnerReference) *corev1apply.ServiceApplyConfiguration {
		return corev1apply.Service(instance.InstanceResourceName(instName, "nginx"), namespace).
			WithAPIVersion("v1").
			WithKind("Service").
			WithLabels(map[string]string{
				cosmov1alpha1.LabelKeyInstance: instName,
				cosmov1alpha1.LabelKeyTemplate: tmplName,
			}).
			WithOwnerReferences(
				metav1apply.OwnerReference().
					WithAPIVersion(ownerRef.APIVersion).
					WithBlockOwnerDeletion(*ownerRef.BlockOwnerDeletion).
					WithController(*ownerRef.Controller).
					WithKind(ownerRef.Kind).
					WithName(ownerRef.Name).
					WithUID(ownerRef.UID),
			).
			WithSpec(corev1apply.ServiceSpec().
				WithPorts(corev1apply.ServicePort().
					WithName("main").
					WithPort(int32(80)).
					WithProtocol(corev1.ProtocolTCP)).
				WithSelector(map[string]string{
					cosmov1alpha1.LabelKeyInstance: instName,
					cosmov1alpha1.LabelKeyTemplate: "nginx",
				}).
				WithType(corev1.ServiceTypeClusterIP))
	}

	expectedIngressApply := func(instName, namespace, domain string, ownerRef metav1.OwnerReference) *netv1apply.IngressApplyConfiguration {
		return netv1apply.Ingress(instance.InstanceResourceName(instName, "nginx"), namespace).
			WithAPIVersion("networking.k8s.io/v1").
			WithKind("Ingress").
			WithLabels(map[string]string{
				cosmov1alpha1.LabelKeyInstance: instName,
				cosmov1alpha1.LabelKeyTemplate: tmplName,
			}).
			WithOwnerReferences(
				metav1apply.OwnerReference().
					WithAPIVersion(ownerRef.APIVersion).
					WithBlockOwnerDeletion(*ownerRef.BlockOwnerDeletion).
					WithController(*ownerRef.Controller).
					WithKind(ownerRef.Kind).
					WithName(ownerRef.Name).
					WithUID(ownerRef.UID),
			).
			WithSpec(netv1apply.IngressSpec().
				WithRules(netv1apply.IngressRule().
					WithHost(fmt.Sprintf("%s-%s.%s", instName, namespace, domain)).
					WithHTTP(netv1apply.HTTPIngressRuleValue().
						WithPaths(netv1apply.HTTPIngressPath().
							WithPath("/").
							WithPathType(netv1.PathTypePrefix).
							WithBackend(netv1apply.IngressBackend().
								WithService(netv1apply.IngressServiceBackend().
									WithName(instance.InstanceResourceName(instName, "nginx")).
									WithPort(netv1apply.ServiceBackendPort().
										WithNumber(80))))))))
	}

	Context("when creating Template resource on new cluster", func() {
		It("should do nothing", func() {
			ctx := context.Background()

			By("creating template before instance")

			err := k8sClient.Create(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			var createdTmpl cosmov1alpha1.Template
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: tmplName}, &createdTmpl)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())
		})
	})

	Context("when creating a Instance resource", func() {
		It("should do reconcile once and create child resources", func() {
			ctx := context.Background()

			inst := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      instName,
					Namespace: nsName,
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: tmplName,
					},
					Override: cosmov1alpha1.OverrideSpec{},
					Vars: map[string]string{
						"{{DOMAIN}}":    "example.com",
						"{{IMAGE_TAG}}": "latest",
					},
				},
			}
			err := k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())

			By("fetching instance resource and checking if last applied resources added in instance status")

			var createdInst cosmov1alpha1.Instance
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instName,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &createdInst)
				if err != nil {
					return err
				}
				if len(createdInst.Status.LastApplied) == 0 {
					return errors.New("child resources still not created")
				}
				return nil
			}, time.Second*10).Should(Succeed())

			By("checking if child resources is as expected in template")

			ownerRef := ownerRef(&inst, scheme.Scheme)

			// Deployment
			var deploy appsv1.Deployment
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(instName, "nginx"),
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &deploy)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())
			deploy.GroupVersionKind()

			deployApplyCfg, err := appsv1apply.ExtractDeployment(&deploy, InstControllerFieldManager)
			Expect(err).ShouldNot(HaveOccurred())

			expectedDeployApplyCfg := expectedDeployApply(instName, nsName, "latest", ownerRef)

			By("checking if deployment is as expected")

			clog.PrintObjectDiff(os.Stderr, deployApplyCfg, expectedDeployApplyCfg)
			eq := equality.Semantic.DeepEqual(deployApplyCfg, expectedDeployApplyCfg)
			Expect(eq).Should(BeTrue())

			deploy.SetGroupVersionKind(kubeutil.DeploymentGVK)

			Expect(instance.ExistInLastApplyed(createdInst, &deploy)).Should(BeTrue())

			// Service
			var svc corev1.Service
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(instName, "nginx"),
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &svc)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			svcApplyCfg, err := corev1apply.ExtractService(&svc, InstControllerFieldManager)
			Expect(err).ShouldNot(HaveOccurred())

			expectedServiceApplyCfg := expectedServiceApply(instName, nsName, ownerRef)

			By("checking if service is as expected")

			clog.PrintObjectDiff(os.Stderr, svcApplyCfg, expectedServiceApplyCfg)
			eq = equality.Semantic.DeepEqual(svcApplyCfg, expectedServiceApplyCfg)
			Expect(eq).Should(BeTrue())

			svc.SetGroupVersionKind(kubeutil.ServiceGVK)

			Expect(instance.ExistInLastApplyed(createdInst, &svc)).Should(BeTrue())

			// Ingress
			var ing netv1.Ingress
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(instName, "nginx"),
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &ing)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			ingApplyCfg, err := netv1apply.ExtractIngress(&ing, InstControllerFieldManager)
			Expect(err).ShouldNot(HaveOccurred())

			expectedIngApplyCfg := expectedIngressApply(instName, nsName, "example.com", ownerRef)

			By("checking if ingress is as expected")

			clog.PrintObjectDiff(os.Stderr, ingApplyCfg, expectedIngApplyCfg)
			eq = equality.Semantic.DeepEqual(ingApplyCfg, expectedIngApplyCfg)
			Expect(eq).Should(BeTrue())

			ing.SetGroupVersionKind(kubeutil.IngressGVK)

			Expect(instance.ExistInLastApplyed(createdInst, &ing)).Should(BeTrue())

			By("checking creation time equal to update time")

			for _, v := range createdInst.Status.LastApplied {
				//fmt.Println("CreationTimestamp", v.CreationTimestamp)
				//fmt.Println("UpdateTimestamp", v.UpdateTimestamp)

				Expect(v.CreationTimestamp.Equal(v.UpdateTimestamp)).Should(BeTrue())
			}
		})
	})

	Context("when updating Instance resource", func() {
		It("should do reconcile again and update child resources", func() {
			ctx := context.Background()

			// fetch current instance
			var inst cosmov1alpha1.Instance
			Eventually(func() error {
				key := types.NamespacedName{
					Name:      instName,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &inst)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			// update instance override spec
			prefix := netv1.PathTypePrefix
			inst.Spec.Override = cosmov1alpha1.OverrideSpec{
				Scale: []cosmov1alpha1.ScalingOverrideSpec{
					{
						Target: cosmov1alpha1.ObjectRef{
							ObjectReference: corev1.ObjectReference{
								APIVersion: metav1.GroupVersion{
									Group:   "apps",
									Version: "v1",
								}.String(),
								Kind: "Deployment",
								Name: "nginx",
							},
						},
						Replicas: 3,
					},
				},
				Network: &cosmov1alpha1.NetworkOverrideSpec{
					Service: []cosmov1alpha1.ServiceOverrideSpec{
						{
							TargetName: "nginx",
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
							TargetName: "nginx",
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
															Name: "nginx",
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
								Name: "nginx",
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

			err := k8sClient.Update(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())

			By("checking if child resources updated")

			ownerRef := ownerRef(&inst, scheme.Scheme)

			// expected Deployment
			expectedDeployApplyCfg := expectedDeployApply(instName, nsName, "latest", ownerRef)
			expectedDeployApplyCfg.Spec.WithReplicas(3)

			// expected Service
			expectedServiceApplyCfg := expectedServiceApply(instName, nsName, ownerRef)
			expectedServiceApplyCfg.Spec.Ports = append(expectedServiceApplyCfg.Spec.Ports, *corev1apply.ServicePort().
				WithName("add").
				WithPort(int32(9090)).
				WithProtocol(corev1.ProtocolTCP).
				WithTargetPort(intstr.FromInt(9090)))
			expectedServiceApplyCfg.Spec.WithType(corev1.ServiceTypeLoadBalancer)

			// expected Ingress
			expectedIngApplyCfg := expectedIngressApply(instName, nsName, "example.com", ownerRef)
			expectedIngApplyCfg.Spec.Rules = append(expectedIngApplyCfg.Spec.Rules, *netv1apply.IngressRule().
				WithHost("add.example.com").
				WithHTTP(netv1apply.HTTPIngressRuleValue().
					WithPaths(netv1apply.HTTPIngressPath().
						WithPath("/add").
						WithPathType(netv1.PathTypePrefix).
						WithBackend(netv1apply.IngressBackend().
							WithService(netv1apply.IngressServiceBackend().
								WithName("nginx").
								WithPort(netv1apply.ServiceBackendPort().
									WithNumber(9090)))))))

			By("checking if deployment updated")

			var deploy appsv1.Deployment
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(instName, "nginx"),
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &deploy)
				if err != nil {
					return err
				}

				deployApplyCfg, err := appsv1apply.ExtractDeployment(&deploy, InstControllerFieldManager)
				Expect(err).ShouldNot(HaveOccurred())

				eq := equality.Semantic.DeepEqual(deployApplyCfg, expectedDeployApplyCfg)
				if !eq {
					return errors.New("not equal")
				}
				return nil
			}, time.Second*10).Should(Succeed())

			By("checking if service updated")

			var svc corev1.Service
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(instName, "nginx"),
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &svc)
				if err != nil {
					return err
				}

				svcApplyCfg, err := corev1apply.ExtractService(&svc, InstControllerFieldManager)
				Expect(err).ShouldNot(HaveOccurred())

				eq := equality.Semantic.DeepEqual(svcApplyCfg, expectedServiceApplyCfg)
				if !eq {
					return errors.New("not equal")
				}
				return nil
			}, time.Second*10).Should(Succeed())

			By("checking if ingress updated")

			var ing netv1.Ingress
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      instance.InstanceResourceName(instName, "nginx"),
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &ing)
				if err != nil {
					return err
				}

				ingApplyCfg, err := netv1apply.ExtractIngress(&ing, InstControllerFieldManager)
				Expect(err).ShouldNot(HaveOccurred())

				eq := equality.Semantic.DeepEqual(ingApplyCfg, expectedIngApplyCfg)
				if !eq {
					return errors.New("not equal")
				}

				return nil
			}, time.Second*10).Should(Succeed())
		})
	})

	Context("when creating clusterrole", func() {
		It("should not create cluster-scope resource", func() {
			ctx := context.Background()

			clusterLevelTmplName := "cluster-level-tmpl"
			clusterLevelTmpl := cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterLevelTmplName,
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

			clusterLevelInstName := "cluster-level-inst"
			clusterLevelInst := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterLevelInstName,
					Namespace: nsName,
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: clusterLevelTmplName,
					},
				},
			}
			err = k8sClient.Create(ctx, &clusterLevelInst)
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(time.Second * 3)

			var createdInst cosmov1alpha1.Instance
			key := client.ObjectKey{
				Name:      clusterLevelInstName,
				Namespace: nsName,
			}
			err = k8sClient.Get(ctx, key, &createdInst)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(len(createdInst.Status.LastApplied)).Should(BeZero())

			var cr rbacv1.ClusterRole
			key = client.ObjectKey{
				Name: instance.InstanceResourceName(clusterLevelInstName, "privileged"),
			}
			err = k8sClient.Get(ctx, key, &cr)
			Expect(apierrs.IsNotFound(err)).Should(BeTrue())
		})
	})
})

func ownerRef(obj runtime.Object, scheme *runtime.Scheme) metav1.OwnerReference {
	type ownerObject interface {
		GetName() string
		GetUID() types.UID
	}

	owner, ok := obj.(ownerObject)
	Expect(ok).Should(BeTrue())

	gvk, err := apiutil.GVKForObject(obj, scheme)
	Expect(err).ShouldNot(HaveOccurred())
	return metav1.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		Name:               owner.GetName(),
		UID:                owner.GetUID(),
		BlockOwnerDeletion: pointer.BoolPtr(true),
		Controller:         pointer.BoolPtr(true),
	}
}

func Test_unstToObjectRef(t *testing.T) {
	creationTimestamp := "2021-07-13T01:50:08Z"
	creationTime, err := time.Parse("2006-01-02T03:04:05Z", creationTimestamp)
	if err != nil {
		t.Fatal(err)
	}
	creationTime = creationTime.Local()
	metaCreationTime := metav1.NewTime(creationTime)

	now := metav1.Now()

	type args struct {
		obj             unstructured.Unstructured
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
				obj: unstructured.Unstructured{
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
				updateTimestamp: now,
			},
			want: cosmov1alpha1.ObjectRef{
				ObjectReference: corev1.ObjectReference{
					APIVersion: "networking.k8s.io/v1",
					Kind:       "Ingress",
					Name:       "test",
					Namespace:  "default",
				},
				CreationTimestamp: &metaCreationTime,
				UpdateTimestamp:   &now,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unstToObjectRef(tt.args.obj, tt.args.updateTimestamp); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unstToObjectRef() = %v, want %v", got, tt.want)
			}
		})
	}
}
