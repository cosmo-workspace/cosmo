package workspace

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	traefikv1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
)

func TestTraefikIngressRouteConfig_PatchTraefikIngressRouteAsDesired(t *testing.T) {
	utilruntime.Must(cosmov1alpha1.AddToScheme(scheme.Scheme))
	type fields struct {
		Entrypoints              []string
		TLS                      *traefikv1.TLS
		AuthenMiddleware         traefikv1.MiddlewareRef
		UserNameHeaderMiddleware traefikv1.MiddlewareRef
	}
	type args struct {
		ir     *traefikv1.IngressRoute
		ws     cosmov1alpha1.Workspace
		scheme *runtime.Scheme
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				Entrypoints: []string{"web", "websecure"},
				TLS:         nil,
				AuthenMiddleware: traefikv1.MiddlewareRef{
					Name:      "cosmo-auth",
					Namespace: "cosmo-system",
				},
				UserNameHeaderMiddleware: traefikv1.MiddlewareRef{
					Name: "userNameHeader",
				},
			},
			args: args{
				ir: &traefikv1.IngressRoute{},
				ws: cosmov1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ws1",
						Namespace: "cosmo-user-xxx",
					},
					Spec: cosmov1alpha1.WorkspaceSpec{
						Network: []cosmov1alpha1.NetworkRule{
							{
								PortNumber:       8080,
								HTTPPath:         "/",
								CustomHostPrefix: "main",
								TargetPortNumber: pointer.Int32(18080),
								Public:           false,
							},
							{
								PortNumber: 8080,
								HTTPPath:   "/",
								Public:     true,
							},
							{
								PortNumber: 8080,
								HTTPPath:   "/dev",
								Public:     false,
							},
						},
					},
					Status: cosmov1alpha1.WorkspaceStatus{
						Config: cosmov1alpha1.Config{
							ServiceName: "svc",
						},
					},
				},
				scheme: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &TraefikIngressRouteConfig{
				Entrypoints:              tt.fields.Entrypoints,
				TLS:                      tt.fields.TLS,
				AuthenMiddleware:         tt.fields.AuthenMiddleware,
				UserNameHeaderMiddleware: tt.fields.UserNameHeaderMiddleware,
			}
			if err := c.PatchTraefikIngressRouteAsDesired(tt.args.ir, tt.args.ws, tt.args.scheme); (err != nil) != tt.wantErr {
				t.Errorf("TraefikIngressRouteConfig.PatchTraefikIngressRouteAsDesired() error = %v, wantErr %v", err, tt.wantErr)
			}
			snaps.MatchJSON(t, tt.args.ir)
		})
	}
}

func TestTraefikIngressRouteConfig_TraefikRoute(t *testing.T) {
	type fields struct {
		Entrypoints              []string
		TLS                      *traefikv1.TLS
		AuthenMiddleware         traefikv1.MiddlewareRef
		UserNameHeaderMiddleware traefikv1.MiddlewareRef
	}
	type args struct {
		r  cosmov1alpha1.NetworkRule
		ws cosmov1alpha1.Workspace
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "default",
			fields: fields{
				AuthenMiddleware: traefikv1.MiddlewareRef{
					Name:      "cosmo-auth",
					Namespace: "cosmo-system",
				},
				UserNameHeaderMiddleware: traefikv1.MiddlewareRef{
					Name: "userNameHeader",
				},
			},
			args: args{
				r: cosmov1alpha1.NetworkRule{
					PortNumber: 8080,
					HTTPPath:   "/",
					Public:     false,
				},
				ws: cosmov1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ws1",
						Namespace: "cosmo-user-xxx",
					},
					Status: cosmov1alpha1.WorkspaceStatus{
						Config: cosmov1alpha1.Config{
							ServiceName: "backend-svc-name",
						},
					},
				},
			},
		},
		{
			name: "no hostname",
			fields: fields{
				AuthenMiddleware: traefikv1.MiddlewareRef{
					Name:      "cosmo-auth",
					Namespace: "cosmo-system",
				},
				UserNameHeaderMiddleware: traefikv1.MiddlewareRef{
					Name: "userNameHeader",
				},
			},
			args: args{
				r: cosmov1alpha1.NetworkRule{
					PortNumber: 8080,
					HTTPPath:   "/path",
					Public:     false,
				},

				ws: cosmov1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ws1",
						Namespace: "cosmo-user-xxx",
					},
					Status: cosmov1alpha1.WorkspaceStatus{
						Config: cosmov1alpha1.Config{
							ServiceName: "backend-svc-name",
						},
					},
				},
			},
		},
		{
			name: "public",
			fields: fields{
				AuthenMiddleware: traefikv1.MiddlewareRef{
					Name:      "cosmo-auth",
					Namespace: "cosmo-system",
				},
				UserNameHeaderMiddleware: traefikv1.MiddlewareRef{
					Name: "userNameHeader",
				},
			},
			args: args{
				r: cosmov1alpha1.NetworkRule{
					PortNumber: 8080,
					HTTPPath:   "/path",
					Public:     true,
				},

				ws: cosmov1alpha1.Workspace{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ws1",
						Namespace: "cosmo-user-xxx",
					},
					Status: cosmov1alpha1.WorkspaceStatus{
						Config: cosmov1alpha1.Config{
							ServiceName: "backend-svc-name",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &TraefikIngressRouteConfig{
				Entrypoints:              tt.fields.Entrypoints,
				TLS:                      tt.fields.TLS,
				AuthenMiddleware:         tt.fields.AuthenMiddleware,
				UserNameHeaderMiddleware: tt.fields.UserNameHeaderMiddleware,
			}
			got := c.TraefikRoute(tt.args.r, tt.args.ws)
			snaps.MatchJSON(t, got)
		})
	}
}
