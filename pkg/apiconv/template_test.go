package apiconv

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

func TestC2D_Templates(t *testing.T) {
	type args struct {
		tmpls []cosmov1alpha1.TemplateObject
		opts  []TemplateConvertOptions
	}
	tests := []struct {
		name string
		args args
		want []*dashv1alpha1.Template
	}{
		{
			name: "empty",
			args: args{
				tmpls: []cosmov1alpha1.TemplateObject{
					&cosmov1alpha1.Template{
						ObjectMeta: metav1.ObjectMeta{
							Name: "tmpl1",
							Annotations: map[string]string{
								cosmov1alpha1.TemplateAnnKeyDisableNamePrefix: "true",
								cosmov1alpha1.TemplateAnnKeyRequiredAddons:    "xxx,yyy",
								cosmov1alpha1.TemplateAnnKeyUserRoles:         "aaa,bbb",
							},
						},
						Spec: cosmov1alpha1.TemplateSpec{
							Description: "tmpl1 desc",
							RequiredVars: []cosmov1alpha1.RequiredVarSpec{
								{
									Var:     "var1",
									Default: "def1",
								},
								{
									Var: "var2",
								},
							},
						},
					},
					&cosmov1alpha1.ClusterTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Name: "tmpl2",
						},
						Spec: cosmov1alpha1.TemplateSpec{
							Description: "tmpl2 desc",
						},
					},
				},
			},
			want: []*dashv1alpha1.Template{
				{
					Name:        "tmpl1",
					Description: "tmpl1 desc",
					RequiredVars: []*dashv1alpha1.TemplateRequiredVars{
						{
							VarName:      "var1",
							DefaultValue: "def1",
						},
						{
							VarName: "var2",
						},
					},
					RequiredUseraddons: []string{
						"xxx",
						"yyy",
					},
					Userroles: []string{
						"aaa",
						"bbb",
					},
				},
				{
					Name:           "tmpl2",
					Description:    "tmpl2 desc",
					IsClusterScope: true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := C2D_Templates(tt.args.tmpls, tt.args.opts...)
			newGot := make([]string, len(got))
			for _, v := range got {
				newGot = append(newGot, v.String())
			}
			want := make([]string, len(tt.want))
			for _, v := range tt.want {
				want = append(want, v.String())
			}
			if !slices.Equal(want, newGot) {
				t.Errorf("C2D_Templates() = %v, want %v\ndiff = %v", newGot, want, cmp.Diff(want, newGot))
			}
		})
	}
}

func TestC2D_Template(t *testing.T) {
	type args struct {
		tmpl cosmov1alpha1.TemplateObject
		opts []TemplateConvertOptions
	}
	tests := []struct {
		name string
		args args
		want *dashv1alpha1.Template
	}{
		{
			name: "OK",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tmpl1",
						Annotations: map[string]string{
							cosmov1alpha1.TemplateAnnKeyDisableNamePrefix:         "true",
							cosmov1alpha1.TemplateAnnKeyRequiredAddons:            "xxx,yyy",
							cosmov1alpha1.TemplateAnnKeyUserRoles:                 "aaa,bbb",
							cosmov1alpha1.UserAddonTemplateAnnKeyDefaultUserAddon: "true",
						},
					},
					Spec: cosmov1alpha1.TemplateSpec{
						Description: "tmpl1 desc",
						RequiredVars: []cosmov1alpha1.RequiredVarSpec{
							{
								Var:     "var1",
								Default: "def1",
							},
							{
								Var: "var2",
							},
						},
					},
				},
			},
			want: &dashv1alpha1.Template{
				Name:        "tmpl1",
				Description: "tmpl1 desc",
				RequiredVars: []*dashv1alpha1.TemplateRequiredVars{
					{
						VarName:      "var1",
						DefaultValue: "def1",
					},
					{
						VarName: "var2",
					},
				},
				IsDefaultUserAddon: ptr.To(true),
				RequiredUseraddons: []string{
					"xxx",
					"yyy",
				},
				Userroles: []string{
					"aaa",
					"bbb",
				},
			},
		},
		{
			name: "OK",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tmpl1",
						Annotations: map[string]string{
							cosmov1alpha1.TemplateAnnKeyDisableNamePrefix:         "true",
							cosmov1alpha1.TemplateAnnKeyRequiredAddons:            "xxx,yyy",
							cosmov1alpha1.TemplateAnnKeyUserRoles:                 "aaa,bbb",
							cosmov1alpha1.UserAddonTemplateAnnKeyDefaultUserAddon: "true",
						},
					},
					Spec: cosmov1alpha1.TemplateSpec{
						Description: "tmpl1 desc",
						RequiredVars: []cosmov1alpha1.RequiredVarSpec{
							{
								Var:     "var1",
								Default: "def1",
							},
							{
								Var: "var2",
							},
						},
					},
				},
				opts: []TemplateConvertOptions{
					WithTemplateRaw(ptr.To(true)),
				},
			},
			want: &dashv1alpha1.Template{
				Name:        "tmpl1",
				Description: "tmpl1 desc",
				RequiredVars: []*dashv1alpha1.TemplateRequiredVars{
					{
						VarName:      "var1",
						DefaultValue: "def1",
					},
					{
						VarName: "var2",
					},
				},
				IsDefaultUserAddon: ptr.To(true),
				RequiredUseraddons: []string{
					"xxx",
					"yyy",
				},
				Userroles: []string{
					"aaa",
					"bbb",
				},
				Raw: ptr.To(`apiVersion: cosmo-workspace.github.io/v1alpha1
kind: Template
metadata:
  annotations:
    cosmo-workspace.github.io/disable-nameprefix: "true"
    cosmo-workspace.github.io/required-useraddons: xxx,yyy
    cosmo-workspace.github.io/userroles: aaa,bbb
    useraddon.cosmo-workspace.github.io/default: "true"
  creationTimestamp: null
  name: tmpl1
spec:
  description: tmpl1 desc
  requiredVars:
  - default: def1
    var: var1
  - var: var2
`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := C2D_Template(tt.args.tmpl, tt.args.opts...); !reflect.DeepEqual(got.String(), tt.want.String()) {
				t.Errorf("C2D_Template() raw diff = %v", cmp.Diff(*got.Raw, *tt.want.Raw))
				t.Errorf("C2D_Template() obj diff = %v", cmp.Diff(got.String(), tt.want.String()))
				t.Errorf("C2D_Template() = %v, want %v", got, tt.want)
			}
		})
	}
}
