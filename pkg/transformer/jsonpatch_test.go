package transformer

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

func TestJSONPatchTransformer_Transform(t *testing.T) {
	type fields struct {
		patch    []cosmov1alpha1.Json6902
		instName string
	}
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "No target",
			fields: fields{
				patch: []cosmov1alpha1.Json6902{
					{
						Target: cosmov1alpha1.ObjectRef{
							ObjectReference: corev1.ObjectReference{
								APIVersion: metav1.GroupVersion{
									Group:   netv1.GroupName,
									Version: "v1",
								}.String(),
								Kind: "Ingress",
								Name: "test",
							},
						},
						Patch: `[
  {
    "op": "replace",
    "path": "/spec/ports/1/port",
    "value": 9999
  }
]
`,
					},
				},
				instName: "instance",
			},
			args: args{
				src: `apiVersion: v1
kind: Service
metadata:
  name: instance-test
  namespace: default
spec:
  ports:
  - name: port1
    port: 8080
    protocol: TCP
  - name: port2
    port: 8081
    protocol: TCP
  type: ClusterIP
`,
			},
			want: `apiVersion: v1
kind: Service
metadata:
  name: instance-test
  namespace: default
spec:
  ports:
  - name: port1
    port: 8080
    protocol: TCP
  - name: port2
    port: 8081
    protocol: TCP
  type: ClusterIP
`,
			wantErr: false,
		},
		{
			name: "Single replace patch",
			fields: fields{
				patch: []cosmov1alpha1.Json6902{
					{
						Target: cosmov1alpha1.ObjectRef{
							ObjectReference: corev1.ObjectReference{
								APIVersion: metav1.GroupVersion{
									Group:   "",
									Version: "v1",
								}.String(),
								Kind: "Service",
								Name: "test",
							},
						},
						Patch: `[
  {
    "op": "replace",
    "path": "/spec/ports/1/port",
    "value": 9999
  }
]
`,
					},
				},
				instName: "instance",
			},
			args: args{
				src: `apiVersion: v1
kind: Service
metadata:
  name: instance-test
  namespace: default
spec:
  ports:
  - name: port1
    port: 8080
    protocol: TCP
  - name: port2
    port: 8081
    protocol: TCP
  type: ClusterIP
`,
			},
			want: `apiVersion: v1
kind: Service
metadata:
  name: instance-test
  namespace: default
spec:
  ports:
  - name: port1
    port: 8080
    protocol: TCP
  - name: port2
    port: 9999
    protocol: TCP
  type: ClusterIP
`,
			wantErr: false,
		},
		{
			name: "Multiple replace patch",
			fields: fields{
				patch: []cosmov1alpha1.Json6902{
					{
						Target: cosmov1alpha1.ObjectRef{
							ObjectReference: corev1.ObjectReference{
								APIVersion: metav1.GroupVersion{
									Group:   "",
									Version: "v1",
								}.String(),
								Kind:      "Service",
								Name:      "test",
								Namespace: "default",
							},
						},
						Patch: `[
  {
    "op": "replace",
    "path": "/spec/ports/1/port",
    "value": 9999
  },
  {
    "op": "add",
    "path": "/spec/ports/-",
    "value": {"name":"port3","port":7777,"protocol":"TCP"}
  }
]
`,
					},
				},
				instName: "instance",
			},
			args: args{
				src: `apiVersion: v1
kind: Service
metadata:
  name: instance-test
  namespace: default
spec:
  ports:
  - name: port1
    port: 8080
    protocol: TCP
  - name: port2
    port: 8081
    protocol: TCP
  type: ClusterIP
`,
			},
			want: `apiVersion: v1
kind: Service
metadata:
  name: instance-test
  namespace: default
spec:
  ports:
  - name: port1
    port: 8080
    protocol: TCP
  - name: port2
    port: 9999
    protocol: TCP
  - name: port3
    port: 7777
    protocol: TCP
  type: ClusterIP
`,
			wantErr: false,
		},
		{
			name: "Invalid JSONPath format",
			fields: fields{
				patch: []cosmov1alpha1.Json6902{
					{
						Target: cosmov1alpha1.ObjectRef{
							ObjectReference: corev1.ObjectReference{
								APIVersion: metav1.GroupVersion{
									Group:   "",
									Version: "v1",
								}.String(),
								Kind:      "Service",
								Name:      "test",
								Namespace: "default",
							},
						},
						Patch: `[
  {
    "op": "replace",
    "path": "/invalid",
    "value": 9999
  }
]
`,
					},
				},
				instName: "instance",
			},
			args: args{
				src: `apiVersion: v1
kind: Service
metadata:
  name: instance-test
  namespace: default
spec:
  ports:
  - name: port1
    port: 8080
    protocol: TCP
  - name: port2
    port: 8081
    protocol: TCP
  type: ClusterIP
`,
			},
			want: `null
`,
			wantErr: true,
		},
		{
			name: "Missing target",
			fields: fields{
				patch: []cosmov1alpha1.Json6902{
					{
						Target: cosmov1alpha1.ObjectRef{
							ObjectReference: corev1.ObjectReference{
								Name: "test",
							},
						},
						Patch: `[
  {
    "op": "replace",
    "path": "/spec/ports/1/port",
    "value": 9999
  }
]
`,
					},
				},
				instName: "instance",
			},
			args: args{
				src: `apiVersion: v1
kind: Service
metadata:
  name: instance-test
  namespace: default
spec:
  ports:
  - name: port1
    port: 8080
    protocol: TCP
  - name: port2
    port: 8081
    protocol: TCP
  type: ClusterIP
`,
			},
			want: `apiVersion: v1
kind: Service
metadata:
  name: instance-test
  namespace: default
spec:
  ports:
  - name: port1
    port: 8080
    protocol: TCP
  - name: port2
    port: 8081
    protocol: TCP
  type: ClusterIP
`,
			wantErr: false,
		},
		{
			name: "Operation error",
			fields: fields{
				patch: []cosmov1alpha1.Json6902{
					{
						Target: cosmov1alpha1.ObjectRef{
							ObjectReference: corev1.ObjectReference{
								APIVersion: metav1.GroupVersion{
									Group:   "",
									Version: "v1",
								}.String(),
								Kind: "Service",
								Name: "test",
							},
						},
						Patch: `[
  {
    "op": "NoOperation",
    "path": "/spec/ports/1/port",
    "value": 9999
  }
]
`,
					},
				},
				instName: "instance",
			},
			args: args{
				src: `apiVersion: v1
kind: Service
metadata:
  name: instance-test
  namespace: default
spec:
  ports:
  - name: port1
    port: 8080
    protocol: TCP
  - name: port2
    port: 8081
    protocol: TCP
  type: ClusterIP
`,
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Patch format error",
			fields: fields{
				patch: []cosmov1alpha1.Json6902{
					{
						Target: cosmov1alpha1.ObjectRef{
							ObjectReference: corev1.ObjectReference{
								APIVersion: metav1.GroupVersion{
									Group:   "",
									Version: "v1",
								}.String(),
								Kind: "Service",
								Name: "test",
							},
						},
						Patch: `[
  {
    "op": "replace",,,
    "path": "/spec/ports/1/port",
    "value": 9999
  }
]
`,
					},
				},
				instName: "instance",
			},
			args: args{
				src: `apiVersion: v1
kind: Service
metadata:
  name: instance-test
  namespace: default
spec:
  ports:
  - name: port1
    port: 8080
    protocol: TCP
  - name: port2
    port: 8081
    protocol: TCP
  type: ClusterIP
`,
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewJSONPatchTransformer(tt.fields.patch, tt.fields.instName)
			_, obj, err := template.StringToUnstructured(tt.args.src)
			if err != nil {
				t.Errorf("JSONPatchTransformer.Transform() template.StringToUnstructured error = %v", err)
				return
			}
			gotObj, err := tr.Transform(obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONPatchTransformer.Transform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				got, err := yaml.Marshal(gotObj)
				if err != nil {
					t.Errorf("yaml.Marshal err = %v", err)
				}
				if string(got) != tt.want {
					t.Errorf("JSONPatchTransformer.Transform() = %v, want %v", string(got), tt.want)
					t.Errorf("JSONPatchTransformer.Transform() = %v, want %v", got, []byte(tt.want))
				}

			} else {
				if gotObj != nil {
					t.Errorf("JSONPatchTransformer.Transform() gotObj not nil %v", gotObj)
				}
			}
		})
	}
}
