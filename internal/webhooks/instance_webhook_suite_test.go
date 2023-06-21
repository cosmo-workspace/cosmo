package webhooks

import (
	"context"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

var _ = Describe("Instance webhook", func() {
	wsConfig := cosmov1alpha1.Config{
		DeploymentName:      "ws-dep",
		ServiceName:         "ws-svc",
		ServiceMainPortName: "mainPort",
	}

	cstmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "code-server-test",
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeWorkspace,
			},
			Annotations: map[string]string{
				cosmov1alpha1.WorkspaceTemplateAnnKeyDeploymentName:  wsConfig.DeploymentName,
				cosmov1alpha1.WorkspaceTemplateAnnKeyServiceName:     wsConfig.ServiceName,
				cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort: wsConfig.ServiceMainPortName,
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  labels:
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-test
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
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-test
  name: ws-svc
  namespace: '{{NAMESPACE}}'
spec:
  ports:
  - name: main
    port: 8080
    protocol: TCP
  selector:
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-test
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: code-server-test
  name: ws-dep
  namespace: '{{NAMESPACE}}'
spec:
  replicas: 1
  selector:
    matchLabels:
      cosmo-workspace.github.io/instance: '{{INSTANCE}}'
      cosmo-workspace.github.io/template: code-server-test
  template:
    metadata:
      labels:
        cosmo-workspace.github.io/instance: '{{INSTANCE}}'
        cosmo-workspace.github.io/template: code-server-test
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

	Context("when creating instance with existing template and vars", func() {
		It("should pass with defaulting instance name prefix", func() {
			ctx := context.Background()

			var err error
			err = k8sClient.Create(ctx, &cstmpl)
			Expect(err).ShouldNot(HaveOccurred())

			var createdTmpl cosmov1alpha1.Template
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: cstmpl.Name}, &createdTmpl)
			}, time.Second*10).Should(Succeed())

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
						PatchesJson6902: []cosmov1alpha1.Json6902{
							{
								Target: cosmov1alpha1.ObjectRef{
									ObjectReference: corev1.ObjectReference{
										APIVersion: metav1.GroupVersion{Group: "", Version: "v1"}.String(),
										Kind:       "Service",
										Name:       "ws-svc",
									},
								},
								Patch: `[{"op": "replace","path": "/spec/type","value": "LoadBalancer"}]`,
							},
						},
					},
				},
			}

			err = k8sClient.Create(ctx, &inst)
			Expect(err).ShouldNot(HaveOccurred())

			var createdInst cosmov1alpha1.Instance
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: inst.GetName(), Namespace: inst.GetNamespace()}, &createdInst)
			}, time.Second*10).Should(Succeed())

			Expect(InstanceSnapshot(&createdInst)).Should(MatchSnapShot())
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

			var createdInst cosmov1alpha1.Instance
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: inst.GetName(), Namespace: inst.GetNamespace()}, &createdInst)
			}, time.Second*10).Should(Succeed())
			Expect(InstanceSnapshot(&createdInst)).Should(MatchSnapShot())
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
								Patch: `[{"op": "replace","path": "/spec/type","value": "LoadBalancer"}]`,
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

			var createdTmpl cosmov1alpha1.Template
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: tmpl.Name}, &createdTmpl)
			}, time.Second*10).Should(Succeed())

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

			var createdTmpl cosmov1alpha1.ClusterTemplate
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: tmpl.Name}, &createdTmpl)
			}, time.Second*10).Should(Succeed())

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

			var createdTmpl cosmov1alpha1.ClusterTemplate
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: t.Name}, &createdTmpl)
			}, time.Second*10).Should(Succeed())

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
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: '{{TEMPLATE}}'
  name: '{{INSTANCE}}-role'
  namespace: '{{NAMESPACE}}'
rules:
- apiGroups:
  - cosmo-workspace.github.io
  resources:
  - workspaces
  verbs:
  - patch
  - update
  - get
  - list
  - watch
- apiGroups:
  - cosmo-workspace.github.io
  resources:
  - workspaces/status
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - cosmo-workspace.github.io
  resources:
  - instances
  verbs:
  - patch
  - update
  - get
  - list
  - watch
- apiGroups:
  - cosmo-workspace.github.io
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
    cosmo-workspace.github.io/instance: '{{INSTANCE}}'
    cosmo-workspace.github.io/template: '{{TEMPLATE}}'
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

			var createdTmpl cosmov1alpha1.Template
			Eventually(func() error {
				return k8sClient.Get(ctx, client.ObjectKey{Name: t.Name}, &createdTmpl)
			}, time.Second*10).Should(Succeed())

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
