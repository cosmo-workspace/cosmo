package v1alpha1

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	corev1 "k8s.io/api/core/v1"
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

func TestNetworkRulesByService(t *testing.T) {
	type args struct {
		svc corev1.Service
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NetworkRulesByService(tt.args.svc)
			snaps.MatchSnapshot(t, got)
		})
	}
}
