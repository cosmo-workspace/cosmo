package apiconv

import (
	"reflect"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/google/go-cmp/cmp"
)

func TestToYAML(t *testing.T) {
	type args struct {
		obj client.Object
	}
	tests := []struct {
		name string
		args args
		want *string
	}{
		{
			name: "OK",
			args: args{
				obj: &cosmov1alpha1.Instance{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Instance",
						APIVersion: "cosmo-workspace.github.io/v1alpha1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "test-tmpl",
						},
					},
				},
			},
			want: ptr.To(`apiVersion: cosmo-workspace.github.io/v1alpha1
kind: Instance
metadata:
  creationTimestamp: null
  name: test
spec:
  override: {}
  template:
    name: test-tmpl
status: {}
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToYAML(tt.args.obj); *got != *tt.want {
				t.Errorf("ToYAML() = %v, want %v\ndiff: %s", *got, *tt.want, cmp.Diff(*got, *tt.want))
			}
		})
	}
}

func TestDecodeYAML(t *testing.T) {
	type args struct {
		raw string
		obj client.Object
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK: Instance",
			args: args{
				raw: `apiVersion: cosmo-workspace.github.io/v1alpha1
kind: Instance
metadata:
  creationTimestamp: null
  name: test
spec:
  override: {}
  template:
    name: test-tmpl
status: {}
`,
				obj: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "test-tmpl",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "OK: Deployment",
			args: args{
				raw: `apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  name: test
spec:
  replicas: 1
  selector:`,
				obj: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "test",
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: ptr.To(int32(1)),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DecodeYAML(tt.args.raw, tt.args.obj); (err != nil) != tt.wantErr {
				t.Errorf("DecodeYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_removeUnnecessaryFields(t *testing.T) {
	now := time.Now()
	type args struct {
		obj *appsv1.Deployment
	}
	tests := []struct {
		name string
		args args
		want *appsv1.Deployment
	}{
		{
			name: "OK: Deployment",
			args: args{
				obj: &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"kubectl.kubernetes.io/restartedAt": "2023-11-04T04:42:34Z",
						},
						CreationTimestamp: metav1.NewTime(now),
						GenerateName:      "code-server-dev-code-server-598c87f6f6-",
						Labels: map[string]string{
							"app.kubernetes.io/instance": "code-server",
							"app.kubernetes.io/name":     "dev-code-server",
							"pod-template-hash":          "598c87f6f6",
						},
						Name:      "code-server-dev-code-server-598c87f6f6-qtwj2",
						Namespace: "code-server",
						OwnerReferences: []metav1.OwnerReference{
							{
								APIVersion:         "apps/v1",
								BlockOwnerDeletion: ptr.To(true),
								Controller:         ptr.To(true),
								Kind:               "ReplicaSet",
								Name:               "code-server-dev-code-server-598c87f6f6",
								UID:                "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
							},
						},
						ResourceVersion: "102253865",
						UID:             "beab71d2-d500-4655-bebf-dea1335989bb",
						ManagedFields: []metav1.ManagedFieldsEntry{
							{
								Manager:    "kubectl-client-side-apply",
								Operation:  "Apply",
								FieldsType: "FieldsV1",
								FieldsV1: &metav1.FieldsV1{
									Raw: []byte(`{"f:metadata":{"f:annotations":{"f:kubectl.kubernetes.io/restartedAt":{}},"f:creationTimestamp":{},"f:generateName":{},"f:labels":{"f:app.kubernetes.io/instance":{},"f:app.kubernetes.io/name":{},"f:pod-template-hash":{}},"f:name":{},"f:namespace":{},"f:ownerReferences":{"k:{"f:apiVersion":{},"f:blockOwnerDeletion":{},"f:controller":{},"f:kind":{},"f:name":{},"f:uid":{}}}},"f:spec":{"f:replicas":{}}}`),
								},
								APIVersion: "apps/v1",
							},
						},
					},
					Spec: appsv1.DeploymentSpec{
						Replicas: ptr.To(int32(1)),
					},
				},
			},
			want: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubectl.kubernetes.io/restartedAt": "2023-11-04T04:42:34Z",
					},
					CreationTimestamp: metav1.NewTime(now),
					GenerateName:      "code-server-dev-code-server-598c87f6f6-",
					Labels: map[string]string{
						"app.kubernetes.io/instance": "code-server",
						"app.kubernetes.io/name":     "dev-code-server",
						"pod-template-hash":          "598c87f6f6",
					},
					Name:      "code-server-dev-code-server-598c87f6f6-qtwj2",
					Namespace: "code-server",
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion:         "apps/v1",
							BlockOwnerDeletion: ptr.To(true),
							Controller:         ptr.To(true),
							Kind:               "ReplicaSet",
							Name:               "code-server-dev-code-server-598c87f6f6",
							UID:                "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
						},
					},
					ResourceVersion: "102253865",
					UID:             "beab71d2-d500-4655-bebf-dea1335989bb",
				},
				Spec: appsv1.DeploymentSpec{
					Replicas: ptr.To(int32(1)),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeUnnecessaryFields(tt.args.obj); !reflect.DeepEqual(*got.(*appsv1.Deployment), *tt.want) {
				t.Errorf("removeUnnecessaryFields() = %v, want %v\ndiff: %s", *got.(*appsv1.Deployment), *tt.want, cmp.Diff(*got.(*appsv1.Deployment), *tt.want))
			}
		})
	}
}
