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

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/google/go-cmp/cmp"
)

func TestPatchWorkspaceInstanceAsDesired(t *testing.T) {
	validScheme := runtime.NewScheme()
	cosmov1alpha1.AddToScheme(validScheme)
	cosmov1alpha1.AddToScheme(validScheme)
	invalidScheme := runtime.NewScheme()

	prefix := netv1.PathTypePrefix

	type args struct {
		inst   *cosmov1alpha1.Instance
		ws     cosmov1alpha1.Workspace
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
								Name:             "port1",
								PortNumber:       8080,
								HTTPPath:         "/",
								TargetPortNumber: pointer.Int32(18080),
							},
						},
					},
					Status: cosmov1alpha1.WorkspaceStatus{
						Config: cosmov1alpha1.Config{
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
						"{{USER_NAME}}":                        "default",
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
																		Name: "port18080",
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
											Name:       "port18080",
											Port:       18080,
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
						"{{USER_NAME}}":                        "default",
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
						APIVersion:         cosmov1alpha1.GroupVersion.String(),
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

func TestSvcPorts(t *testing.T) {
	netRule := func(ruleName, host, path string, portNumber, targetPortNumber int32) cosmov1alpha1.NetworkRule {
		var hostp *string
		if host != "" {
			hostp = &host
		}
		var targetp *int32
		if targetPortNumber != 0 {
			targetp = pointer.Int32(int32(targetPortNumber))
		}
		return cosmov1alpha1.NetworkRule{
			Name:             ruleName,
			PortNumber:       portNumber,
			HTTPPath:         path,
			TargetPortNumber: targetp,
			Host:             hostp,
			Group:            nil,
			Public:           false,
		}
	}

	type args struct {
		netRules []cosmov1alpha1.NetworkRule
	}
	tests := []struct {
		name string
		args args
		want []corev1.ServicePort
	}{
		{
			name: "OK1",
			args: args{
				netRules: []cosmov1alpha1.NetworkRule{netRule("rule1", "host1", "/", 1111, 2222)},
			},
			want: []corev1.ServicePort{
				{
					Name:        "port2222",
					Protocol:    "TCP",
					AppProtocol: nil,
					Port:        2222,
					TargetPort:  intstr.FromInt(2222),
					NodePort:    0,
				},
			},
		},
		{
			name: "OK2",
			args: args{
				netRules: []cosmov1alpha1.NetworkRule{
					netRule("rule1", "host1", "/", 1111, 2222),
					netRule("rule2", "host1", "/", 3333, 4444),
				},
			},
			want: []corev1.ServicePort{
				{
					Name:        "port2222",
					Protocol:    "TCP",
					AppProtocol: nil,
					Port:        2222,
					TargetPort:  intstr.FromInt(2222),
					NodePort:    0,
				},
				{
					Name:        "port4444",
					Protocol:    "TCP",
					AppProtocol: nil,
					Port:        4444,
					TargetPort:  intstr.FromInt(4444),
					NodePort:    0,
				},
			},
		},
		{
			name: "OK3",
			args: args{
				netRules: []cosmov1alpha1.NetworkRule{
					netRule("rule1", "host1", "/", 1111, 2222),
					netRule("rule2", "host1", "/", 3333, 2222),
				},
			},
			want: []corev1.ServicePort{
				{
					Name:        "port2222",
					Protocol:    "TCP",
					AppProtocol: nil,
					Port:        2222,
					TargetPort:  intstr.FromInt(2222),
					NodePort:    0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svcPorts(tt.args.netRules); !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf(cmp.Diff(got, tt.want))
			}
		})
	}
}

func TestIngressRules(t *testing.T) {

	ingPath := func(path, backendSvcName, svcBackPortName string) netv1.HTTPIngressPath {
		pathTypePrefix := netv1.PathTypePrefix
		return netv1.HTTPIngressPath{
			Path:     path,
			PathType: &pathTypePrefix,
			Backend: netv1.IngressBackend{
				Service: &netv1.IngressServiceBackend{
					Name: backendSvcName,
					Port: netv1.ServiceBackendPort{
						Name: svcBackPortName,
					},
				},
			},
		}
	}
	ingRule := func(host string, ingPathes ...netv1.HTTPIngressPath) netv1.IngressRule {
		return netv1.IngressRule{
			Host: host,
			IngressRuleValue: netv1.IngressRuleValue{
				HTTP: &netv1.HTTPIngressRuleValue{
					Paths: ingPathes,
				},
			},
		}
	}

	netRule := func(ruleName, host, path string, portNumber, targetPortNumber int32) cosmov1alpha1.NetworkRule {
		var hostp *string
		if host != "" {
			hostp = &host
		}
		var targetp *int32
		if targetPortNumber != 0 {
			targetp = pointer.Int32(int32(targetPortNumber))
		}
		return cosmov1alpha1.NetworkRule{
			Name:             ruleName,
			PortNumber:       portNumber,
			HTTPPath:         path,
			TargetPortNumber: targetp,
			Host:             hostp,
			Group:            nil,
			Public:           false,
		}
	}

	type args struct {
		netRules       []cosmov1alpha1.NetworkRule
		backendSvcName string
	}
	tests := []struct {
		name string
		args args
		want []netv1.IngressRule
	}{
		{
			name: "OK",
			args: args{
				netRules:       []cosmov1alpha1.NetworkRule{netRule("rule1", "host1", "/", 1111, 2222)},
				backendSvcName: "bksvc",
			},
			want: []netv1.IngressRule{
				ingRule("host1",
					ingPath("/", "bksvc", "port2222"),
				),
			},
		},
		{
			name: "OK2",
			args: args{
				netRules: []cosmov1alpha1.NetworkRule{
					netRule("rule1", "host1", "/", 1111, 2222),
					netRule("rule2", "host2", "/", 3333, 4444),
				},
				backendSvcName: "bksvc",
			},
			want: []netv1.IngressRule{
				ingRule("host1",
					ingPath("/", "bksvc", "port2222"),
				),
				ingRule("host2",
					ingPath("/", "bksvc", "port4444"),
				),
			},
		},
		{
			name: "OK3",
			args: args{

				netRules: []cosmov1alpha1.NetworkRule{
					netRule("rule1", "host1", "/", 1111, 2222),
					netRule("rule2", "host1", "/aaa", 3333, 4444),
				},
				backendSvcName: "bksvc",
			},
			want: []netv1.IngressRule{
				ingRule("host1",
					ingPath("/", "bksvc", "port2222"),
					ingPath("/aaa", "bksvc", "port4444"),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ingressRules(tt.args.netRules, tt.args.backendSvcName); !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf(cmp.Diff(tt.want, got))
			}
		})
	}
}
