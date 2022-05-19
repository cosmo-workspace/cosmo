package webhooks

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
)

var _ = Describe("Template webhook", func() {
	Context("when creating template without urlbase", func() {
		It("should pass with defaulting urlbase", func() {
			ctx := context.Background()

			wsConfig := wsv1alpha1.Config{
				DeploymentName:      "ws-dep",
				ServiceName:         "ws-svc",
				IngressName:         "ws-ing",
				ServiceMainPortName: "mainPort",
				URLBase:             "https://{{NETRULE_GROUP}}-{{INSTANCE}}-{{NAMESPACE}}.example.com",
			}

			tmpl := cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "code-server-test-wh",
					Labels: map[string]string{
						cosmov1alpha1.TemplateLabelKeyType: wsv1alpha1.TemplateTypeWorkspace,
					},
					Annotations: map[string]string{
						wsv1alpha1.TemplateAnnKeyWorkspaceDeployment:      wsConfig.DeploymentName,
						wsv1alpha1.TemplateAnnKeyWorkspaceIngress:         wsConfig.IngressName,
						wsv1alpha1.TemplateAnnKeyWorkspaceService:         wsConfig.ServiceName,
						wsv1alpha1.TemplateAnnKeyWorkspaceServiceMainPort: wsConfig.ServiceMainPortName,
						// no urlbase
						// wsv1alpha1.TemplateAnnKeyURLBase:                  wsConfig.URLBase,
					},
				},
				Spec: cosmov1alpha1.TemplateSpec{
					RawYaml: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  labels:
	cosmo/template: '{{INSTANCE}}'
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
	cosmo/template: '{{INSTANCE}}'
	cosmo/template: code-server-test
  name: ws-svc
  namespace: '{{NAMESPACE}}'
spec:
  ports:
  - name: main
	port: 8080
	protocol: TCP
  selector:
	cosmo/template: '{{INSTANCE}}'
	cosmo/template: code-server-test
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
	cosmo/template: '{{INSTANCE}}'
	cosmo/template: code-server-test
  name: ws-dep
  namespace: '{{NAMESPACE}}'
spec:
  replicas: 1
  selector:
	matchLabels:
	  cosmo/template: '{{INSTANCE}}'
	  cosmo/template: code-server-test
  template:
	metadata:
	  labels:
		cosmo/template: '{{INSTANCE}}'
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

			err := k8sClient.Create(ctx, &tmpl)
			Expect(err).ShouldNot(HaveOccurred())

			expectedTmpl := tmpl.DeepCopy()
			expectedTmpl.TypeMeta = metav1.TypeMeta{
				Kind:       "Template",
				APIVersion: cosmov1alpha1.GroupVersion.String(),
			}
			expectedTmpl.Annotations[wsv1alpha1.TemplateAnnKeyURLBase] = DefaultURLBase

			var createdTmpl cosmov1alpha1.Template
			Eventually(func() error {
				err := k8sClient.Get(ctx, client.ObjectKey{Name: tmpl.GetName(), Namespace: tmpl.GetNamespace()}, &createdTmpl)
				if err != nil {
					return err
				}
				return nil
			}, time.Second*10).Should(Succeed())

			eq := kubeutil.LooseDeepEqual(&createdTmpl, expectedTmpl, kubeutil.WithPrintDiff(os.Stderr))
			Expect(eq).Should(BeTrue())
		})
	})
})
