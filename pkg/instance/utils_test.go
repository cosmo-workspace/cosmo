package instance

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
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
			name: "fix",
			args: args{
				instanceName: "inst",
				resourceName: "res",
			},
			want: "inst-res",
		},
		{
			name: "no fix",
			args: args{
				instanceName: "inst",
				resourceName: "inst-res",
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
		{
			name: "fix a",
			args: args{
				instanceName: "inst",
				a:            "inst-res",
				b:            "res",
			},
			want: true,
		},
		{
			name: "no fix",
			args: args{
				instanceName: "inst",
				a:            "inst-res",
				b:            "inst-res",
			},
			want: true,
		},
		{
			name: "fix b",
			args: args{
				instanceName: "inst",
				a:            "res",
				b:            "inst-res",
			},
			want: true,
		},
		{
			name: "fix both",
			args: args{
				instanceName: "inst",
				a:            "res",
				b:            "res",
			},
			want: true,
		},
		{
			name: "no match",
			args: args{
				instanceName: "inst",
				a:            "instres",
				b:            "res",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EqualInstanceResourceName(tt.args.instanceName, tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("EqualInstanceResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExistInLastApplyed(t *testing.T) {
	type args struct {
		inst   cosmov1alpha1.Instance
		gvkObj GVKNameGetter
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Exist",
			want: true,
			args: args{
				inst: cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inst",
						Namespace: "default",
					},
					Status: cosmov1alpha1.InstanceStatus{
						LastApplied: []cosmov1alpha1.ObjectRef{
							{
								ObjectReference: corev1.ObjectReference{
									APIVersion: "v1",
									Kind:       "Service",
									Name:       "inst-svc",
									Namespace:  "defualt",
								},
							},
							{
								ObjectReference: corev1.ObjectReference{
									APIVersion: "apps/v1",
									Kind:       "Deployment",
									Name:       "inst-deploy",
									Namespace:  "defualt",
								},
							},
						},
					},
				},
				gvkObj: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inst-deploy",
						Namespace: "default",
					},
				},
			},
		},
		{
			name: "Not Exist",
			want: false,
			args: args{
				inst: cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inst",
						Namespace: "default",
					},
					Status: cosmov1alpha1.InstanceStatus{
						LastApplied: []cosmov1alpha1.ObjectRef{
							{
								ObjectReference: corev1.ObjectReference{
									APIVersion: "v1",
									Kind:       "Service",
									Name:       "inst-svc",
									Namespace:  "defualt",
								},
							},
							{
								ObjectReference: corev1.ObjectReference{
									APIVersion: "extention/v1",
									Kind:       "Deployment",
									Name:       "inst-deploy",
									Namespace:  "defualt",
								},
							},
						},
					},
				},
				gvkObj: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inst-deploy",
						Namespace: "default",
					},
				},
			},
		},
		{
			name: "No last applied",
			want: false,
			args: args{
				inst: cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inst",
						Namespace: "default",
					},
					Status: cosmov1alpha1.InstanceStatus{},
				},
				gvkObj: &appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inst-deploy",
						Namespace: "default",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExistInLastApplyed(&tt.args.inst, tt.args.gvkObj); got != tt.want {
				t.Errorf("ExistInLastApplyed() = %v, want %v", got, tt.want)
			}
		})
	}
}
