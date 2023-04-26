package kosmo

import (
	"context"
	"reflect"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestClient_GetTemplate(t *testing.T) {
	type fields struct {
		Client client.Client
	}
	type args struct {
		ctx      context.Context
		tmplName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *cosmov1alpha1.Template
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				Client: k8sFakeClient,
			},
			args: args{
				ctx:      context.TODO(),
				tmplName: tmpl1.Name,
			},
			want:    tmpl1,
			wantErr: false,
		},
		{
			name: "not found",
			fields: fields{
				Client: k8sFakeClient,
			},
			args: args{
				ctx:      context.TODO(),
				tmplName: "notfound",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Client: tt.fields.Client,
			}
			got, err := c.GetTemplate(tt.args.ctx, tt.args.tmplName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil {
				tt.want.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   cosmov1alpha1.GroupVersion.Group,
					Version: cosmov1alpha1.GroupVersion.Version,
					Kind:    "Template",
				})
			}
			if !tt.wantErr && !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isAllowedToUseTemplate(t *testing.T) {
	type args struct {
		tmpl  cosmov1alpha1.TemplateObject
		roles []cosmov1alpha1.UserRole
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "no annotations, all roles are allowed",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "hogwarts-common",
					},
				},
				roles: []cosmov1alpha1.UserRole{
					{Name: "gryffindor-developer"},
				},
			},
			want: true,
		},
		{
			name: "forbidden if role is matched to forbidden role",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sword-of-gryffindor",
						Annotations: map[string]string{
							cosmov1alpha1.TemplateAnnKeyForbiddenUserRoles: "slytherin",
						},
					},
				},
				roles: []cosmov1alpha1.UserRole{
					{Name: "slytherin"},
				},
			},
			want: false,
		},
		{
			name: "forbidden if role is not matched to allowed role",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sword-of-gryffindor",
						Annotations: map[string]string{
							cosmov1alpha1.TemplateAnnKeyUserRoles: "gryffindor",
						},
					},
				},
				roles: []cosmov1alpha1.UserRole{
					{Name: "slytherin"},
				},
			},
			want: false,
		},
		{
			name: "forbidden if role is matched to allowed role but also matched to forbidden role",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sword-of-gryffindor",
						Annotations: map[string]string{
							cosmov1alpha1.TemplateAnnKeyForbiddenUserRoles: "slytherin",
							cosmov1alpha1.TemplateAnnKeyUserRoles:          "gryffindor",
						},
					},
				},
				roles: []cosmov1alpha1.UserRole{
					{Name: "slytherin"},
					{Name: "gryffindor"},
				},
			},
			want: false,
		},
		{
			name: "allowed if wildcard match for allowed role",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sword-of-gryffindor",
						Annotations: map[string]string{
							cosmov1alpha1.TemplateAnnKeyUserRoles: "gryffindor-*",
						},
					},
				},
				roles: []cosmov1alpha1.UserRole{
					{Name: "gryffindor-developer"},
				},
			},
			want: true,
		},
		{
			name: "forbidden if wildcard match for forbidden role",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sword-of-gryffindor",
						Annotations: map[string]string{
							cosmov1alpha1.TemplateAnnKeyUserRoles: "sly*",
						},
					},
				},
				roles: []cosmov1alpha1.UserRole{
					{Name: "slytherin"},
				},
			},
			want: true,
		},
		{
			name: "forbidden if allowed role wildcard not match",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sword-of-gryffindor",
						Annotations: map[string]string{
							cosmov1alpha1.TemplateAnnKeyUserRoles: "gryffindor-*",
						},
					},
				},
				roles: []cosmov1alpha1.UserRole{
					{Name: "gryffindor"},
				},
			},
			want: false,
		},
		{
			name: "forbidden if both allowed role wildcard and forbidden role matches",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sword-of-gryffindor",
						Annotations: map[string]string{
							cosmov1alpha1.TemplateAnnKeyUserRoles:          "gryffindor-*",
							cosmov1alpha1.TemplateAnnKeyForbiddenUserRoles: "gryffindor-faker",
						},
					},
				},
				roles: []cosmov1alpha1.UserRole{
					{Name: "gryffindor-faker"},
				},
			},
			want: false,
		},
		{
			name: "forbidden if both allowed role wildcard and forbidden wildcard matches",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sword-of-gryffindor",
						Annotations: map[string]string{
							cosmov1alpha1.TemplateAnnKeyUserRoles:          "gryffindor-*",
							cosmov1alpha1.TemplateAnnKeyForbiddenUserRoles: "gryffindor-f*",
						},
					},
				},
				roles: []cosmov1alpha1.UserRole{
					{Name: "gryffindor-faker"},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAllowedToUseTemplate(context.TODO(), tt.args.tmpl, tt.args.roles); got != tt.want {
				t.Errorf("isAllowedToUseTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filterTemplates(t *testing.T) {
	type args struct {
		tmpls []cosmov1alpha1.TemplateObject
		roles []cosmov1alpha1.UserRole
	}
	tests := []struct {
		name string
		args args
		want []cosmov1alpha1.TemplateObject
	}{
		{
			name: "filter",
			args: args{
				tmpls: []cosmov1alpha1.TemplateObject{
					&cosmov1alpha1.Template{
						ObjectMeta: metav1.ObjectMeta{
							Name:        "hogwarts-common",
							Annotations: map[string]string{},
						},
					},
					&cosmov1alpha1.Template{
						ObjectMeta: metav1.ObjectMeta{
							Name: "sword-of-gryffindor",
							Annotations: map[string]string{
								cosmov1alpha1.TemplateAnnKeyUserRoles: "gryffindor-*",
							},
						},
					},
					&cosmov1alpha1.Template{
						ObjectMeta: metav1.ObjectMeta{
							Name: "serpent-of-slytherin",
							Annotations: map[string]string{
								cosmov1alpha1.TemplateAnnKeyUserRoles: "slytherin",
							},
						},
					},
				},
				roles: []cosmov1alpha1.UserRole{
					{Name: "gryffindor-developer"},
				},
			},
			want: []cosmov1alpha1.TemplateObject{
				&cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "hogwarts-common",
						Annotations: map[string]string{},
					},
				},
				&cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "sword-of-gryffindor",
						Annotations: map[string]string{
							cosmov1alpha1.TemplateAnnKeyUserRoles: "gryffindor-*",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterTemplates(context.TODO(), tt.args.tmpls, tt.args.roles); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterTemplates() = %v, want %v", got, tt.want)
				t.Errorf(cmp.Diff(got, tt.want))
			}
		})
	}
}
