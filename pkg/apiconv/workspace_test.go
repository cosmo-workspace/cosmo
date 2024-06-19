package apiconv

import (
	"reflect"
	"slices"
	"testing"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestC2D_Workspaces(t *testing.T) {
	type args struct {
		wss  []cosmov1alpha1.Workspace
		opts []WorkspaceConvertOptions
	}
	tests := []struct {
		name string
		args args
		want []*dashv1alpha1.Workspace
	}{
		{
			name: "OK",
			args: args{
				wss: []cosmov1alpha1.Workspace{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "ws1",
						},
						Spec: cosmov1alpha1.WorkspaceSpec{
							Template: cosmov1alpha1.TemplateRef{
								Name: "tmpl1",
							},
							Replicas: ptr.To(int64(1)),
							Vars: map[string]string{
								"key1": "val1",
								"key2": "val2",
							},
							Network: []cosmov1alpha1.NetworkRule{
								{
									Protocol:         "http",
									PortNumber:       8080,
									CustomHostPrefix: "xxx",
									HTTPPath:         "/path",
									Public:           true,
								},
								{
									Protocol:   "http",
									PortNumber: 8443,
									HTTPPath:   "/",
								},
							},
						},
						Status: cosmov1alpha1.WorkspaceStatus{
							Phase: "Running",
							Config: cosmov1alpha1.Config{
								ServiceMainPortName: "main",
							},
							URLs: map[string]string{
								"http://main/":     "https://main.example.com",
								"http://xxx/path":  "https://xxx.example.com/path",
								"http://port8443/": "https://port8443.example.com/",
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "ws2",
						},
						Spec: cosmov1alpha1.WorkspaceSpec{
							Template: cosmov1alpha1.TemplateRef{
								Name: "tmpl2",
							},
						},
					},
				},
			},
			want: []*dashv1alpha1.Workspace{
				{
					Name: "ws1",
					Spec: &dashv1alpha1.WorkspaceSpec{
						Template: "tmpl1",
						Replicas: int64(1),
						Vars: map[string]string{
							"key1": "val1",
							"key2": "val2",
						},
						Network: []*dashv1alpha1.NetworkRule{
							{
								PortNumber:       8080,
								CustomHostPrefix: "xxx",
								HttpPath:         "/path",
								Public:           true,
								Url:              "https://xxx.example.com/path",
							},
							{
								PortNumber: 8443,
								HttpPath:   "/",
								Url:        "https://port8443.example.com/",
							},
						},
					},
					Status: &dashv1alpha1.WorkspaceStatus{
						MainUrl: "https://main.example.com",
						Phase:   "Running",
					},
				},
				{
					Name: "ws2",
					Spec: &dashv1alpha1.WorkspaceSpec{
						Template: "tmpl2",
						Replicas: 1,
					},
					Status: &dashv1alpha1.WorkspaceStatus{},
				},
			},
		},
		{
			name: "empty",
			args: args{
				wss: []cosmov1alpha1.Workspace{},
			},
			want: []*dashv1alpha1.Workspace{},
		},
		{
			name: "nil",
			args: args{
				wss: nil,
			},
			want: []*dashv1alpha1.Workspace{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := C2D_Workspaces(tt.args.wss, tt.args.opts...)
			newGot := make([]string, len(got))
			for _, v := range got {
				newGot = append(newGot, v.String())
			}
			want := make([]string, len(tt.want))
			for _, v := range tt.want {
				want = append(want, v.String())
			}
			if !slices.Equal(want, newGot) {
				t.Errorf("C2D_Workspaces() = %v, want %v\ndiff = %v", newGot, want, cmp.Diff(want, newGot))
			}
		})
	}
}

func TestC2D_Workspace(t *testing.T) {
	type args struct {
		ws   cosmov1alpha1.Workspace
		opts []WorkspaceConvertOptions
	}
	tests := []struct {
		name string
		args args
		want *dashv1alpha1.Workspace
	}{
		{
			name: "OK",
			args: args{
				ws: cosmov1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ws1",
					},
					Spec: cosmov1alpha1.WorkspaceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "tmpl1",
						},
						Replicas: ptr.To(int64(1)),
						Vars: map[string]string{
							"key1": "val1",
							"key2": "val2",
						},
						Network: []cosmov1alpha1.NetworkRule{
							{
								Protocol:         "http",
								PortNumber:       8080,
								CustomHostPrefix: "xxx",
								HTTPPath:         "/path",
								Public:           true,
							},
							{
								Protocol:     "http",
								PortNumber:   8443,
								HTTPPath:     "/",
								AllowedUsers: []string{"share1"},
							},
						},
					},
					Status: cosmov1alpha1.WorkspaceStatus{
						Phase: "Running",
						Config: cosmov1alpha1.Config{
							ServiceMainPortName: "main",
						},
						URLs: map[string]string{
							"http://main/":     "https://main.example.com",
							"http://xxx/path":  "https://xxx.example.com/path",
							"http://port8443/": "https://port8443.example.com/",
						},
					},
				},
			},
			want: &dashv1alpha1.Workspace{
				Name: "ws1",
				Spec: &dashv1alpha1.WorkspaceSpec{
					Template: "tmpl1",
					Replicas: int64(1),
					Vars: map[string]string{
						"key1": "val1",
						"key2": "val2",
					},
					Network: []*dashv1alpha1.NetworkRule{
						{
							PortNumber:       8080,
							CustomHostPrefix: "xxx",
							HttpPath:         "/path",
							Public:           true,
							Url:              "https://xxx.example.com/path",
						},
						{
							PortNumber:   8443,
							HttpPath:     "/",
							Url:          "https://port8443.example.com/",
							AllowedUsers: []string{"share1"},
						},
					},
				},
				Status: &dashv1alpha1.WorkspaceStatus{
					MainUrl: "https://main.example.com",
					Phase:   "Running",
				},
			},
		},
		{
			name: "WithRaw",
			args: args{
				ws: cosmov1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ws1",
					},
					Spec: cosmov1alpha1.WorkspaceSpec{
						Template: cosmov1alpha1.TemplateRef{
							Name: "tmpl1",
						},
						Replicas: ptr.To(int64(1)),
						Vars: map[string]string{
							"key1": "val1",
							"key2": "val2",
						},
						Network: []cosmov1alpha1.NetworkRule{
							{
								Protocol:         "http",
								PortNumber:       8080,
								CustomHostPrefix: "xxx",
								HTTPPath:         "/path",
								Public:           true,
							},
							{
								Protocol:   "http",
								PortNumber: 8443,
								HTTPPath:   "/",
							},
						},
					},
					Status: cosmov1alpha1.WorkspaceStatus{
						Phase: "Running",
						Config: cosmov1alpha1.Config{
							ServiceMainPortName: "main",
						},
						URLs: map[string]string{
							"http://main/":     "https://main.example.com",
							"http://xxx/path":  "https://xxx.example.com/path",
							"http://port8443/": "https://port8443.example.com/",
						},
					},
				},
				opts: []WorkspaceConvertOptions{
					WithWorkspaceRaw(ptr.To(true)),
				},
			},
			want: &dashv1alpha1.Workspace{
				Name: "ws1",
				Spec: &dashv1alpha1.WorkspaceSpec{
					Template: "tmpl1",
					Replicas: int64(1),
					Vars: map[string]string{
						"key1": "val1",
						"key2": "val2",
					},
					Network: []*dashv1alpha1.NetworkRule{
						{
							PortNumber:       8080,
							CustomHostPrefix: "xxx",
							HttpPath:         "/path",
							Public:           true,
							Url:              "https://xxx.example.com/path",
						},
						{
							PortNumber: 8443,
							HttpPath:   "/",
							Url:        "https://port8443.example.com/",
						},
					},
				},
				Status: &dashv1alpha1.WorkspaceStatus{
					MainUrl: "https://main.example.com",
					Phase:   "Running",
				},
				Raw: ptr.To(`apiVersion: cosmo-workspace.github.io/v1alpha1
kind: Workspace
metadata:
  creationTimestamp: null
  name: ws1
spec:
  network:
  - customHostPrefix: xxx
    httpPath: /path
    portNumber: 8080
    protocol: http
    public: true
  - httpPath: /
    portNumber: 8443
    protocol: http
    public: false
  replicas: 1
  template:
    name: tmpl1
  vars:
    key1: val1
    key2: val2
status:
  config:
    mainServicePortName: main
  instance: {}
  phase: Running
  urls:
    http://main/: https://main.example.com
    http://port8443/: https://port8443.example.com/
    http://xxx/path: https://xxx.example.com/path
`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := C2D_Workspace(tt.args.ws, tt.args.opts...); !reflect.DeepEqual(got.String(), tt.want.String()) {
				t.Errorf("C2D_Workspace() raw diff %v", cmp.Diff(*got.Raw, *tt.want.Raw))
				t.Errorf("C2D_Workspace() obj diff %v", cmp.Diff(got.String(), tt.want.String()))
				t.Errorf("C2D_Workspace() = %v, want %v", got.String(), tt.want.String())
			}
		})
	}
}

func TestC2D_NetworkRules(t *testing.T) {
	type args struct {
		netRules []cosmov1alpha1.NetworkRule
		urlMap   map[string]string
	}
	tests := []struct {
		name string
		args args
		want []*dashv1alpha1.NetworkRule
	}{
		{
			name: "OK",
			args: args{
				netRules: []cosmov1alpha1.NetworkRule{
					{
						Protocol:         "http",
						PortNumber:       8080,
						CustomHostPrefix: "xxx",
						HTTPPath:         "/path",
						Public:           true,
					},
					{
						Protocol:   "http",
						PortNumber: 8443,
						HTTPPath:   "/",
					},
				},
				urlMap: map[string]string{
					"http://main/":     "https://main.example.com",
					"http://xxx/path":  "https://xxx.example.com/path",
					"http://port8443/": "https://port8443.example.com/",
				},
			},
			want: []*dashv1alpha1.NetworkRule{
				{
					PortNumber:       8080,
					CustomHostPrefix: "xxx",
					HttpPath:         "/path",
					Public:           true,
					Url:              "https://xxx.example.com/path",
				},
				{
					PortNumber: 8443,
					HttpPath:   "/",
					Url:        "https://port8443.example.com/",
				},
			},
		},
		{
			name: "empty",
			args: args{
				netRules: []cosmov1alpha1.NetworkRule{},
			},
			want: []*dashv1alpha1.NetworkRule{},
		},
		{
			name: "empty",
			args: args{
				netRules: nil,
			},
			want: []*dashv1alpha1.NetworkRule{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := C2D_NetworkRules(tt.args.netRules, tt.args.urlMap); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("C2D_NetworkRules() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestC2D_NetworkRule(t *testing.T) {
	type args struct {
		v cosmov1alpha1.NetworkRule
	}
	tests := []struct {
		name string
		args args
		want *dashv1alpha1.NetworkRule
	}{
		{
			name: "OK",
			args: args{
				v: cosmov1alpha1.NetworkRule{
					Protocol:         "http",
					PortNumber:       8080,
					CustomHostPrefix: "xxx",
					HTTPPath:         "/path",
					Public:           true,
				},
			},
			want: &dashv1alpha1.NetworkRule{
				PortNumber:       8080,
				CustomHostPrefix: "xxx",
				HttpPath:         "/path",
				Public:           true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := C2D_NetworkRule(tt.args.v); !reflect.DeepEqual(got.String(), tt.want.String()) {
				t.Errorf("C2D_NetworkRule() = %s, want %s", got.String(), tt.want.String())
			}
		})
	}
}

func TestD2C_NetworkRules(t *testing.T) {
	type args struct {
		netRules []*dashv1alpha1.NetworkRule
	}
	tests := []struct {
		name string
		args args
		want []cosmov1alpha1.NetworkRule
	}{
		{
			name: "OK",
			args: args{
				netRules: []*dashv1alpha1.NetworkRule{
					{
						PortNumber:       8080,
						CustomHostPrefix: "xxx",
						HttpPath:         "/path",
						Url:              "https://xxx.example.com/path",
						Public:           true,
					},
					{
						PortNumber: 8443,
					},
				},
			},
			want: []cosmov1alpha1.NetworkRule{
				{
					Protocol:         "http",
					PortNumber:       8080,
					CustomHostPrefix: "xxx",
					HTTPPath:         "/path",
					Public:           true,
				},
				{
					Protocol:   "http",
					PortNumber: 8443,
					HTTPPath:   "/",
				},
			},
		},
		{
			name: "empty",
			args: args{
				netRules: []*dashv1alpha1.NetworkRule{},
			},
			want: []cosmov1alpha1.NetworkRule{},
		},
		{
			name: "empty",
			args: args{
				netRules: nil,
			},
			want: []cosmov1alpha1.NetworkRule{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := D2C_NetworkRules(tt.args.netRules); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("D2C_NetworkRules() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestD2C_NetworkRule(t *testing.T) {
	type args struct {
		v *dashv1alpha1.NetworkRule
	}
	tests := []struct {
		name string
		args args
		want cosmov1alpha1.NetworkRule
	}{
		{
			name: "OK",
			args: args{
				v: &dashv1alpha1.NetworkRule{
					PortNumber:       8080,
					CustomHostPrefix: "xxx",
					HttpPath:         "/path",
					Url:              "https://xxx.example.com/path",
					Public:           true,
				},
			},
			want: cosmov1alpha1.NetworkRule{
				Protocol:         "http",
				PortNumber:       8080,
				CustomHostPrefix: "xxx",
				HTTPPath:         "/path",
				Public:           true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := D2C_NetworkRule(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("D2C_NetworkRule() = %v, want %v", got, tt.want)
			}
		})
	}
}
