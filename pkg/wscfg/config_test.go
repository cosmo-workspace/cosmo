package wscfg

import (
	"reflect"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestSetConfigOnTemplateAnnotations(t *testing.T) {
	type args struct {
		cfg cosmov1alpha1.Config
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
				cfg: cosmov1alpha1.Config{
					DeploymentName:      "workspace1",
					ServiceName:         "workspace2",
					IngressName:         "workspace3",
					ServiceMainPortName: "main",
				},
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tmpl",
						Annotations: map[string]string{
							cosmov1alpha1.WorkspaceTemplateAnnKeyServiceName: "workspace",
							cosmov1alpha1.WorkspaceTemplateAnnKeyURLBase:     "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
						},
					},
				},
			},
			want: &cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tmpl",
					Annotations: map[string]string{
						cosmov1alpha1.WorkspaceTemplateAnnKeyDeploymentName:  "workspace1",
						cosmov1alpha1.WorkspaceTemplateAnnKeyServiceName:     "workspace2",
						cosmov1alpha1.WorkspaceTemplateAnnKeyIngressName:     "workspace3",
						cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort: "main",
						cosmov1alpha1.WorkspaceTemplateAnnKeyURLBase:         "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
					},
				},
			},
		},
		{
			name: "no annotations",
			args: args{
				cfg: cosmov1alpha1.Config{
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
						cosmov1alpha1.WorkspaceTemplateAnnKeyDeploymentName:  "workspace1",
						cosmov1alpha1.WorkspaceTemplateAnnKeyServiceName:     "workspace2",
						cosmov1alpha1.WorkspaceTemplateAnnKeyIngressName:     "workspace3",
						cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort: "main",
						cosmov1alpha1.WorkspaceTemplateAnnKeyURLBase:         "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
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
		want    cosmov1alpha1.Config
		wantErr error
	}{
		{
			name: "found",
			args: args{
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tmpl1",
						Namespace: cosmov1alpha1.UserNamespace("tom"),
						Labels: map[string]string{
							cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeWorkspace,
						},
						Annotations: map[string]string{
							cosmov1alpha1.WorkspaceTemplateAnnKeyDeploymentName:  "workspace1",
							cosmov1alpha1.WorkspaceTemplateAnnKeyServiceName:     "workspace2",
							cosmov1alpha1.WorkspaceTemplateAnnKeyIngressName:     "workspace3",
							cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort: "main",
							cosmov1alpha1.WorkspaceTemplateAnnKeyURLBase:         "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
						},
					},
				},
			},
			want: cosmov1alpha1.Config{
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
						Namespace: cosmov1alpha1.UserNamespace("tom"),
						Annotations: map[string]string{
							cosmov1alpha1.WorkspaceTemplateAnnKeyDeploymentName:  "workspace1",
							cosmov1alpha1.WorkspaceTemplateAnnKeyServiceName:     "workspace2",
							cosmov1alpha1.WorkspaceTemplateAnnKeyIngressName:     "workspace3",
							cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort: "main",
							cosmov1alpha1.WorkspaceTemplateAnnKeyURLBase:         "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
						},
					},
				},
			},
			wantErr: ErrNotTypeWorkspace,
		},
		{
			name: "not type workspace",
			args: args{
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tmpl1",
						Namespace: cosmov1alpha1.UserNamespace("tom"),
						Labels: map[string]string{
							cosmov1alpha1.TemplateLabelKeyType: "invalid",
						},
						Annotations: map[string]string{
							cosmov1alpha1.WorkspaceTemplateAnnKeyDeploymentName:  "workspace1",
							cosmov1alpha1.WorkspaceTemplateAnnKeyServiceName:     "workspace2",
							cosmov1alpha1.WorkspaceTemplateAnnKeyIngressName:     "workspace3",
							cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort: "main",
							cosmov1alpha1.WorkspaceTemplateAnnKeyURLBase:         "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
						},
					},
				},
			},
			wantErr: ErrNotTypeWorkspace,
		},
		{
			name: "no annotations",
			args: args{
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tmpl1",
						Namespace: cosmov1alpha1.UserNamespace("tom"),
						Labels: map[string]string{
							cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeWorkspace,
						},
					},
				},
			},
			want: cosmov1alpha1.Config{},
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

	builder := template.NewRawYAMLBuilder(rawTmpl, &inst)
	return builder.Build()
}
