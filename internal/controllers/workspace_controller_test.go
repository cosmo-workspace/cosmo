package controllers

import (
	"context"
	"time"

	. "github.com/cosmo-workspace/cosmo/pkg/kubeutil/test/gomega"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

var _ = Describe("Workspace controller", func() {
	const tmplName string = "code-server-test"
	const wsName string = "ws-test"
	const userName string = "wsctltest"
	var nsName string = cosmov1alpha1.UserNamespace(userName)

	wsConfig := cosmov1alpha1.Config{
		DeploymentName:      "ws-dep",
		ServiceName:         "ws-svc",
		IngressName:         "ws-ing",
		ServiceMainPortName: "mainPort",
		URLBase:             "https://{{NETRULE_GROUP}}-{{WOKRSPACE}}-{{USER}}.domain",
	}

	tmpl := cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: tmplName,
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeWorkspace,
			},
			Annotations: map[string]string{
				cosmov1alpha1.WorkspaceTemplateAnnKeyDeploymentName:  wsConfig.DeploymentName,
				cosmov1alpha1.WorkspaceTemplateAnnKeyIngressName:     wsConfig.IngressName,
				cosmov1alpha1.WorkspaceTemplateAnnKeyServiceName:     wsConfig.ServiceName,
				cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort: wsConfig.ServiceMainPortName,
				cosmov1alpha1.WorkspaceTemplateAnnKeyURLBase:         wsConfig.URLBase,
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
					Var: "{{IMAGE_TAG}}",
				},
			},
		},
	}

	ws := cosmov1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      wsName,
			Namespace: nsName,
		},
		Spec: cosmov1alpha1.WorkspaceSpec{
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
				return k8sClient.Get(ctx, client.ObjectKey{Name: tmplName}, &createdTmpl)
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
				return k8sClient.Get(ctx, key, &createdInst)
			}, time.Second*10).Should(Succeed())

			instRef := corev1.ObjectReference{
				APIVersion:      cosmov1alpha1.GroupVersion.String(),
				Kind:            "Instance",
				Name:            createdInst.Name,
				Namespace:       createdInst.Namespace,
				UID:             createdInst.UID,
				ResourceVersion: createdInst.ResourceVersion,
			}

			created := looseDeepCopyObject(createdInst)
			Expect(created).To(MatchSnapShot())

			By("fetching workspace resource and checking workspace status")

			var createdWs cosmov1alpha1.Workspace
			Eventually(func() corev1.ObjectReference {
				key := client.ObjectKey{
					Name:      wsName,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &createdWs)
				Expect(err).ShouldNot(HaveOccurred())

				return createdWs.Status.Instance.ObjectReference
			}, time.Second*10).Should(BeEqualityDeepEqual(instRef))
		})
	})

	Context("when updating Workspace spec", func() {
		It("should do reconcile again and update child Instance", func() {
			ctx := context.Background()

			// fetch current workspace
			var ws cosmov1alpha1.Workspace
			Eventually(func() error {
				key := types.NamespacedName{
					Name:      wsName,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &ws)
				if err != nil {
					return err
				}

				// update Workspace spec
				ws.Spec.Replicas = pointer.Int64(0)
				ws.Spec.Network = []cosmov1alpha1.NetworkRule{
					{
						Name:             "port1",
						PortNumber:       3000,
						HTTPPath:         "/path",
						TargetPortNumber: pointer.Int32(30000),
						Group:            pointer.String("group1"),
						Public:           false,
					},
				}
				return k8sClient.Update(ctx, &ws)

			}, time.Second*60).Should(Succeed())

			var createdInst cosmov1alpha1.Instance
			Eventually(func() int64 {
				key := client.ObjectKey{
					Name:      wsName,
					Namespace: nsName,
				}
				err := k8sClient.Get(ctx, key, &createdInst)
				Expect(err).ShouldNot(HaveOccurred())

				return createdInst.Spec.Override.Scale[0].Replicas
			}, time.Second*10).Should(BeEquivalentTo(0))

			created := looseDeepCopyObject(createdInst)
			Expect(created).To(MatchSnapShot())
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
	for i := range loose.OwnerReferences {
		loose.OwnerReferences[i].UID = ""
	}
	return loose
}
