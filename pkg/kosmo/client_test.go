package kosmo

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
)

var k8sFakeClient client.Client
var tmpl1 *cosmov1alpha1.Template
var inst1 *cosmov1alpha1.Instance
var tmpl2 *cosmov1alpha1.Template
var inst2 *cosmov1alpha1.Instance
var inst2Pod *corev1.Pod

var ctmpl1 *cosmov1alpha1.ClusterTemplate
var cinst1 *cosmov1alpha1.ClusterInstance

func init() {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme))

	tmpl1 = &cosmov1alpha1.Template{
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

	inst1 = &cosmov1alpha1.Instance{
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

	tmpl2 = &cosmov1alpha1.Template{
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

	inst2 = &cosmov1alpha1.Instance{
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

	inst2Pod = &corev1.Pod{
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

	cinst1 = &cosmov1alpha1.ClusterInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cinst1",
		},
		Spec: cosmov1alpha1.InstanceSpec{
			Template: cosmov1alpha1.TemplateRef{
				Name: "ctmpl1",
			},
			Override: cosmov1alpha1.OverrideSpec{},
			Vars: map[string]string{
				"{{DOMAIN}}":    "example.com",
				"{{IMAGE_TAG}}": "latest",
			},
		},
	}

	ctmpl1 = &cosmov1alpha1.ClusterTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ctmpl1",
		},
		Spec: cosmov1alpha1.TemplateSpec{
			RawYaml: `
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: namespace-reader
rules:
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - list
  - watch
`,
		},
	}

	etmpl1 = tmpl1.DeepCopy()
	einst1 = inst1.DeepCopy()
	etmpl2 = tmpl2.DeepCopy()
	einst2 = inst2.DeepCopy()
	einst2Pod = inst2Pod.DeepCopy()

	ecinst1 = cinst1.DeepCopy()
	ectmpl1 = ctmpl1.DeepCopy()

	k8sFakeClient = fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(tmpl1, inst1, tmpl2, inst2, inst2Pod, cinst1, ctmpl1).
		Build()
}

func TestNewClientByRestConfig(t *testing.T) {

	type args struct {
		cfg    *rest.Config
		scheme *runtime.Scheme
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Invalid cfg",
			args: args{
				cfg:    cfg,
				scheme: scheme.Scheme,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClientByRestConfig(tt.args.cfg, tt.args.scheme)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClientByRestConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
