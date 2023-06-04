package v1alpha1

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetControllerManaged(t *testing.T) {
	type args struct {
		obj LabelHolder
	}
	tests := []struct {
		name string
		args args
		want LabelHolder
	}{
		{
			name: "no labels",
			args: args{
				obj: &corev1.Secret{},
			},
			want: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelControllerManaged: "1",
					},
				},
			},
		},
		{
			name: "has labels",
			args: args{
				obj: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"KEY": "VAL",
						},
					},
				},
			},
			want: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"KEY":                  "VAL",
						LabelControllerManaged: "1",
					},
				},
			},
		},
		{
			name: "has managed labels with other value",
			args: args{
				obj: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							LabelControllerManaged: "0",
						},
					},
				},
			},
			want: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						LabelControllerManaged: "1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetControllerManaged(tt.args.obj)
			if !reflect.DeepEqual(tt.args.obj, tt.want) {
				t.Errorf("SetControllerManaged() = %v, want %v", tt.args.obj, tt.want)
			}
		})
	}
}
