package kosmo

import (
	"context"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var k8sFakeClient client.Client
var tmpl1 *cosmov1alpha1.Template
var inst1 *cosmov1alpha1.Instance
var tmpl2 *cosmov1alpha1.Template
var inst2 *cosmov1alpha1.Instance
var inst2Pod *corev1.Pod

func init() {
	scheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(scheme)
	cosmov1alpha1.AddToScheme(scheme)

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
	cosmo/instance: '{{INSTANCE}}'
	cosmo/template: nginx
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
			Name:      cosmov1alpha1.InstanceResourceName(inst2.Name, "alpine"),
			Namespace: inst2.Namespace,
			Labels: map[string]string{
				cosmov1alpha1.LabelKeyInstance: "inst2",
				cosmov1alpha1.LabelKeyTemplate: "tmpl2",
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

	etmpl1 = tmpl1.DeepCopy()
	einst1 = inst1.DeepCopy()
	etmpl2 = tmpl2.DeepCopy()
	einst2 = inst2.DeepCopy()
	einst2Pod = inst2Pod.DeepCopy()

	k8sFakeClient = fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(tmpl1, inst1, tmpl2, inst2, inst2Pod).
		Build()
}

func TestClient_GetInstance(t *testing.T) {
	type fields struct {
		Client client.Client
	}
	type args struct {
		ctx       context.Context
		name      string
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *cosmov1alpha1.Instance
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				Client: k8sFakeClient,
			},
			args: args{
				ctx:       context.TODO(),
				name:      inst1.Name,
				namespace: "default",
			},
			want:    inst1,
			wantErr: false,
		},
		{
			name: "not found",
			fields: fields{
				Client: k8sFakeClient,
			},
			args: args{
				ctx:       context.TODO(),
				name:      "not found",
				namespace: "default",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Client: tt.fields.Client,
			}
			got, err := c.GetInstance(tt.args.ctx, tt.args.name, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetInstance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetInstance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_ListInstances(t *testing.T) {
	type fields struct {
		Client client.Client
	}
	type args struct {
		ctx       context.Context
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []cosmov1alpha1.Instance
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				Client: k8sFakeClient,
			},
			args: args{
				ctx:       context.TODO(),
				namespace: "default",
			},
			want:    []cosmov1alpha1.Instance{*inst1, *inst2},
			wantErr: false,
		},
		{
			name: "not found",
			fields: fields{
				Client: k8sFakeClient,
			},
			args: args{
				ctx:       context.TODO(),
				namespace: "kube-system",
			},
			want:    []cosmov1alpha1.Instance{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Client: tt.fields.Client,
			}
			got, err := c.ListInstances(tt.args.ctx, tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListInstances() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListInstances() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_ListTemplates(t *testing.T) {
	type fields struct {
		Client client.Client
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []cosmov1alpha1.Template
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				Client: k8sFakeClient,
			},
			args: args{
				ctx: context.TODO(),
			},
			want:    []cosmov1alpha1.Template{*tmpl1, *tmpl2},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Client: tt.fields.Client,
			}
			got, err := c.ListTemplates(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListTemplates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListTemplates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_ListTemplatesByType(t *testing.T) {
	type fields struct {
		Client client.Client
	}
	type args struct {
		ctx       context.Context
		tmplTypes []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []cosmov1alpha1.Template
		wantErr bool
	}{
		{
			name: "type test",
			fields: fields{
				Client: k8sFakeClient,
			},
			args: args{
				ctx:       context.TODO(),
				tmplTypes: []string{"test"},
			},
			want:    []cosmov1alpha1.Template{*tmpl1},
			wantErr: false,
		},
		{
			name: "not found",
			fields: fields{
				Client: k8sFakeClient,
			},
			args: args{
				ctx:       context.TODO(),
				tmplTypes: []string{"notfound"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Client: tt.fields.Client,
			}
			got, err := c.ListTemplatesByType(tt.args.ctx, tt.args.tmplTypes)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListTemplatesByType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListTemplatesByType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetTemplate(t *testing.T) {
	type fields struct {
		Client client.Client
	}
	type args struct {
		ctx      context.Context
		tmplName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *cosmov1alpha1.Template
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				Client: k8sFakeClient,
			},
			args: args{
				ctx:      context.TODO(),
				tmplName: tmpl1.Name,
			},
			want:    tmpl1,
			wantErr: false,
		},
		{
			name: "not found",
			fields: fields{
				Client: k8sFakeClient,
			},
			args: args{
				ctx:      context.TODO(),
				tmplName: "notfound",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Client: tt.fields.Client,
			}
			got, err := c.GetTemplate(tt.args.ctx, tt.args.tmplName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   cosmov1alpha1.GroupVersion.Group,
					Version: cosmov1alpha1.GroupVersion.Version,
					Kind:    "Template",
				})
			}
			if !tt.wantErr && !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}
