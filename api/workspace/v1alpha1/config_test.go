package v1alpha1

import (
	"reflect"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetConfigOnTemplateAnnotations(t *testing.T) {
	type args struct {
		cfg Config
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
				cfg: Config{
					DeploymentName:      "workspace1",
					ServiceName:         "workspace2",
					IngressName:         "workspace3",
					ServiceMainPortName: "main",
				},
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tmpl",
						Annotations: map[string]string{
							InstanceAnnKeyWorkspaceService: "workspace",
							InstanceAnnKeyURLBase:          "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain{{NETRULE_PATH}}",
						},
					},
				},
			},
			want: &cosmov1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tmpl",
					Annotations: map[string]string{
						InstanceAnnKeyWorkspaceDeployment:      "workspace1",
						InstanceAnnKeyWorkspaceService:         "workspace2",
						InstanceAnnKeyWorkspaceIngress:         "workspace3",
						InstanceAnnKeyWorkspaceServiceMainPort: "main",
						InstanceAnnKeyURLBase:                  "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain{{NETRULE_PATH}}",
					},
				},
			},
		},
		{
			name: "no annotations",
			args: args{
				cfg: Config{
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
						InstanceAnnKeyWorkspaceDeployment:      "workspace1",
						InstanceAnnKeyWorkspaceService:         "workspace2",
						InstanceAnnKeyWorkspaceIngress:         "workspace3",
						InstanceAnnKeyWorkspaceServiceMainPort: "main",
						InstanceAnnKeyURLBase:                  "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain{{NETRULE_PATH}}",
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
		want    Config
		wantErr error
	}{
		{
			name: "found",
			args: args{
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tmpl1",
						Namespace: UserNamespace("tom"),
						Labels: map[string]string{
							cosmov1alpha1.LabelKeyTemplateType: TemplateTypeWorkspace,
						},
						Annotations: map[string]string{
							InstanceAnnKeyWorkspaceDeployment:      "workspace1",
							InstanceAnnKeyWorkspaceService:         "workspace2",
							InstanceAnnKeyWorkspaceIngress:         "workspace3",
							InstanceAnnKeyWorkspaceServiceMainPort: "main",
							InstanceAnnKeyURLBase:                  "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain{{NETRULE_PATH}}",
						},
					},
				},
			},
			want: Config{
				DeploymentName:      "workspace1",
				ServiceName:         "workspace2",
				IngressName:         "workspace3",
				ServiceMainPortName: "main",
				URLBase:             "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain{{NETRULE_PATH}}",
			},
		},
		{
			name: "defaulting",
			args: args{
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tmpl1",
						Labels: map[string]string{
							cosmov1alpha1.LabelKeyTemplateType: TemplateTypeWorkspace,
						},
						Annotations: map[string]string{
							InstanceAnnKeyURLBase: "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain{{NETRULE_PATH}}",
						},
					},
				},
			},
			want: Config{
				DeploymentName:      DefaultWorkspaceResourceName,
				ServiceName:         DefaultWorkspaceResourceName,
				IngressName:         DefaultWorkspaceResourceName,
				ServiceMainPortName: DefaultWorkspaceServiceMainPortName,
				URLBase:             "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain{{NETRULE_PATH}}",
			},
		},
		{
			name: "not found ann",
			args: args{
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tmpl1",
						Labels: map[string]string{
							cosmov1alpha1.LabelKeyTemplateType: TemplateTypeWorkspace,
						},
					},
				},
			},
			wantErr: ErrNoAnnotations,
		},
		{
			name: "not found",
			args: args{
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tmpl1",
						Labels: map[string]string{
							cosmov1alpha1.LabelKeyTemplateType: TemplateTypeWorkspace,
						},
						Annotations: map[string]string{
							InstanceAnnKeyWorkspaceDeployment: "workspace1",
							InstanceAnnKeyWorkspaceService:    "workspace2",
						},
					},
				},
			},
			wantErr: ErrURLBaseNotFound,
		},
		{
			name: "no label",
			args: args{
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tmpl1",
						Namespace: UserNamespace("tom"),
						Annotations: map[string]string{
							InstanceAnnKeyWorkspaceDeployment:      "workspace1",
							InstanceAnnKeyWorkspaceService:         "workspace2",
							InstanceAnnKeyWorkspaceIngress:         "workspace3",
							InstanceAnnKeyWorkspaceServiceMainPort: "main",
							InstanceAnnKeyURLBase:                  "https://{{NETRULE_PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain{{NETRULE_PATH}}",
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
