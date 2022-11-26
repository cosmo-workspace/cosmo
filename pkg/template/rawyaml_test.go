package template

import (
	"reflect"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewRawYAMLBuilder(t *testing.T) {
	type args struct {
		data string
		inst *cosmov1alpha1.Instance
	}
	tests := []struct {
		name string
		args args
		want *RawYAMLBuilder
	}{
		{
			name: "OK",
			args: args{
				data: `apiVersion: networking.k8s.io/v1
kind: XXXX
metadata:
  annotations:
    kubernetes.io/ingress.class: alb
    cosmo/ingress-patch-enable: "true"
  labels:
    key: val
  name: test
  namespace: default
spec:
  host: example.com
---
apiVersion: v1
kind: Pod
metadata:
  name: test
  namespace: default
spec:
  hello: world`,
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"{{TEST}}": "OK"},
					},
				},
			},
			want: &RawYAMLBuilder{
				rawYaml: `apiVersion: networking.k8s.io/v1
kind: XXXX
metadata:
  annotations:
    kubernetes.io/ingress.class: alb
    cosmo/ingress-patch-enable: "true"
  labels:
    key: val
  name: test
  namespace: default
spec:
  host: example.com
---
apiVersion: v1
kind: Pod
metadata:
  name: test
  namespace: default
spec:
  hello: world`,
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"{{TEST}}": "OK"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRawYAMLBuilder(tt.args.data, tt.args.inst); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRawYAMLBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawYAMLBuilder_Build(t *testing.T) {
	type fields struct {
		data string
		inst *cosmov1alpha1.Instance
	}
	tests := []struct {
		name    string
		fields  fields
		want    []unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				data: `apiVersion: networking.k8s.io/v1
kind: XXXX
metadata:
  annotations:
    kubernetes.io/ingress.class: alb
    cosmo/ingress-patch-enable: "true"
  labels:
    key: val
  name: test
  namespace: default
spec:
  host: example.com
---
apiVersion: v1
kind: Pod
metadata:
  name: test
  namespace: default
spec:
  hello: world`,
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"{{TEST}}": "OK"},
					},
				},
			},
			want: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "networking.k8s.io/v1",
						"kind":       "XXXX",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "default",
							"labels": map[string]interface{}{
								"key": "val",
							},
							"annotations": map[string]interface{}{
								"kubernetes.io/ingress.class": "alb",
								"cosmo/ingress-patch-enable":  "true",
							},
						},
						"spec": map[string]interface{}{
							"host": "example.com",
						},
					},
				},
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "default",
						},
						"spec": map[string]interface{}{
							"hello": "world",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "separator line",
			fields: fields{
				data: `---
apiVersion: networking.k8s.io/v1
kind: XXXX
metadata:
  annotations:
    kubernetes.io/ingress.class: alb
    cosmo/ingress-patch-enable: "true"
  labels:
    key: val
  name: test
  namespace: default
spec:
  host: example.com
---
apiVersion: v1
kind: Pod
metadata:
  name: ---
  namespace: default
spec:
  hello: world
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: files
data:
  file1: "---\naaa\n---\nbbb\n"
  file2: |
    ---
    ccc
    ---
    ddd
---
`,
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"{{TEST}}": "OK"},
					},
				},
			},
			want: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "networking.k8s.io/v1",
						"kind":       "XXXX",
						"metadata": map[string]interface{}{
							"name":      "test",
							"namespace": "default",
							"labels": map[string]interface{}{
								"key": "val",
							},
							"annotations": map[string]interface{}{
								"kubernetes.io/ingress.class": "alb",
								"cosmo/ingress-patch-enable":  "true",
							},
						},
						"spec": map[string]interface{}{
							"host": "example.com",
						},
					},
				},
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": map[string]interface{}{
							"name":      "---",
							"namespace": "default",
						},
						"spec": map[string]interface{}{
							"hello": "world",
						},
					},
				},
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "ConfigMap",
						"metadata": map[string]interface{}{
							"name": "files",
						},
						"data": map[string]interface{}{
							"file1": "---\naaa\n---\nbbb\n",
							"file2": "---\nccc\n---\nddd\n",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "err",
			fields: fields{
				data: `no data`,
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"{{TEST}}": "OK"},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &RawYAMLBuilder{
				rawYaml: tt.fields.data,
				inst:    tt.fields.inst,
			}
			got, err := tr.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("RawYAMLBuilder.Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RawYAMLBuilder.Build() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawYAMLBuilder_ReplaceDefaultVars(t *testing.T) {
	type fields struct {
		rawYaml string
		inst    *cosmov1alpha1.Instance
	}
	tests := []struct {
		name   string
		fields fields
		want   *RawYAMLBuilder
	}{
		{
			name: "OK",
			fields: fields{
				rawYaml: "{{INSTANCE}}-{{NAMESPACE}}-{{TEMPLATE}}",
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"{{TEST}}": "OK"},
					},
				},
			},
			want: &RawYAMLBuilder{
				rawYaml: "cs1-cosmo-user-tom-code-server",
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"{{TEST}}": "OK"},
					},
				},
			},
		},
		{
			name: "without brackets",
			fields: fields{
				rawYaml: "{{INSTANCE}}-{{NAMESPACE}}-{{TEMPLATE}}",
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"TEST": "OK"},
					},
				},
			},
			want: &RawYAMLBuilder{
				rawYaml: "cs1-cosmo-user-tom-code-server",
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"TEST": "OK"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &RawYAMLBuilder{
				rawYaml: tt.fields.rawYaml,
				inst:    tt.fields.inst,
			}
			if got := tr.ReplaceDefaultVars(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RawYAMLBuilder.ReplaceDefaultVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawYAMLBuilder_ReplaceCustomVars(t *testing.T) {
	type fields struct {
		rawYaml string
		inst    *cosmov1alpha1.Instance
	}
	tests := []struct {
		name   string
		fields fields
		want   *RawYAMLBuilder
	}{
		{
			name: "OK",
			fields: fields{
				rawYaml: "{{INSTANCE}}-{{NAMESPACE}}-{{TEMPLATE}}-{{TEST}}",
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"{{TEST}}": "OK"},
					},
				},
			},
			want: &RawYAMLBuilder{
				rawYaml: "{{INSTANCE}}-{{NAMESPACE}}-{{TEMPLATE}}-OK",
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cs1",
						Namespace: "cosmo-user-tom",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "code-server",
						},
						Override: cosmov1alpha1.OverrideSpec{},
						Vars:     map[string]string{"{{TEST}}": "OK"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &RawYAMLBuilder{
				rawYaml: tt.fields.rawYaml,
				inst:    tt.fields.inst,
			}
			if got := tr.ReplaceCustomVars(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RawYAMLBuilder.ReplaceCustomVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidCustomVars(t *testing.T) {
	type args struct {
		varString string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				varString: "{{INSTACE}}",
			},
			wantErr: false,
		},
		{
			name: "Invalid",
			args: args{
				varString: "INSTACE",
			},
			wantErr: true,
		},
		{
			name: "Invalid sufix",
			args: args{
				varString: "{{INSTACE)",
			},
			wantErr: true,
		},
		{
			name: "Invalid prefix",
			args: args{
				varString: "$(INSTACE}}",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidCustomVars(tt.args.varString); (err != nil) != tt.wantErr {
				t.Errorf("ValidCustomVars() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFixupTemplateVarKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "OK",
			args: args{
				key: "INSTANCE",
			},
			want: "{{INSTANCE}}",
		},
		{
			name: "Valid prefix",
			args: args{
				key: "{{INSTANCE",
			},
			want: "{{INSTANCE}}",
		},
		{
			name: "Valid sufix",
			args: args{
				key: "INSTANCE}}",
			},
			want: "{{INSTANCE}}",
		},
		{
			name: "No change",
			args: args{
				key: "{{INSTANCE}}",
			},
			want: "{{INSTANCE}}",
		},
		{
			name: "No change 2",
			args: args{
				key: "{INSTA{{NCE}}}",
			},
			want: "{{{INSTA{{NCE}}}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FixupTemplateVarKey(tt.args.key); got != tt.want {
				t.Errorf("FixupTemplateVarKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
