package workspace

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/google/go-cmp/cmp"
)

func TestPatchWorkspaceInstanceAsDesired(t *testing.T) {
	validScheme := runtime.NewScheme()
	cosmov1alpha1.AddToScheme(validScheme)
	wsv1alpha1.AddToScheme(validScheme)
	invalidScheme := runtime.NewScheme()

	prefix := netv1.PathTypePrefix

	type args struct {
		inst   *cosmov1alpha1.Instance
		ws     wsv1alpha1.Workspace
		scheme *runtime.Scheme
	}
	tests := []struct {
		name         string
		args         args
		want         *cosmov1alpha1.Instance
		wantErr      bool
		wantOwnerref bool
	}{
		{
			name: "OK",
			args: args{
				ws: wsv1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ws1",
						Namespace: "cosmo-user-default",
					},
					Spec: wsv1alpha1.WorkspaceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "tmpl1",
						},
						Replicas: pointer.Int64(1),
						Vars: map[string]string{
							"VAR1": "VAL1",
						},
						Network: []wsv1alpha1.NetworkRule{
							{
								Name:             "port1",
								PortNumber:       8080,
								HTTPPath:         "/",
								TargetPortNumber: pointer.Int32(18080),
							},
						},
					},
					Status: wsv1alpha1.WorkspaceStatus{
						Config: wsv1alpha1.Config{
							DeploymentName:      "ws-deploy",
							IngressName:         "ws-ing",
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
			want: &cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "inst1",
					Namespace: "cosmo-user-default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: "tmpl1",
					},
					Vars: map[string]string{
						"VAR1":                                 "VAL1",
						"{{WORKSPACE}}":                        "ws1",
						"{{USERID}}":                           "default",
						"{{WORKSPACE_DEPLOYMENT_NAME}}":        "ws-deploy",
						"{{WORKSPACE_INGRESS_NAME}}":           "ws-ing",
						"{{WORKSPACE_SERVICE_NAME}}":           "ws-svc",
						"{{WORKSPACE_SERVICE_MAIN_PORT_NAME}}": "main",
					},
					Override: cosmov1alpha1.OverrideSpec{
						Scale: []cosmov1alpha1.ScalingOverrideSpec{
							{
								Target: cosmov1alpha1.ObjectRef{
									ObjectReference: corev1.ObjectReference{
										APIVersion: "apps/v1",
										Kind:       "Deployment",
										Name:       "ws-deploy",
									},
								},
								Replicas: 1,
							},
						},
						Network: &cosmov1alpha1.NetworkOverrideSpec{
							Ingress: []cosmov1alpha1.IngressOverrideSpec{
								{
									TargetName: "ws-ing",
									Rules: []netv1.IngressRule{
										{
											IngressRuleValue: netv1.IngressRuleValue{
												HTTP: &netv1.HTTPIngressRuleValue{
													Paths: []netv1.HTTPIngressPath{
														{
															Path:     "/",
															PathType: &prefix,
															Backend: netv1.IngressBackend{
																Service: &netv1.IngressServiceBackend{
																	Name: "ws1-ws-svc",
																	Port: netv1.ServiceBackendPort{
																		Name: "port1",
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
							Service: []cosmov1alpha1.ServiceOverrideSpec{
								{
									TargetName: "ws-svc",
									Ports: []corev1.ServicePort{
										{
											Name:       "port1",
											Port:       8080,
											TargetPort: intstr.FromInt(18080),
											Protocol:   "TCP",
										},
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "OK with scheme",
			args: args{
				ws: wsv1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ws1",
						Namespace: "cosmo-user-default",
					},
					Spec: wsv1alpha1.WorkspaceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "tmpl1",
						},
						Replicas: pointer.Int64(0),
					},
					Status: wsv1alpha1.WorkspaceStatus{
						Config: wsv1alpha1.Config{
							DeploymentName:      "ws-deploy",
							IngressName:         "ws-ing",
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
			want: &cosmov1alpha1.Instance{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "inst1",
					Namespace: "cosmo-user-default",
				},
				Spec: cosmov1alpha1.InstanceSpec{
					Template: cosmov1alpha1.TemplateRef{
						Name: "tmpl1",
					},
					Vars: map[string]string{
						"{{WORKSPACE}}":                        "ws1",
						"{{USERID}}":                           "default",
						"{{WORKSPACE_DEPLOYMENT_NAME}}":        "ws-deploy",
						"{{WORKSPACE_INGRESS_NAME}}":           "ws-ing",
						"{{WORKSPACE_SERVICE_NAME}}":           "ws-svc",
						"{{WORKSPACE_SERVICE_MAIN_PORT_NAME}}": "main",
					},
					Override: cosmov1alpha1.OverrideSpec{
						Scale: []cosmov1alpha1.ScalingOverrideSpec{
							{
								Target: cosmov1alpha1.ObjectRef{
									ObjectReference: corev1.ObjectReference{
										APIVersion: "apps/v1",
										Kind:       "Deployment",
										Name:       "ws-deploy",
									},
								},
								Replicas: 0,
							},
						},
						Network: &cosmov1alpha1.NetworkOverrideSpec{
							Ingress: []cosmov1alpha1.IngressOverrideSpec{
								{TargetName: "ws-ing", Rules: []netv1.IngressRule{}},
							},
							Service: []cosmov1alpha1.ServiceOverrideSpec{
								{TargetName: "ws-svc", Ports: []corev1.ServicePort{}},
							},
						},
					},
				},
			},
			wantErr:      false,
			wantOwnerref: true,
		},
		{
			name: "Err witr invalid scheme",
			args: args{
				ws: wsv1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ws1",
						Namespace: "cosmo-user-default",
					},
					Spec: wsv1alpha1.WorkspaceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "tmpl1",
						},
						Replicas: pointer.Int64(0),
					},
					Status: wsv1alpha1.WorkspaceStatus{
						Config: wsv1alpha1.Config{
							DeploymentName:      "ws-deploy",
							IngressName:         "ws-ing",
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
			wantErr:      true,
			wantOwnerref: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PatchWorkspaceInstanceAsDesired(tt.args.inst, tt.args.ws, tt.args.scheme)
			if (err != nil) != tt.wantErr {
				t.Errorf("PatchWorkspaceInstanceAsDesired() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				ownerRef := tt.args.inst.GetOwnerReferences()
				tt.args.inst.SetOwnerReferences(nil)

				if !equality.Semantic.DeepEqual(tt.args.inst, tt.want) {
					t.Errorf("PatchWorkspaceInstanceAsDesired() got = %v, want %v, diff = %s", tt.args.inst, tt.want, cmp.Diff(tt.args.inst, tt.want))
				}

				if (ownerRef != nil) != tt.wantOwnerref {
					t.Errorf("PatchWorkspaceInstanceAsDesired() ownerRef = %v, wantOwnerref %v", ownerRef, tt.wantOwnerref)
				}
				if len(ownerRef) > 0 {
					if len(ownerRef) != 1 {
						t.Errorf("PatchWorkspaceInstanceAsDesired() ownerRef should be 1 but %v", len(ownerRef))
					}
					expectedRef := metav1.OwnerReference{
						APIVersion:         wsv1alpha1.GroupVersion.String(),
						Kind:               "Workspace",
						Name:               tt.args.ws.GetName(),
						UID:                tt.args.ws.GetUID(),
						BlockOwnerDeletion: pointer.BoolPtr(true),
						Controller:         pointer.BoolPtr(true),
					}
					if !equality.Semantic.DeepEqual(ownerRef[0], expectedRef) {
						t.Errorf("PatchWorkspaceInstanceAsDesired() owner ref = %v, want %v", ownerRef[0], expectedRef)
					}
				}
			}
		})
	}
}
