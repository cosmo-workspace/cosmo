package v1alpha1

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

func TestNetworkRule_Default(t *testing.T) {
	tests := []struct {
		name    string
		netRule *NetworkRule
	}{
		{
			name: "✅ TargetPortNumber is nil",
			netRule: &NetworkRule{
				Name:             "name",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: nil,
				Host:             pointer.String("host"),
				Group:            pointer.String("group"),
				Public:           false,
			},
		},
		{
			name: "✅ TargetPortNumber is 0",
			netRule: &NetworkRule{
				Name:             "name",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: pointer.Int32(0),
				Host:             pointer.String("host"),
				Group:            pointer.String("group"),
				Public:           false,
			},
		},
		{
			name: "✅ Public is true",
			netRule: &NetworkRule{
				Name:             "name",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: pointer.Int32(2222),
				Host:             pointer.String("host"),
				Group:            pointer.String("group"),
				Public:           true,
			},
		},
		{
			name: "✅ path is empty",
			netRule: &NetworkRule{
				Name:             "name",
				PortNumber:       1111,
				HTTPPath:         "",
				TargetPortNumber: pointer.Int32(2222),
				Host:             pointer.String("host"),
				Group:            pointer.String("group"),
				Public:           false,
			},
		},
		{
			name: "✅ group is nil",
			netRule: &NetworkRule{
				Name:             "name",
				PortNumber:       1111,
				HTTPPath:         "path",
				TargetPortNumber: pointer.Int32(2222),
				Host:             pointer.String("host"),
				Group:            nil,
				Public:           false,
			},
		},
		{
			name: "✅ group is empty",
			netRule: &NetworkRule{
				Name:             "name",
				PortNumber:       1111,
				HTTPPath:         "path",
				TargetPortNumber: pointer.Int32(2222),
				Host:             pointer.String("host"),
				Group:            pointer.String(""),
				Public:           false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.netRule.Default()
			snaps.MatchSnapshot(t, tt.netRule)
		})
	}
}

func TestNetworkRule_portName(t *testing.T) {
	tests := []struct {
		name    string
		netRule *NetworkRule
		want    string
	}{
		{
			name: "✅ OK",
			netRule: &NetworkRule{
				PortNumber:       1111,
				TargetPortNumber: pointer.Int32(2222),
			},
			want: "port1111",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.netRule.portName(); got != tt.want {
				t.Errorf("NetworkRule.portName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNetworkRule_ServicePort(t *testing.T) {
	tests := []struct {
		name    string
		netRule *NetworkRule
	}{
		{
			name: "✅ OK",
			netRule: &NetworkRule{
				Name:             "name",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: pointer.Int32(2222),
				Host:             pointer.String("host"),
				Group:            pointer.String("group"),
				Public:           true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.netRule.ServicePort()
			snaps.MatchJSON(t, got)
		})
	}
}

func TestNetworkRule_IngressRule(t *testing.T) {
	type args struct {
		backendSvcName string
	}
	tests := []struct {
		name    string
		netRule *NetworkRule
		args    args
	}{
		{
			name: "✅ OK",
			netRule: &NetworkRule{
				Name:             "name",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: pointer.Int32(2222),
				Host:             pointer.String("host"),
				Group:            pointer.String("group"),
				Public:           true,
			},
			args: args{
				backendSvcName: "svcname",
			},
		},
		{
			name: "✅ host is nil",
			netRule: &NetworkRule{
				Name:             "name",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: pointer.Int32(2222),
				Host:             nil,
				Group:            pointer.String("group"),
				Public:           true,
			},
			args: args{
				backendSvcName: "svcname",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.netRule.IngressRule(tt.args.backendSvcName)
			snaps.MatchJSON(t, got)
		})
	}
}

func TestNetworkRule_TraefikRoute(t *testing.T) {
	type args struct {
		backendSvcName       string
		headerMiddlewareName string
	}
	tests := []struct {
		name    string
		netRule *NetworkRule
		args    args
	}{
		{
			name: "✅ public",
			netRule: &NetworkRule{
				Name:             "name",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: pointer.Int32(2222),
				Host:             pointer.String("host"),
				Group:            pointer.String("group"),
				Public:           true,
			},
			args: args{
				backendSvcName: "svcname",
			},
		},
		{
			name: "✅ not public",
			netRule: &NetworkRule{
				Name:             "name",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: pointer.Int32(2222),
				Host:             pointer.String("host"),
				Group:            pointer.String("group"),
				Public:           false,
			},
			args: args{
				backendSvcName:       "svcname",
				headerMiddlewareName: "headers",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.netRule.TraefikRoute(tt.args.backendSvcName, tt.args.headerMiddlewareName)
			snaps.MatchJSON(t, got)
		})
	}
}

func TestNetworkRulesByServiceAndIngress(t *testing.T) {
	pathTypePrefix := netv1.PathTypePrefix
	type args struct {
		svc corev1.Service
		ing netv1.Ingress
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "✅ OK",
			args: args{
				svc: corev1.Service{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-svc",
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name:       "main",
								Port:       int32(7777),
								Protocol:   corev1.ProtocolTCP,
								TargetPort: intstr.FromInt(32000),
							},
							{
								Name:       "main2",
								Port:       int32(7778),
								Protocol:   corev1.ProtocolTCP,
								TargetPort: intstr.FromInt(32001),
							},
						},
					},
				},
				ing: netv1.Ingress{
					Spec: netv1.IngressSpec{
						Rules: []netv1.IngressRule{
							{
								Host: "host.example.com",
								IngressRuleValue: netv1.IngressRuleValue{
									HTTP: &netv1.HTTPIngressRuleValue{
										Paths: []netv1.HTTPIngressPath{
											{
												Path:     "/",
												PathType: &pathTypePrefix,
												Backend: netv1.IngressBackend{
													Service: &netv1.IngressServiceBackend{
														Name: "test-svc",
														Port: netv1.ServiceBackendPort{
															Name: "main",
														},
													},
												},
											},
											{
												Path:     "/aaa",
												PathType: &pathTypePrefix,
												Backend: netv1.IngressBackend{
													Service: &netv1.IngressServiceBackend{
														Name: "test-svc",
														Port: netv1.ServiceBackendPort{
															Number: 7778,
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
			},
		},
		{
			name: "✅ no ingress",
			args: args{
				svc: corev1.Service{
					ObjectMeta: v1.ObjectMeta{
						Name: "test-svc",
					},
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name:       "main",
								Port:       int32(7777),
								Protocol:   corev1.ProtocolTCP,
								TargetPort: intstr.FromInt(32000),
							},
						},
					},
				},
				ing: netv1.Ingress{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NetworkRulesByServiceAndIngress(tt.args.svc, tt.args.ing)
			snaps.MatchSnapshot(t, got)
		})
	}
}
