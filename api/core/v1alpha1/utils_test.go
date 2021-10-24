package v1alpha1

import (
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Test_InstanceResourceName(t *testing.T) {
	type args struct {
		instanceName string
		resourceName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "OK",
			args: args{
				instanceName: "inst",
				resourceName: "res",
			},
			want: "inst-res",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InstanceResourceName(tt.args.instanceName, tt.args.resourceName); got != tt.want {
				t.Errorf("InstanceResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnstructuredToResourceRef(t *testing.T) {
	creationTimestamp := "2021-07-13T01:50:08Z"
	creationTime, err := time.Parse("2006-01-02T03:04:05Z", creationTimestamp)
	if err != nil {
		t.Fatal(err)
	}
	creationTime = creationTime.Local()
	metaCreationTime := metav1.NewTime(creationTime)

	now := v1.Now()

	type args struct {
		obj             unstructured.Unstructured
		updateTimestamp v1.Time
	}
	tests := []struct {
		name string
		args args
		want ObjectRef
	}{
		{
			name: "OK",
			args: args{
				obj: unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "networking.k8s.io/v1",
						"kind":       "Ingress",
						"metadata": map[string]interface{}{
							"name":              "test",
							"namespace":         "default",
							"creationTimestamp": "2021-07-13T01:50:08Z",
						},
					},
				},
				updateTimestamp: now,
			},
			want: ObjectRef{
				APIVersion:        "networking.k8s.io/v1",
				Kind:              "Ingress",
				Name:              "test",
				Namespace:         "default",
				CreationTimestamp: &metaCreationTime,
				UpdateTimestamp:   &now,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnstructuredToResourceRef(tt.args.obj, tt.args.updateTimestamp); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnstructuredToResourceRef() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInstanceResourceName(t *testing.T) {
	type args struct {
		instanceName string
		resourceName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InstanceResourceName(tt.args.instanceName, tt.args.resourceName); got != tt.want {
				t.Errorf("InstanceResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEqualInstanceResourceName(t *testing.T) {
	type args struct {
		instanceName string
		a            string
		b            string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EqualInstanceResourceName(tt.args.instanceName, tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("EqualInstanceResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsGVKEqual(t *testing.T) {
	type args struct {
		a schema.GroupVersionKind
		b schema.GroupVersionKind
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGVKEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("IsGVKEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExistInLastApplyed(t *testing.T) {
	type args struct {
		inst   Instance
		gvkObj gvkObject
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExistInLastApplyed(tt.args.inst, tt.args.gvkObj); got != tt.want {
				t.Errorf("ExistInLastApplyed() = %v, want %v", got, tt.want)
			}
		})
	}
}
