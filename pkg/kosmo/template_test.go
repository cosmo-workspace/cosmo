package kosmo

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
)

func TestClient_ListTemplates(t *testing.T) {
	type fields struct {
		Client client.Client
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []cosmov1alpha1.Template
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				Client: k8sFakeClient,
			},
			args: args{
				ctx: context.TODO(),
			},
			want:    []cosmov1alpha1.Template{*tmpl1, *tmpl2},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Client: tt.fields.Client,
			}
			got, err := c.ListTemplates(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListTemplates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListTemplates() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
