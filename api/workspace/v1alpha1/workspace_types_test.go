package v1alpha1

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

func TestNetworkRulesByServiceAndIngress(t *testing.T) {
	pathTypePrefix := netv1.PathTypePrefix
	type args struct {
		svc corev1.Service
		ing netv1.Ingress
	}
	tests := []struct {
		name string
		args args
		want []NetworkRule
	}{
		{
			name: "OK",
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
			want: []NetworkRule{
				{
					PortName:         "main",
					PortNumber:       7777,
					TargetPortNumber: pointer.Int32(32000),
					HTTPPath:         "/",
					Host:             pointer.String("host.example.com"),
					Group:            pointer.String("main"),
					Public:           false,
				},
				{
					PortName:         "main2",
					PortNumber:       7778,
					TargetPortNumber: pointer.Int32(32001),
					HTTPPath:         "/aaa",
					Host:             pointer.String("host.example.com"),
					Group:            pointer.String("main2"),
					Public:           false,
				},
			},
		},
		{
			name: "no ingress",
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
			want: []NetworkRule{
				{
					PortName:         "main",
					PortNumber:       7777,
					TargetPortNumber: pointer.Int32(32000),
					HTTPPath:         "/",
					Host:             nil,
					Group:            pointer.String("main"),
					Public:           false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NetworkRulesByServiceAndIngress(tt.args.svc, tt.args.ing); !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("NetworkRulesByServiceAndIngress() = %v, want %v", got, tt.want)
			}
		})
	}
}
