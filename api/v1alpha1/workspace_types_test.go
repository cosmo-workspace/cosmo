package v1alpha1

import (
	"encoding/json"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

func TestNetworkRule_Default(t *testing.T) {
	tests := []struct {
		name    string
		netRule NetworkRule
		want    NetworkRule
	}{
		{
			name: "✅ TargetPortNumber is nil",
			netRule: NetworkRule{
				CustomHostPrefix: "name",
				Protocol:         "http",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: nil,
				Public:           false,
			},
			want: NetworkRule{
				CustomHostPrefix: "name",
				Protocol:         "http",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: nil,
				Public:           false,
			},
		},
		{
			name: "✅ TargetPortNumber is 0",
			netRule: NetworkRule{
				CustomHostPrefix: "name",
				Protocol:         "http",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: pointer.Int32(0),
				Public:           false,
			},
			want: NetworkRule{
				CustomHostPrefix: "name",
				Protocol:         "http",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: pointer.Int32(0),
				Public:           false,
			},
		},
		{
			name: "✅ Public is true",
			netRule: NetworkRule{
				CustomHostPrefix: "name",
				Protocol:         "http",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: pointer.Int32(2222),
				Public:           true,
			},
			want: NetworkRule{
				CustomHostPrefix: "name",
				Protocol:         "http",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: pointer.Int32(2222),
				Public:           true,
			},
		},
		{
			name: "✅ path is empty",
			netRule: NetworkRule{
				CustomHostPrefix: "name",
				Protocol:         "http",
				PortNumber:       1111,
				HTTPPath:         "",
				TargetPortNumber: pointer.Int32(2222),
				Public:           false,
			},
			want: NetworkRule{
				CustomHostPrefix: "name",
				Protocol:         "http",
				PortNumber:       1111,
				HTTPPath:         "/",
				TargetPortNumber: pointer.Int32(2222),
				Public:           false,
			},
		},
		{
			name: "✅ protocol is empty",
			netRule: NetworkRule{
				CustomHostPrefix: "port1111",
				Protocol:         "",
				PortNumber:       1111,
				HTTPPath:         "path",
				TargetPortNumber: pointer.Int32(2222),
				Public:           false,
			},
			want: NetworkRule{
				CustomHostPrefix: "port1111",
				Protocol:         "http",
				PortNumber:       1111,
				HTTPPath:         "path",
				TargetPortNumber: pointer.Int32(2222),
				Public:           false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.netRule.Default()
			got, _ := json.Marshal(tt.netRule)
			want, _ := json.Marshal(tt.want)
			if !reflect.DeepEqual(string(got), string(want)) {
				t.Errorf("NetworkRule.Default() = %v, want %v", string(got), string(want))
			}
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
		want    corev1.ServicePort
	}{
		{
			name: "✅ OK",
			netRule: &NetworkRule{
				CustomHostPrefix: "name",
				PortNumber:       1111,
				HTTPPath:         "/path",
				TargetPortNumber: pointer.Int32(2222),
			},
			want: corev1.ServicePort{
				Name:       "port1111",
				Port:       1111,
				TargetPort: intstr.FromInt(2222),
				Protocol:   corev1.ProtocolTCP,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.netRule.ServicePort()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NetworkRule.ServicePort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenURL(t *testing.T) {
	type args struct {
		protocol string
		host     string
		path     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "✅ OK",
			args: args{
				protocol: "https",
				host:     "example.com",
				path:     "/path",
			},
			want: "https://example.com/path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenURL(tt.args.protocol, tt.args.host, tt.args.path)
			if got != tt.want {
				t.Errorf("GenURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenHost(t *testing.T) {
	type args struct {
		hostbase string
		domain   string
		name     string
		ws       Workspace
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "✅ OK",
			args: args{
				hostbase: "{{NETRULE}}-{{WORKSPACE}}-{{USER}}-k3d",
				domain:   "example.com",
				name:     "name",
				ws: Workspace{
					ObjectMeta: v1.ObjectMeta{
						Name:      "ws",
						Namespace: "cosmo-user-xxx",
					},
				},
			},
			want: "name-ws-xxx-k3d.example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenHost(tt.args.hostbase, tt.args.domain, tt.args.name, tt.args.ws)
			if got != tt.want {
				t.Errorf("GenURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
