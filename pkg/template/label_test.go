package template

import (
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetTemplateType(t *testing.T) {
	type args struct {
		inst     *cosmov1alpha1.Instance
		tmplType string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "OK",
			args: args{
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name: "code-server",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
				},
				tmplType: "workspace",
			},
			want: map[string]string{
				"foo":                              "bar",
				cosmov1alpha1.TemplateLabelKeyType: "workspace",
			},
		},
		{
			name: "if exist override",
			args: args{
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name: "code-server",
						Labels: map[string]string{
							cosmov1alpha1.TemplateLabelKeyType: "workspace",
						},
					},
				},
				tmplType: "workspace",
			},
			want: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: "workspace",
			},
		},
		{
			name: "if no annotation set",
			args: args{
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name: "code-server",
					},
				},
				tmplType: "workspace",
			},
			want: map[string]string{
				cosmov1alpha1.TemplateLabelKeyType: "workspace",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetTemplateType(tt.args.inst, tt.args.tmplType)
		})
	}
}

func TestGetTemplateType(t *testing.T) {
	type args struct {
		tmpl *cosmov1alpha1.Template
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "found",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "code-server",
						Labels: map[string]string{
							"foo":                              "bar",
							cosmov1alpha1.TemplateLabelKeyType: "workspace",
						},
					},
				},
			},
			want:  "workspace",
			want1: true,
		},
		{
			name: "not found",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "code-server",
						Labels: map[string]string{
							"foo": "bar",
						},
					},
				},
			},
			want:  "",
			want1: false,
		},
		{
			name: "no Labels",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "code-server",
					},
				},
			},
			want:  "",
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetTemplateType(tt.args.tmpl)
			if got != tt.want {
				t.Errorf("GetTemplateType() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetTemplateType() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
