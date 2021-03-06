package webhooks

import (
	"context"
	"testing"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/kubeutil/test/gomega"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
)

var _ = Describe("Instance webhook", func() {
	wsConfig := wsv1alpha1.Config{
		DeploymentName:      "ws-dep",
		ServiceName:         "ws-svc",
		IngressName:         "ws-ing",
		ServiceMainPortName: "mainPort",
		URLBase:             "https://{{NETRULE_GROUP}}-{{INSTANCE}}-{{NAMESPACE}}.example.com",
	}

	cstmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "code-server-test",
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: wsv1alpha1.TemplateTypeWorkspace,
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
  - host: 'main-{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}'
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: ws-svc
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
					Var:     "{{IMAGE_TAG}}",
					Default: "latest",
				},
			},
		},
	}
	prefix := netv1.PathTypePrefix

	Context("when creating instance with existing template and vars", func() {
		It("should pass with defaulting instance name prefix", func() {
			ctx := context.Background()

			var err error
			err = k8sClient.Create(ctx, &cstmpl)
			Expect(err).ShouldNot(HaveOccurred())

			inst := cosmov1alpha1.Instance{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Instance",
					APIVersion: cosmov1alpha1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testinst1",
					Namespace: "default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: cstmpl.GetName()},
					Vars: map[string]string{
						"DOMAIN":    "example.com",
						"IMAGE_TAG": "latest",
					},
					Override: cosmov1alpha1.OverrideSpec{
						Scale: []cosmov1alpha1.ScalingOverrideSpec{
							{
								Target: cosmov1alpha1.ObjectRef{
									ObjectReference: corev1.ObjectReference{
										APIVersion: metav1.GroupVersion{Group: "apps", Version: "v1"}.String(),
										Kind:       "Deployment",
										Name:       "ws-dep",
									},
								},
								Replicas: 3,
							},
						},
						Network: &cosmov1alpha1.NetworkOverrideSpec{
							Service: []cosmov1alpha1.ServiceOverrideSpec{
								{
									TargetName: "ws-svc",
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
									TargetName: "ws-ing",
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
																	Name: "ws-svc",
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
										APIVersion: metav1.GroupVersion{Group: "", Version: "v1"}.String(),
										Kind:       "Service",
										Name:       "ws-svc",
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
					},
				},
			}

			err = k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())

			expectedInst := cosmov1alpha1.Instance{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Instance",
					APIVersion: cosmov1alpha1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testinst1",
					Namespace: "default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: cstmpl.GetName()},
					Vars: map[string]string{
						"DOMAIN":    "example.com",
						"IMAGE_TAG": "latest",
					},
					Override: cosmov1alpha1.OverrideSpec{
						Scale: []cosmov1alpha1.ScalingOverrideSpec{
							{
								Target: cosmov1alpha1.ObjectRef{
									ObjectReference: corev1.ObjectReference{
										APIVersion: metav1.GroupVersion{Group: "apps", Version: "v1"}.String(),
										Kind:       "Deployment",
										Name:       "testinst1-ws-dep",
									},
								},
								Replicas: 3,
							},
						},
						Network: &cosmov1alpha1.NetworkOverrideSpec{
							Service: []cosmov1alpha1.ServiceOverrideSpec{
								{
									TargetName: "testinst1-ws-svc",
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
									TargetName: "testinst1-ws-ing",
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
																	Name: "testinst1-ws-svc",
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
										APIVersion: metav1.GroupVersion{Group: "", Version: "v1"}.String(),
										Kind:       "Service",
										Name:       "testinst1-ws-svc",
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
					},
				},
			}

			var createdInst cosmov1alpha1.Instance
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: inst.GetName(), Namespace: inst.GetNamespace()}, &createdInst)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			expectedInst.ObjectMeta = createdInst.ObjectMeta

			Expect(&createdInst).Should(BeLooseDeepEqual(&expectedInst))
		})
	})

	Context("when creating instance with existing template and no default vars", func() {
		It("should pass with defaulting vars", func() {
			ctx := context.Background()

			inst := cosmov1alpha1.Instance{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Instance",
					APIVersion: cosmov1alpha1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testinst2",
					Namespace: "default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: cstmpl.GetName()},
					Vars: map[string]string{
						"DOMAIN": "example.com",
					},
				},
			}

			err := k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())

			expectedInst := cosmov1alpha1.Instance{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Instance",
					APIVersion: cosmov1alpha1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testinst2",
					Namespace: "default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: cstmpl.GetName()},
					Vars: map[string]string{
						"DOMAIN":        "example.com",
						"{{IMAGE_TAG}}": "latest",
					},
				},
			}

			var createdInst cosmov1alpha1.Instance
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: inst.GetName(), Namespace: inst.GetNamespace()}, &createdInst)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			expectedInst.ObjectMeta = createdInst.ObjectMeta
			Expect(&createdInst).Should(BeLooseDeepEqual(&expectedInst))
		})
	})

	Context("when creating instance with non-existing template", func() {
		It("should deny", func() {
			ctx := context.Background()

			inst := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testinst3",
					Namespace: "default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: "notfound",
					},
				},
			}

			err := k8sClient.Create(ctx, &inst)
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("when creating instance without required vars", func() {
		It("should deny", func() {
			ctx := context.Background()

			inst := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testinst4",
					Namespace: "default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: cstmpl.GetName(),
					},
				},
			}

			err := k8sClient.Create(ctx, &inst)
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("when creating instance with invalid apiVersion in scaling target", func() {
		It("should deny", func() {
			ctx := context.Background()

			inst := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testinst5",
					Namespace: "default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: cstmpl.GetName()},
					Vars: map[string]string{
						"DOMAIN": "example.com",
					},
					Override: cosmov1alpha1.OverrideSpec{
						Scale: []cosmov1alpha1.ScalingOverrideSpec{
							{
								Target: cosmov1alpha1.ObjectRef{
									ObjectReference: corev1.ObjectReference{
										APIVersion: "apps/v1/v1",
										Kind:       "Deployment",
										Name:       "testinst1-ws-dep",
									},
								},
								Replicas: 3,
							},
						},
					},
				},
			}

			err := k8sClient.Create(ctx, &inst)
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("when creating instance with invalid apiVersion in JSON patch target", func() {
		It("should deny", func() {
			ctx := context.Background()

			inst := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testinst6",
					Namespace: "default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: cstmpl.GetName()},
					Vars: map[string]string{
						"DOMAIN": "example.com",
					},
					Override: cosmov1alpha1.OverrideSpec{
						PatchesJson6902: []cosmov1alpha1.Json6902{
							{
								Target: cosmov1alpha1.ObjectRef{
									ObjectReference: corev1.ObjectReference{
										APIVersion: "v1/v1/v1",
										Kind:       "Service",
										Name:       "ws-svc",
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
					},
				},
			}

			err := k8sClient.Create(ctx, &inst)
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("when creatinng instance but dryrun failed", func() {
		It("should deny", func() {
			ctx := context.Background()

			inst := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testinst7-invalid-var",
					Namespace: "default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: cstmpl.GetName(),
					},
					Vars: map[string]string{
						"DOMAIN": "{{DOMAIN}}",
					},
				},
			}

			err := k8sClient.Create(ctx, &inst)
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("when including ClusterRole in Template", func() {
		It("should deny with invalid scope", func() {
			ctx := context.Background()

			tmpl := cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testtmpl7",
				},
				Spec: cosmov1alpha1.TemplateSpec{
					RawYaml: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: privileged
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
			err := k8sClient.Create(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			inst := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testinst8",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: tmpl.GetName()},
				},
			}

			err = k8sClient.Create(ctx, &inst)
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("when including ClusterRole in ClusterTemplate", func() {
		It("should pass", func() {
			ctx := context.Background()

			tmpl := cosmov1alpha1.ClusterTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testctmpl1",
				},
				Spec: cosmov1alpha1.TemplateSpec{
					RawYaml: `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: privileged
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
			err := k8sClient.Create(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			inst := cosmov1alpha1.ClusterInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testcinst1",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: tmpl.GetName()},
				},
			}

			err = k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("when including Pod in ClusterTemplate and ClusterInstance has namespace var", func() {
		It("should pass", func() {
			ctx := context.Background()

			t := cosmov1alpha1.ClusterTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testctmpl2",
				},
				Spec: cosmov1alpha1.TemplateSpec{
					RawYaml: `apiVersion: v1
kind: Pod
metadata:
  name: nginx
  namespace: "{{NAMESPACE}}"
spec:
  containers:
  - name: nginx
    image: {{IMAGE}}:alpine
`,
				},
			}
			err := k8sClient.Create(ctx, &t)
			Expect(err).ShouldNot(HaveOccurred())

			time.Sleep(3 * time.Second)

			inst := cosmov1alpha1.ClusterInstance{
				ObjectMeta: metav1.ObjectMeta{
					Name: "testcinst2",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{Name: t.GetName()},
					Vars: map[string]string{
						"NAMESPACE": "kube-system",
						"IMAGE":     "var",
					},
				},
			}

			err = k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("when creating role and its rolebinding at the same time", func() {
		It("should pass even though role is not found", func() {
			ctx := context.Background()

			name := "role-not-found-rolebinding"
			t := cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
				},
				Spec: cosmov1alpha1.TemplateSpec{
					RequiredVars: []cosmov1alpha1.RequiredVarSpec{
						{
							Var:     "SERVICE_ACCOUNT",
							Default: "default",
						},
					},
					RawYaml: `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  labels:
    cosmo/instance: '{{INSTANCE}}'
    cosmo/template: '{{TEMPLATE}}'
  name: '{{INSTANCE}}-role'
  namespace: '{{NAMESPACE}}'
rules:
- apiGroups:
  - workspace.cosmo-workspace.github.io
  resources:
  - workspaces
  verbs:
  - patch
  - update
  - get
  - list
  - watch
- apiGroups:
  - workspace.cosmo-workspace.github.io
  resources:
  - workspaces/status
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - cosmo.cosmo-workspace.github.io
  resources:
  - instances
  verbs:
  - patch
  - update
  - get
  - list
  - watch
- apiGroups:
  - cosmo.cosmo-workspace.github.io
  resources:
  - instances/status
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - services
  - secrets
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - get
  - list
  - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  labels:
    cosmo/instance: '{{INSTANCE}}'
    cosmo/template: '{{TEMPLATE}}'
  name: '{{INSTANCE}}-rolebinding'
  namespace: '{{NAMESPACE}}'
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: '{{INSTANCE}}-role'
subjects:
- kind: ServiceAccount
  name: '{{SERVICE_ACCOUNT}}'
  namespace: '{{NAMESPACE}}'
`,
				},
			}
			err := k8sClient.Create(ctx, &t)
			Expect(err).ShouldNot(HaveOccurred())

			i := cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: "default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: name,
					},
				},
			}
			err = k8sClient.Create(ctx, &i)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})

func Test_fixServiceNameInIngressBackend(t *testing.T) {
	type args struct {
		ingRules []netv1.IngressRule
		instName string
	}
	tests := []struct {
		name string
		args args
		want []netv1.IngressRule
	}{
		{
			name: "OK",
			args: args{
				ingRules: []netv1.IngressRule{
					{
						Host: "example.com",
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{
									{
										Backend: netv1.IngressBackend{
											Service: &netv1.IngressServiceBackend{
												Name: "test-svc",
											},
										},
									},
								},
							},
						},
					},
					{
						Host: "example.com",
						IngressRuleValue: netv1.IngressRuleValue{
							HTTP: &netv1.HTTPIngressRuleValue{
								Paths: []netv1.HTTPIngressPath{
									{
										Backend: netv1.IngressBackend{
											Service: &netv1.IngressServiceBackend{
												Name: "test-svc2",
											},
										},
									},
								},
							},
						},
					},
				},
				instName: "instance",
			},
			want: []netv1.IngressRule{
				{
					Host: "example.com",
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: "instance-test-svc",
										},
									},
								},
							},
						},
					},
				},
				{
					Host: "example.com",
					IngressRuleValue: netv1.IngressRuleValue{
						HTTP: &netv1.HTTPIngressRuleValue{
							Paths: []netv1.HTTPIngressPath{
								{
									Backend: netv1.IngressBackend{
										Service: &netv1.IngressServiceBackend{
											Name: "instance-test-svc2",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixIngressBackendName(tt.args.ingRules, tt.args.instName)
			if !equality.Semantic.DeepEqual(tt.args.ingRules, tt.want) {
				t.Error(tt.args, tt.want)
			}
		})
	}
}
