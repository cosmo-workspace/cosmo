package wscfg

import (
	"reflect"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestSetConfigOnTemplateAnnotations(t *testing.T) {
	type args struct {
		cfg wsv1alpha1.Config
		obj *cosmov1alpha1.Template
	}
	tests := []struct {
		name string
		args args
		want *cosmov1alpha1.Template
	}{
		{
			name: "OK",
			args: args{
				cfg: wsv1alpha1.Config{
					DeploymentName:      "workspace1",
					ServiceName:         "workspace2",
					IngressName:         "workspace3",
					ServiceMainPortName: "main",
				},
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tmpl",
						Annotations: map[string]string{
							wsv1alpha1.TemplateAnnKeyWorkspaceService: "workspace",
							wsv1alpha1.TemplateAnnKeyURLBase:          "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
						},
					},
				},
			},
			want: &cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tmpl",
					Annotations: map[string]string{
						wsv1alpha1.TemplateAnnKeyWorkspaceDeployment:      "workspace1",
						wsv1alpha1.TemplateAnnKeyWorkspaceService:         "workspace2",
						wsv1alpha1.TemplateAnnKeyWorkspaceIngress:         "workspace3",
						wsv1alpha1.TemplateAnnKeyWorkspaceServiceMainPort: "main",
						wsv1alpha1.TemplateAnnKeyURLBase:                  "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
					},
				},
			},
		},
		{
			name: "no annotations",
			args: args{
				cfg: wsv1alpha1.Config{
					DeploymentName:      "workspace1",
					ServiceName:         "workspace2",
					IngressName:         "workspace3",
					ServiceMainPortName: "main",
				},
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tmpl",
					},
				},
			},
			want: &cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tmpl",
					Annotations: map[string]string{
						wsv1alpha1.TemplateAnnKeyWorkspaceDeployment:      "workspace1",
						wsv1alpha1.TemplateAnnKeyWorkspaceService:         "workspace2",
						wsv1alpha1.TemplateAnnKeyWorkspaceIngress:         "workspace3",
						wsv1alpha1.TemplateAnnKeyWorkspaceServiceMainPort: "main",
						wsv1alpha1.TemplateAnnKeyURLBase:                  "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetConfigOnTemplateAnnotations(tt.args.obj, tt.args.cfg)
		})
	}
}

func TestConfigFromTemplateAnnotations(t *testing.T) {
	type args struct {
		obj *cosmov1alpha1.Template
	}
	tests := []struct {
		name    string
		args    args
		want    wsv1alpha1.Config
		wantErr error
	}{
		{
			name: "found",
			args: args{
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tmpl1",
						Namespace: wsv1alpha1.UserNamespace("tom"),
						Labels: map[string]string{
							cosmov1alpha1.TemplateLabelKeyType: wsv1alpha1.TemplateTypeWorkspace,
						},
						Annotations: map[string]string{
							wsv1alpha1.TemplateAnnKeyWorkspaceDeployment:      "workspace1",
							wsv1alpha1.TemplateAnnKeyWorkspaceService:         "workspace2",
							wsv1alpha1.TemplateAnnKeyWorkspaceIngress:         "workspace3",
							wsv1alpha1.TemplateAnnKeyWorkspaceServiceMainPort: "main",
							wsv1alpha1.TemplateAnnKeyURLBase:                  "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
						},
					},
				},
			},
			want: wsv1alpha1.Config{
				DeploymentName:      "workspace1",
				ServiceName:         "workspace2",
				IngressName:         "workspace3",
				ServiceMainPortName: "main",
				URLBase:             "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
			},
		},
		{
			name: "no label",
			args: args{
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tmpl1",
						Namespace: wsv1alpha1.UserNamespace("tom"),
						Annotations: map[string]string{
							wsv1alpha1.TemplateAnnKeyWorkspaceDeployment:      "workspace1",
							wsv1alpha1.TemplateAnnKeyWorkspaceService:         "workspace2",
							wsv1alpha1.TemplateAnnKeyWorkspaceIngress:         "workspace3",
							wsv1alpha1.TemplateAnnKeyWorkspaceServiceMainPort: "main",
							wsv1alpha1.TemplateAnnKeyURLBase:                  "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
						},
					},
				},
			},
			wantErr: ErrNotTypeWorkspace,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConfigFromTemplateAnnotations(tt.args.obj)
			if err != tt.wantErr {
				t.Errorf("ConfigFromTemplateAnnotations() gotErr = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ConfigFromTemplateAnnotations() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func preTemplateBuild(rawTmpl string) ([]unstructured.Unstructured, error) {
	var inst cosmov1alpha1.Instance
	inst.SetName("dummy")
	inst.SetNamespace("dummy")

	builder := template.NewUnstructuredBuilder(rawTmpl, &inst)
	return builder.Build()
}
