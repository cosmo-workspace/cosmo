package workspace

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/utils/pointer"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

func TestNewURLVars(t *testing.T) {
	type args struct {
		netRule cosmov1alpha1.NetworkRule
	}
	tests := []struct {
		name string
		args args
		want URLVars
	}{
		{
			name: "defaulting",
			args: args{
				netRule: cosmov1alpha1.NetworkRule{
					Name:       "name",
					PortNumber: 8080,
					Group:      pointer.String("app"),
					HTTPPath:   "/app",
				},
			},
			want: URLVars{
				NetworkRuleName: "name",
				PortNumber:      "8080",
				NetRuleGroup:    "app",
				IngressPath:     "/app",
			},
		},
		{
			name: "not defaulting",
			args: args{
				netRule: cosmov1alpha1.NetworkRule{
					Name:       "name",
					PortNumber: 8080,
					HTTPPath:   "/app",
					Group:      pointer.String("app"),
				},
			},
			want: URLVars{
				NetworkRuleName: "name",
				PortNumber:      "8080",
				IngressPath:     "/app",
				NetRuleGroup:    "app",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewURLVars(tt.args.netRule); !equality.Semantic.DeepEqual(got, tt.want) {
				t.Errorf("NewURLVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLBase_GenURL(t *testing.T) {
	type fields struct {
		Base string
		Vars URLVars
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "OK1",
			fields: fields{
				Base: "http://localhost:{{PORT_NUMBER}}",
				Vars: URLVars{
					PortNumber:  "8080",
					IngressPath: "/app",
				},
			},
			want: "http://localhost:8080/app",
		},
		{
			name: "OK2",
			fields: fields{
				Base: "https://{{NETRULE_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
				Vars: URLVars{
					NetworkRuleName: "main",
					IngressPath:     "/",
					InstanceName:    "inst",
					Namespace:       "ns",
				},
			},
			want: "https://main-inst-ns.domain/",
		},
		{
			name: "OK3",
			fields: fields{
				Base: "https://{{PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}.domain",
				Vars: URLVars{
					NetworkRuleName: "main",
					IngressPath:     "/",
					InstanceName:    "inst",
					Namespace:       "ns",
				},
			},
			want: "https://main-inst-ns.domain/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := URLBase(tt.fields.Base)
			if got := u.GenURL(tt.fields.Vars); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("URLBoilerPlate.GenURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateIngressHost(t *testing.T) {
	type args struct {
		r         cosmov1alpha1.NetworkRule
		name      string
		namespace string
		urlBase   URLBase
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "NETRULE_NAME",
			args: args{
				r: cosmov1alpha1.NetworkRule{
					Name:             "http",
					PortNumber:       3000,
					HTTPPath:         "/",
					TargetPortNumber: pointer.Int32(3001),
					Group:            pointer.String("nodejs"),
					Public:           false,
				},
				name:      "cs1",
				namespace: cosmov1alpha1.UserNamespace("tom"),
				urlBase:   URLBase("https://{{NETRULE_NAME}}-{{INSTANCE}}-{{NAMESPACE}}"),
			},
			want: "http-cs1-cosmo-user-tom",
		},
		{
			name: "PORT_NAME",
			args: args{
				r: cosmov1alpha1.NetworkRule{
					Name:             "http",
					PortNumber:       3000,
					HTTPPath:         "/",
					TargetPortNumber: pointer.Int32(3001),
					Group:            pointer.String("nodejs"),
					Public:           false,
				},
				name:      "cs1",
				namespace: cosmov1alpha1.UserNamespace("tom"),
				urlBase:   URLBase("https://{{PORT_NAME}}-{{INSTANCE}}-{{NAMESPACE}}"),
			},
			want: "http-cs1-cosmo-user-tom",
		},
		{
			name: "NETRULE_GROUP",
			args: args{
				r: cosmov1alpha1.NetworkRule{
					Name:             "nodejs",
					PortNumber:       3000,
					HTTPPath:         "/",
					TargetPortNumber: pointer.Int32(3002),
					Group:            pointer.String("myapp"),
					Public:           false,
				},
				name:      "cs1",
				namespace: cosmov1alpha1.UserNamespace("tom"),
				urlBase:   URLBase("https://{{NETRULE_GROUP}}-{{WORKSPACE}}-{{USER_NAME}}"),
			},
			want: "myapp-cs1-tom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateIngressHost(tt.args.r, tt.args.name, tt.args.namespace, tt.args.urlBase); got != tt.want {
				t.Errorf("GenerateIngressHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractHost(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with proto, port and path",
			args: args{
				url: "http://localhost:8080/hello",
			},
			want: "localhost",
		},
		{
			name: "with proto and path",
			args: args{
				url: "https://cosmo-workspace.github.io/hello",
			},
			want: "cosmo-workspace.github.io",
		},
		{
			name: "with proto and port",
			args: args{
				url: "https://cosmo-workspace.github.io:8080",
			},
			want: "cosmo-workspace.github.io",
		},
		{
			name: "with nothing",
			args: args{
				url: "cosmo-workspace.github.io",
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractHost(tt.args.url); got != tt.want {
				t.Errorf("extractHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURLVars_setDefault(t *testing.T) {
	type fields struct {
		URLVars
	}
	tests := []struct {
		name   string
		fields fields
		want   URLVars
	}{
		{
			name:   "All",
			fields: fields{},
			want: URLVars{
				NetworkRuleName: "undefined",
				PortNumber:      "0",
				NetRuleGroup:    "undefined",
				IngressPath:     "/",
				InstanceName:    "undefined",
				Namespace:       "undefined",
				NodePortNumber:  "0",
				LoadBalancer:    "undefined",
				WorkspaceName:   "undefined",
				UserName:        "undefined",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &URLVars{
				NetworkRuleName: tt.fields.NetworkRuleName,
				PortNumber:      tt.fields.PortNumber,
				NetRuleGroup:    tt.fields.NetRuleGroup,
				IngressPath:     tt.fields.IngressPath,
				InstanceName:    tt.fields.InstanceName,
				Namespace:       tt.fields.Namespace,
			}
			v.setDefault()

			if !reflect.DeepEqual(*v, tt.want) {
				t.Errorf("setDefault() = %v, want %v", *v, tt.want)
			}
		})
	}
}
