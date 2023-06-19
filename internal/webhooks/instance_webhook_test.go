package webhooks

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

func Test_mutateInstanceObject(t *testing.T) {
	type args struct {
		inst cosmov1alpha1.InstanceObject
		tmpl cosmov1alpha1.TemplateObject
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "mutate workspace instance",
			args: args{
				tmpl: &cosmov1alpha1.Template{
					ObjectMeta: metav1.ObjectMeta{
						Name: "workspace-template",
						Labels: map[string]string{
							cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeWorkspace,
						},
					},
					Spec: cosmov1alpha1.TemplateSpec{
						RequiredVars: []cosmov1alpha1.RequiredVarSpec{
							{
								Var:     "XXX",
								Default: "xxx",
							},
						},
					},
				},
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name: "workspace-instance",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "workspace-template",
						},
						Override: cosmov1alpha1.OverrideSpec{
							PatchesJson6902: []cosmov1alpha1.Json6902{
								{
									Target: cosmov1alpha1.ObjectRef{
										ObjectReference: corev1.ObjectReference{
											APIVersion: "apps/v1", Kind: "Deployment", Name: "deployment",
										},
									},
								},
								{
									Target: cosmov1alpha1.ObjectRef{
										ObjectReference: corev1.ObjectReference{
											APIVersion: "v1", Kind: "Service", Name: "service",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "mutate useraddon clusterinstance",
			args: args{
				tmpl: &cosmov1alpha1.ClusterTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "workspace-template",
						Labels: map[string]string{
							cosmov1alpha1.TemplateLabelKeyType: cosmov1alpha1.TemplateLabelEnumTypeUserAddon,
						},
					},
					Spec: cosmov1alpha1.TemplateSpec{
						RequiredVars: []cosmov1alpha1.RequiredVarSpec{
							{
								Var:     "XXX",
								Default: "xxx",
							},
							{
								Var:     "YYY",
								Default: "yyy",
							},
						},
					},
				},
				inst: &cosmov1alpha1.ClusterInstance{
					ObjectMeta: metav1.ObjectMeta{
						Name: "workspace-instance",
					},
					Spec: cosmov1alpha1.InstanceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "workspace-template",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mutateInstanceObject(tt.args.inst, tt.args.tmpl)
			snaps.MatchJSON(t, tt.args.inst)
		})
	}
}
