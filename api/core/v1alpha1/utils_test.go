package v1alpha1

import (
	"testing"

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
