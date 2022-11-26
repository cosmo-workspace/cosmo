package kubeutil

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var k8sClient client.Client
var testEnv *envtest.Environment

func TestKubeutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Kubeutil Suite")
}

func init() {
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme.Scheme))
	//+kubebuilder:scaffold:scheme
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	createInitObjects(context.Background())

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func createInitObjects(ctx context.Context) {
	tmpl1 := &cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tmpl1",
			Labels: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: "test",
			},
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  labels:
	cosmo-workspace.github.io/instance: '{{INSTANCE}}'
	cosmo-workspace.github.io/template: nginx
  name: nginx
spec:
  rules:
  - host: '{{INSTANCE}}-{{NAMESPACE}}.{{DOMAIN}}'
	http:
	  paths:
	  - path:
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
	cosmo-workspace.github.io/instance: '{{INSTANCE}}'
	cosmo-workspace.github.io/template: nginx
  name: nginx
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
  labels:
	cosmo-workspace.github.io/instance: '{{INSTANCE}}'
	cosmo-workspace.github.io/template: nginx
  name: nginx
  namespace: '{{NAMESPACE}}'
spec:
  replicas: 1
  selector:
	matchLabels:
	  cosmo-workspace.github.io/instance: '{{INSTANCE}}'
	  cosmo-workspace.github.io/template: nginx
  template:
	metadata:
	  labels:
		cosmo-workspace.github.io/instance: '{{INSTANCE}}'
		cosmo-workspace.github.io/template: nginx
	spec:
	  containers:
	  - image: 'nginx:{{IMAGE_TAG}}'
		name: nginx
		ports:
		- containerPort: 80
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

	inst1 := &cosmov1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "inst1",
			Namespace: "default",
		},
		Spec: cosmov1alpha1.InstanceSpec{
			Template: cosmov1alpha1.TemplateRef{
				Name: "tmpl1",
			},
			Override: cosmov1alpha1.OverrideSpec{},
			Vars: map[string]string{
				"{{DOMAIN}}":    "example.com",
				"{{IMAGE_TAG}}": "latest",
			},
		},
	}
	inst1.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   cosmov1alpha1.GroupVersion.Group,
		Version: cosmov1alpha1.GroupVersion.Version,
		Kind:    "Instance",
	})

	tmpl2 := &cosmov1alpha1.Template{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tmpl2",
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `apiVersion: v1
kind: Pod
metadata:
  name: alpine
spec:
  containers:
  - image: 'alpine:latest'
    name: alpine
    command:
    - echo
    - helloworld
`,
		},
	}

	inst2 := &cosmov1alpha1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "inst2",
			Namespace: "default",
		},
		Spec: cosmov1alpha1.InstanceSpec{
			Template: cosmov1alpha1.TemplateRef{
				Name: "tmpl2",
			},
		},
	}

	inst2Pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.InstanceResourceName(inst2.Name, "alpine"),
			Namespace: inst2.Namespace,
			Labels: map[string]string{
				cosmov1alpha1.LabelKeyInstanceName: "inst2",
				cosmov1alpha1.LabelKeyTemplateName: "tmpl2",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image:   "alpine:latest",
					Name:    "alpine",
					Command: []string{"echo", "helloworld"},
				},
			},
		},
	}

	Expect(k8sClient.Create(ctx, tmpl1)).ShouldNot(HaveOccurred())
	Expect(k8sClient.Create(ctx, inst1)).ShouldNot(HaveOccurred())
	Expect(k8sClient.Create(ctx, tmpl2)).ShouldNot(HaveOccurred())
	Expect(k8sClient.Create(ctx, inst2)).ShouldNot(HaveOccurred())
	Expect(k8sClient.Create(ctx, inst2Pod)).ShouldNot(HaveOccurred())
}
