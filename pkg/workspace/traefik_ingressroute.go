package workspace

import (
	"fmt"
	"strings"

	traefikv1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
)

type TraefikIngressRouteConfig struct {
	// Entrypoints is the entrypoint of traefik ingress route
	Entrypoints []string
	// TLS is the TLS of traefik ingress route
	TLS *traefikv1.TLS
	// AuthenMiddleware is the name and namespace of middleware for cosmo-auth
	// Namespace must be the same as where trafik LB is running
	AuthenMiddleware traefikv1.MiddlewareRef
	// UserNameHeaderMiddlewareName is the name of middleware for username header
	// Namespace must be empty to be the same as the workspace
	UserNameHeaderMiddleware traefikv1.MiddlewareRef

	// HostBase is a base of hostname
	HostBase string
	// Domain is a domain of hostname
	Domain string
}

func (c *TraefikIngressRouteConfig) PatchTraefikIngressRouteAsDesired(ir *traefikv1.IngressRoute, ws cosmov1alpha1.Workspace, scheme *runtime.Scheme) error {
	// metadata
	cosmov1alpha1.SetControllerManaged(ir)

	// spec.entrypoints
	ir.Spec.EntryPoints = c.Entrypoints

	// spec.tls
	ir.Spec.TLS = c.TLS

	// spec.routes
	routes := make([]traefikv1.Route, 0, len(ws.Spec.Network))
	for _, netRule := range ws.Spec.Network {
		traefikRule := c.TraefikRoute(netRule, ws)
		routes = append(routes, traefikRule)
	}
	ir.Spec.Routes = routes

	if scheme != nil {
		err := ctrl.SetControllerReference(&ws, ir, scheme)
		if err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}
	}
	return nil
}

func (c *TraefikIngressRouteConfig) TraefikRoute(r cosmov1alpha1.NetworkRule, ws cosmov1alpha1.Workspace) traefikv1.Route {
	matches := []string{}

	matches = append(matches, fmt.Sprintf("Host(`%s`)", cosmov1alpha1.GenHost(c.HostBase, c.Domain, r.HostPrefix(), ws)))

	if r.HTTPPath != "" && r.HTTPPath != "/" {
		matches = append(matches, fmt.Sprintf("PathPrefix(`%s`)", r.HTTPPath))
	}
	match := strings.Join(matches[:], " && ")

	var middlewares []traefikv1.MiddlewareRef
	if r.Public {
		middlewares = []traefikv1.MiddlewareRef{}
	} else {
		middlewares = []traefikv1.MiddlewareRef{c.UserNameHeaderMiddleware, c.AuthenMiddleware}
	}

	backendSvcName := instance.InstanceResourceName(ws.Name, ws.Status.Config.ServiceName)

	return traefikv1.Route{
		Kind:     "Rule",
		Match:    match,
		Priority: 100,
		Services: []traefikv1.Service{
			{
				LoadBalancerSpec: traefikv1.LoadBalancerSpec{
					Kind:   "Service",
					Name:   backendSvcName,
					Port:   intstr.FromInt(int(r.PortNumber)),
					Scheme: "http",
				},
			},
		},
		Middlewares: middlewares,
	}
}
