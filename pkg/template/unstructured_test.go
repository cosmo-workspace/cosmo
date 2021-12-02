package template

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
)

func TestNewUnstructuredBuilder(t *testing.T) {
	type args struct {
		data string
		inst *cosmov1alpha1.Instance
	}
	tests := []struct {
		name string
		args args
		want *UnstructuredBuilder
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
			want: &UnstructuredBuilder{
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
			if got := NewUnstructuredBuilder(tt.args.data, tt.args.inst); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUnstructuredBuilder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnstructuredBuilder_Build(t *testing.T) {
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
			tr := &UnstructuredBuilder{
				rawYaml: tt.fields.data,
				inst:    tt.fields.inst,
			}
			got, err := tr.Build()
			if (err != nil) != tt.wantErr {
				t.Errorf("UnstructuredBuilder.Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnstructuredBuilder.Build() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToUnstructured(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		wantGvk schema.GroupVersionKind
		want    unstructured.Unstructured
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				str: `apiVersion: networking.k8s.io/v1
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
  host: example.com`},
			wantGvk: schema.GroupVersionKind{Kind: "XXXX", Group: "networking.k8s.io", Version: "v1"},
			want: unstructured.Unstructured{
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGvk, got, err := StringToUnstructured(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringToUnstructured() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(*gotGvk, tt.wantGvk) {
				t.Errorf("StringToUnstructured() gotGvk = %v, wantGvk %v", *gotGvk, tt.wantGvk)
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("StringToUnstructured() got = %v, want %v", got, &tt.want)
			}
		})
	}
}

func TestUnstructuredToJSONBytes(t *testing.T) {
	type args struct {
		obj *unstructured.Unstructured
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				obj: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "networking.k8s.io/v1",
						"kind":       "Ingress",
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
			},
			want:    `{"apiVersion":"networking.k8s.io/v1","kind":"Ingress","metadata":{"annotations":{"cosmo/ingress-patch-enable":"true","kubernetes.io/ingress.class":"alb"},"labels":{"key":"val"},"name":"test","namespace":"default"},"spec":{"host":"example.com"}}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnstructuredToJSONBytes(tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnstructuredToJSONBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("UnstructuredToJSONBytes() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
