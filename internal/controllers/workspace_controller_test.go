package controllers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

var _ = Describe("Workspace controller", func() {
	const tmplName string = "code-server-test"
	const wsName string = "ws-test"
	const userName string = "wsctltest"
	var nsName string = wsv1alpha1.UserNamespace(userName)

	wsConfig := wsv1alpha1.Config{
		DeploymentName:      "ws-dep",
		ServiceName:         "ws-svc",
		IngressName:         "ws-ing",
		ServiceMainPortName: "mainPort",
		URLBase:             "https://{{NETRULE_PORT_GROUP}}-{{WOKRSPACE}}-{{USER}}.domain",
	}

	tmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: tmplName,
			Labels: map[string]string{
				cosmov1alpha1.LabelKeyTemplateType: wsv1alpha1.TemplateTypeWorkspace,
			},
			Annotations: map[string]string{
				wsv1alpha1.TemplateAnnKeyWorkspaceDeployment:      wsConfig.DeploymentName,
				wsv1alpha1.TemplateAnnKeyWorkspaceIngress:         wsConfig.IngressName,
				wsv1alpha1.TemplateAnnKeyWorkspaceService:         wsConfig.ServiceName,
				wsv1alpha1.TemplateAnnKeyWorkspaceServiceMainPort: wsConfig.ServiceMainPortName,
				wsv1alpha1.TemplateAnnKeyURLBase:                  wsConfig.URLBase,
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  labels:
    cosmo/instance: '{{INSTANCE}}'
    cosmo/template: code-server-test
  name: ws-ing
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
            name: '{{INSTANCE}}-ws-svc'
            port: 
              number: 8080
---
apiVersion: v1
kind: Service
metadata:
  labels:
    cosmo/instance: '{{INSTANCE}}'
    cosmo/template: code-server-test
  name: ws-svc
  namespace: '{{NAMESPACE}}'
spec:
  ports:
  - name: main
    port: 8080
    protocol: TCP
  selector:
    cosmo/instance: '{{INSTANCE}}'
    cosmo/template: code-server-test
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    cosmo/instance: '{{INSTANCE}}'
    cosmo/template: code-server-test
  name: ws-dep
  namespace: '{{NAMESPACE}}'
spec:
  replicas: 1
  selector:
    matchLabels:
      cosmo/instance: '{{INSTANCE}}'
      cosmo/template: code-server-test
  template:
    metadata:
      labels:
        cosmo/instance: '{{INSTANCE}}'
        cosmo/template: code-server-test
    spec:
      containers:
      - image: 'code-server:{{IMAGE_TAG}}'
        name: code-server-test
        ports:
        - containerPort: 8080
          name: main
          protocol: TCP
`,
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

	ws := wsv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wsName,
			Namespace: nsName,
		},
		Spec: wsv1alpha1.WorkspaceSpec{
			Template: cosmov1alpha1.TemplateRef{
				Name: tmplName,
			},
			Replicas: pointer.Int64(1),
			Vars: map[string]string{
				"{{DOMAIN}}":    "example.com",
				"{{IMAGE_TAG}}": "latest",
			},
		},
	}

	expectedInst := cosmov1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wsName,
			Namespace: nsName,
		},
		Spec: cosmov1alpha1.InstanceSpec{
			Template: cosmov1alpha1.TemplateRef{
				Name: tmplName,
			},
			Vars: map[string]string{
				"{{DOMAIN}}":                           "example.com",
				"{{IMAGE_TAG}}":                        "latest",
				"{{WORKSPACE_DEPLOYMENT_NAME}}":        wsConfig.DeploymentName,
				"{{WORKSPACE_INGRESS_NAME}}":           wsConfig.IngressName,
				"{{WORKSPACE_SERVICE_NAME}}":           wsConfig.ServiceName,
				"{{WORKSPACE_SERVICE_MAIN_PORT_NAME}}": wsConfig.ServiceMainPortName,
				"{{WORKSPACE}}":                        wsName,
				"{{USERID}}":                           userName,
			},
			Override: cosmov1alpha1.OverrideSpec{
				Scale: []cosmov1alpha1.ScalingOverrideSpec{
					{
						Target: cosmov1alpha1.ObjectRef{
							ObjectReference: corev1.ObjectReference{
								APIVersion: "apps/v1",
								Kind:       "Deployment",
								Name:       wsConfig.DeploymentName,
							},
						},
						Replicas: 1,
					},
				},
				Network: &cosmov1alpha1.NetworkOverrideSpec{
					Ingress: []cosmov1alpha1.IngressOverrideSpec{{TargetName: wsConfig.IngressName}},
					Service: []cosmov1alpha1.ServiceOverrideSpec{{TargetName: wsConfig.ServiceName}},
				},
			},
		},
	}

	Context("when creating Template resource on new cluster", func() {
		It("should do nothing", func() {
			ctx := context.Background()

			By("creating template before instance")

			ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}
			err := k8sClient.Create(ctx, &ns)
			Expect(err).ShouldNot(HaveOccurred())

			err = k8sClient.Create(ctx, &tmpl)
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

	Context("when creating a Workspace resource", func() {
		It("should do reconcile once and create Instance resources", func() {
			ctx := context.Background()

			err := k8sClient.Create(ctx, &ws)
			Expect(err).ShouldNot(HaveOccurred())

			By("checking if Instance resources is as expected")

			var createdInst cosmov1alpha1.Instance
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      wsName,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &createdInst)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			instRef := corev1.ObjectReference{
				APIVersion:      cosmov1alpha1.GroupVersion.String(),
				Kind:            "Instance",
				Name:            createdInst.Name,
				Namespace:       createdInst.Namespace,
				UID:             createdInst.UID,
				ResourceVersion: createdInst.ResourceVersion,
			}

			expected := expectedInst.DeepCopy()
			ownerRef := ownerRef(&ws, scheme.Scheme)
			expected.OwnerReferences = []metav1.OwnerReference{ownerRef}

			created := looseDeepCopyObject(createdInst)

			clog.PrintObjectDiff(os.Stderr, created, expected)
			eq := equality.Semantic.DeepEqual(created, expected)
			Expect(eq).Should(BeTrue())

			By("fetching workspace resource and checking workspace status")

			var createdWs wsv1alpha1.Workspace
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      wsName,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &createdWs)
				if err != nil {
					return err
				}
				if equality.Semantic.DeepEqual(createdWs.Status.Instance.ObjectReference, instRef) {
					return errors.New("workspace status is not updated")
				}
				return nil
			}, time.Second*10).Should(Succeed())
		})
	})

	Context("when updating Workspace spec", func() {
		It("should do reconcile again and update child Instance", func() {
			ctx := context.Background()

			// fetch current workspace
			var ws wsv1alpha1.Workspace
			Eventually(func() error {
				key := types.NamespacedName{
					Name:      wsName,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &ws)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			// update Workspace spec
			ws.Spec.Replicas = pointer.Int64(0)
			ws.Spec.Network = []wsv1alpha1.NetworkRule{
				{
					PortName:         "port1",
					PortNumber:       3000,
					HTTPPath:         "/path",
					TargetPortNumber: pointer.Int32(30000),
					Group:            pointer.String("group1"),
					Public:           false,
				},
			}

			err := k8sClient.Update(ctx, &ws)
			Expect(err).ShouldNot(HaveOccurred())

			var createdInst cosmov1alpha1.Instance
			Eventually(func() error {
				key := client.ObjectKey{
					Name:      wsName,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &createdInst)
				if err != nil {
					return err
				}
				if createdInst.Spec.Override.Scale[0].Replicas != 0 {
					return errors.New("replica is not zero")
				}
				return nil
			}, time.Second*10).Should(Succeed())

			expected := expectedInst.DeepCopy()
			ownerRef := ownerRef(&ws, scheme.Scheme)
			expected.OwnerReferences = []metav1.OwnerReference{ownerRef}
			prefix := netv1.PathTypePrefix
			expected.Spec.Override.Scale[0].Replicas = 0
			expected.Spec.Override.Network = &cosmov1alpha1.NetworkOverrideSpec{
				Ingress: []cosmov1alpha1.IngressOverrideSpec{
					{
						TargetName: wsConfig.IngressName,
						Rules: []netv1.IngressRule{
							{
								// Host is filled by Webhook
								// Host: fmt.Sprintf("%s-%s-%s.domain", "group1", wsName, userName),
								IngressRuleValue: netv1.IngressRuleValue{
									HTTP: &netv1.HTTPIngressRuleValue{
										Paths: []netv1.HTTPIngressPath{
											{
												Path:     "/path",
												PathType: &prefix,
												Backend: netv1.IngressBackend{
													Service: &netv1.IngressServiceBackend{
														Name: fmt.Sprintf("%s-ws-svc", wsName),
														Port: netv1.ServiceBackendPort{
															Name: "port1",
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
				Service: []cosmov1alpha1.ServiceOverrideSpec{
					{
						TargetName: wsConfig.ServiceName,
						Ports: []corev1.ServicePort{
							{
								Name:       "port1",
								Protocol:   corev1.ProtocolTCP,
								Port:       3000,
								TargetPort: intstr.FromInt(30000),
							},
						},
					},
				},
			}

			created := looseDeepCopyObject(createdInst)

			clog.PrintObjectDiff(os.Stderr, created, expected)
			eq := equality.Semantic.DeepEqual(created, expected)
			Expect(eq).Should(BeTrue())
		})
	})
})

func looseDeepCopyObject(inst cosmov1alpha1.Instance) *cosmov1alpha1.Instance {
	loose := inst.DeepCopy()
	loose.SetSelfLink("")
	loose.SetUID("")
	loose.SetResourceVersion("")
	loose.SetGeneration(0)
	loose.SetCreationTimestamp(metav1.Time{})
	loose.SetManagedFields(nil)
	loose.Status = cosmov1alpha1.InstanceStatus{}
	return loose
}
