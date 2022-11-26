package template

import (
	"reflect"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestBuildObjects(t *testing.T) {
	type args struct {
		tmplSpec cosmov1alpha1.TemplateSpec
		inst     cosmov1alpha1.InstanceObject
	}
	tests := []struct {
		name        string
		args        args
		wantObjects []unstructured.Unstructured
		wantErr     bool
	}{
		{
			name: "Build raw yaml",
			wantObjects: []unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "apps/v1",
						"kind":       "Deployment",
						"metadata": map[string]interface{}{
							"name": "deploy",
						},
						"spec": map[string]interface{}{
							"template": map[string]interface{}{
								"spec": map[string]interface{}{
									"containers": []interface{}{
										map[string]interface{}{
											"name": "app",
											"command": []interface{}{
												"sh", "-c", "echo default/inst",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			args: args{
				tmplSpec: cosmov1alpha1.TemplateSpec{
					RawYaml: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
spec:
  template:
    spec:
      containers:
        - name: "app"
          command:
            - sh
            - -c
            - echo {{NAMESPACE}}/{{INSTANCE}}
`,
				},
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inst",
						Namespace: "default",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Vars: map[string]string{
							"FOO": "BAR",
						},
					},
				},
			},
		},
		{
			name: "Build raw yaml failed",

			wantErr: true,
			args: args{
				tmplSpec: cosmov1alpha1.TemplateSpec{
					RawYaml: `{}`,
				},
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inst",
						Namespace: "default",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Vars: map[string]string{
							"FOO": "BAR",
						},
					},
				},
			},
		},
		{
			name:    "Invalid template",
			wantErr: true,
			args: args{
				tmplSpec: cosmov1alpha1.TemplateSpec{},
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inst",
						Namespace: "default",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Vars: map[string]string{
							"FOO": "BAR",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotObjects, err := BuildObjects(tt.args.tmplSpec, tt.args.inst)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildObjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotObjects, tt.wantObjects) {
				t.Errorf("BuildObjects() = %v, want %v", gotObjects, tt.wantObjects)
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
