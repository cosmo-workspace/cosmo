package workspace

import (
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/gkampitakis/go-snaps/snaps"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetConfigOnTemplateAnnotations(t *testing.T) {
	type args struct {
		cfg cosmov1alpha1.Config
		obj *cosmov1alpha1.Template
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "OK",
			args: args{
				cfg: cosmov1alpha1.Config{
					DeploymentName:      "workspace1",
					ServiceName:         "workspace2",
					ServiceMainPortName: "main",
				},
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tmpl",
						Annotations: map[string]string{
							cosmov1alpha1.WorkspaceTemplateAnnKeyServiceName: "workspace",
						},
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
					ServiceMainPortName: "main",
				},
				obj: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "tmpl",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetConfigOnTemplateAnnotations(tt.args.obj, tt.args.cfg)
			snaps.MatchJSON(t, tt.args.obj)
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
							cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort: "main",
						},
					},
				},
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
							cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort: "main",
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
							cosmov1alpha1.WorkspaceTemplateAnnKeyServiceMainPort: "main",
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConfigFromTemplateAnnotations(tt.args.obj)
			if err != tt.wantErr {
				t.Errorf("ConfigFromTemplateAnnotations() gotErr = %v, wantErr %v", err, tt.wantErr)
			}
			snaps.MatchJSON(t, got)
		})
	}
}
