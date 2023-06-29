package workspace

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/pointer"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

func TestPatchWorkspaceInstanceAsDesired(t *testing.T) {
	validScheme := runtime.NewScheme()
	utilruntime.Must(cosmov1alpha1.AddToScheme(validScheme))
	invalidScheme := runtime.NewScheme()

	type args struct {
		inst   *cosmov1alpha1.Instance
		ws     cosmov1alpha1.Workspace
		scheme *runtime.Scheme
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "✅ OK",
			args: args{
				ws: cosmov1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ws1",
						Namespace: "cosmo-user-default",
					},
					Spec: cosmov1alpha1.WorkspaceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "tmpl1",
						},
						Replicas: pointer.Int64(1),
						Vars: map[string]string{
							"VAR1": "VAL1",
						},
						Network: []cosmov1alpha1.NetworkRule{
							{
								PortNumber:       8080,
								HTTPPath:         "/",
								TargetPortNumber: pointer.Int32(18080),
							},
							{
								PortNumber:       9999,
								HTTPPath:         "/",
								TargetPortNumber: pointer.Int32(19999),
							},
						},
					},
					Status: cosmov1alpha1.WorkspaceStatus{
						Config: cosmov1alpha1.Config{
							DeploymentName:      "ws-deploy",
							ServiceName:         "ws-svc",
							ServiceMainPortName: "main",
						},
					},
				},
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inst1",
						Namespace: "cosmo-user-default",
					},
				},
				scheme: nil,
			},
		},
		{
			name: "✅ OK with scheme",
			args: args{
				ws: cosmov1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ws1",
						Namespace: "cosmo-user-default",
					},
					Spec: cosmov1alpha1.WorkspaceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "tmpl1",
						},
						Replicas: pointer.Int64(0),
					},
					Status: cosmov1alpha1.WorkspaceStatus{
						Config: cosmov1alpha1.Config{
							DeploymentName:      "ws-deploy",
							ServiceName:         "ws-svc",
							ServiceMainPortName: "main",
						},
					},
				},
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inst1",
						Namespace: "cosmo-user-default",
					},
				},
				scheme: validScheme,
			},
		},
		{
			name: "❌ Err witr invalid scheme",
			args: args{
				ws: cosmov1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ws1",
						Namespace: "cosmo-user-default",
					},
					Spec: cosmov1alpha1.WorkspaceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "tmpl1",
						},
						Replicas: pointer.Int64(0),
					},
					Status: cosmov1alpha1.WorkspaceStatus{
						Config: cosmov1alpha1.Config{
							DeploymentName:      "ws-deploy",
							ServiceName:         "ws-svc",
							ServiceMainPortName: "main",
						},
					},
				},
				inst: &cosmov1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "inst1",
						Namespace: "cosmo-user-default",
					},
				},
				scheme: invalidScheme,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PatchWorkspaceInstanceAsDesired(tt.args.inst, tt.args.ws, tt.args.scheme)
			if err != nil {
				snaps.MatchSnapshot(t, err.Error())
			} else {
				snaps.MatchJSON(t, tt.args.inst)
			}
		})
	}
}

func TestSvcPorts(t *testing.T) {
	netRule := func(ruleName, host, path string, portNumber, targetPortNumber int32) cosmov1alpha1.NetworkRule {
		var targetp *int32
		if targetPortNumber != 0 {
			targetp = pointer.Int32(int32(targetPortNumber))
		}
		return cosmov1alpha1.NetworkRule{
			PortNumber:       portNumber,
			HTTPPath:         path,
			TargetPortNumber: targetp,
			Public:           false,
		}
	}

	type args struct {
		netRules []cosmov1alpha1.NetworkRule
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "✅ OK1",
			args: args{
				netRules: []cosmov1alpha1.NetworkRule{netRule("rule1", "host1", "/", 1111, 2222)},
			},
		},
		{
			name: "✅ OK2",
			args: args{
				netRules: []cosmov1alpha1.NetworkRule{
					netRule("rule1", "host1", "/", 1111, 2222),
					netRule("rule2", "host1", "/", 3333, 4444),
				},
			},
		},
		{
			name: "✅ OK3",
			args: args{
				netRules: []cosmov1alpha1.NetworkRule{
					netRule("rule1", "host1", "/", 1111, 2222),
					netRule("rule2", "host1", "/", 3333, 2222),
				},
			},
		},
		{
			name: "✅ OK3",
			args: args{
				netRules: []cosmov1alpha1.NetworkRule{
					netRule("rule1", "host1", "/", 1111, 2222),
					netRule("rule2", "host1", "/", 3333, 2222),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svcPorts(tt.args.netRules)
			snaps.MatchJSON(t, got)
		})
	}
}
