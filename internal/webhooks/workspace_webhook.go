package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	"github.com/cosmo-workspace/cosmo/pkg/wscfg"
	"github.com/cosmo-workspace/cosmo/pkg/wsnet"
)

type WorkspaceMutationWebhookHandler struct {
	Client  kosmo.Client
	Log     *clog.Logger
	decoder *admission.Decoder
}

//+kubebuilder:webhook:path=/mutate-workspace-cosmo-workspace-github-io-v1alpha1-workspace,mutating=true,failurePolicy=fail,sideEffects=None,groups=workspace.cosmo-workspace.github.io,resources=workspaces,verbs=create;update,versions=v1alpha1,name=mworkspace.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *WorkspaceMutationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/mutate-workspace-cosmo-workspace-github-io-v1alpha1-workspace",
		&webhook.Admission{Handler: h},
	)
}

// Handle mutates the fields in workspace
func (h *WorkspaceMutationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	ws := &wsv1alpha1.Workspace{}
	err := h.decoder.Decode(req, ws)
	if err != nil {
		h.Log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}
	before := ws.DeepCopy()
	h.Log.DebugAll().DumpObject(h.Client.Scheme(), before, "request workspace")

	tmpl, err := h.Client.GetTemplate(ctx, ws.Spec.Template.Name)
	if err != nil {
		h.Log.Error(err, "failed to get template")
		return admission.Errored(http.StatusBadRequest, err)
	}

	cfg, err := wscfg.ConfigFromTemplateAnnotations(tmpl)
	if err != nil {
		h.Log.Error(err, "failed to get config")
		return admission.Errored(http.StatusBadRequest, err)
	}
	h.Log.Debug().Info("workspace config in template", "cfg", cfg)

	// default replica 1
	if ws.Spec.Replicas == nil {
		var rep int64 = 1
		ws.Spec.Replicas = &rep
	}

	// migrate template service and ingress to network rule
	if err := h.migrateTmplServiceAndIngressToNetworkRule(ws, tmpl.Spec.RawYaml, cfg); err != nil {
		h.Log.Error(err, "failed to migrate service and ingress to network rule")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// fill default value in network rules
	h.defaultNetworkRules(ws.Spec.Network, ws.GetName(), ws.GetNamespace(), wsnet.URLBase(cfg.URLBase))

	// sort network rules
	ws.Spec.Network = sortNetworkRule(ws.Spec.Network)

	h.Log.Debug().PrintObjectDiff(before, ws)

	marshaled, err := json.Marshal(ws)
	if err != nil {
		h.Log.Error(err, "failed to marshal resoponse")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func (h *WorkspaceMutationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

type WorkspaceValidationWebhookHandler struct {
	Client  kosmo.Client
	Log     *clog.Logger
	decoder *admission.Decoder
}

//+kubebuilder:webhook:path=/validate-workspace-cosmo-workspace-github-io-v1alpha1-workspace,mutating=false,failurePolicy=fail,sideEffects=None,groups=workspace.cosmo-workspace.github.io,resources=workspaces,verbs=create;update,versions=v1alpha1,name=vworkspace.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *WorkspaceValidationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/validate-workspace-cosmo-workspace-github-io-v1alpha1-workspace",
		&webhook.Admission{Handler: h},
	)
}

// Handle validates the fields in Workspace
func (h *WorkspaceValidationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	ws := &wsv1alpha1.Workspace{}
	err := h.decoder.Decode(req, ws)
	if err != nil {
		h.Log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}
	h.Log.DebugAll().DumpObject(h.Client.Scheme(), ws, "request workspace")

	// check namespace for Workspace
	userid := wsv1alpha1.UserIDByNamespace(ws.GetNamespace())
	if userid == "" {
		return admission.Denied(fmt.Sprintf("namespace '%s' is not cosmo user's namespace", ws.GetNamespace()))
	}

	// check netrules
	if err := checkNetworkRules(ws.Spec.Network); err != nil {
		h.Log.Error(err, "network rules check failed")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// check netrule ports duplication
	dupPort := duplicatedPort(ws.Spec.Network)
	if dupPort > 0 {
		return admission.Denied(fmt.Sprintf("port '%d' is duplicated", dupPort))
	}

	return admission.Allowed("Validation OK")
}

func (h *WorkspaceValidationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func sortNetworkRule(netRules []wsv1alpha1.NetworkRule) []wsv1alpha1.NetworkRule {
	sort.SliceStable(netRules, func(i, j int) bool {
		if *netRules[i].Group < *netRules[j].Group {
			return true
		} else if *netRules[i].Group == *netRules[j].Group {
			return netRules[i].HTTPPath > netRules[j].HTTPPath
		}
		return false
	})
	return netRules
}

func (h *WorkspaceMutationWebhookHandler) defaultNetworkRules(netRules []wsv1alpha1.NetworkRule, name, namespace string, urlBase wsnet.URLBase) {
	for i := range netRules {
		netRules[i].Default()
		netRules[i].Host = pointer.String(wsnet.GenerateIngressHost(netRules[i], name, namespace, urlBase))
		h.Log.Debug().Info("defaulting network rule", "netRule", netRules[i])
	}
}

func checkNetworkRules(netRules []wsv1alpha1.NetworkRule) error {
	for _, netRule := range netRules {
		if netRule.PortName == "" {
			return errors.New("port name is empty")
		}
		if netRule.PortNumber == 0 {
			return errors.New("port number is 0")
		}
	}
	return nil
}

func duplicatedPort(netRules []wsv1alpha1.NetworkRule) int {
	for _, netRule := range netRules {
		for _, v := range netRules {
			if netRule.PortName != v.PortName && netRule.PortNumber == v.PortNumber {
				return netRule.PortNumber
			}
		}
	}
	return 0
}

func (h *WorkspaceMutationWebhookHandler) migrateTmplServiceAndIngressToNetworkRule(ws *wsv1alpha1.Workspace, rawTmpl string, cfg wsv1alpha1.Config) error {
	unst, err := preTemplateBuild(*ws, rawTmpl)
	if err != nil {
		return err
	}

	var svc corev1.Service
	var ing netv1.Ingress
	for _, u := range unst {
		h.Log.Debug().Info("template resources", "gvk", u.GroupVersionKind(), "name", u.GetName(), "unstructured", u)

		h.Log.DebugAll().Info("workspace config in template",
			"gvk", u.GroupVersionKind(),
			"cfgServiceName", cfg.ServiceName, "cfgIngressName", cfg.IngressName,
			"instFixedName", cosmov1alpha1.InstanceResourceName(template.DefaultVarsInstance, u.GetName()),
			"svcGvkEqual", cosmov1alpha1.IsGVKEqual(u.GroupVersionKind(), kubeutil.ServiceGVK),
			"ingGvkEqual", cosmov1alpha1.IsGVKEqual(u.GroupVersionKind(), kubeutil.IngressGVK),
			"svcNameEqual", cosmov1alpha1.EqualInstanceResourceName(template.DefaultVarsInstance, u.GetName(), cfg.ServiceName),
			"ingNameEqual", cosmov1alpha1.EqualInstanceResourceName(template.DefaultVarsInstance, u.GetName(), cfg.IngressName),
		)

		if cosmov1alpha1.IsGVKEqual(u.GroupVersionKind(), kubeutil.ServiceGVK) &&
			cosmov1alpha1.EqualInstanceResourceName(template.DefaultVarsInstance, u.GetName(), cfg.ServiceName) {
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &svc)
			if err != nil {
				return err
			}

		} else if cosmov1alpha1.IsGVKEqual(u.GroupVersionKind(), kubeutil.IngressGVK) &&
			cosmov1alpha1.EqualInstanceResourceName(template.DefaultVarsInstance, u.GetName(), cfg.IngressName) {
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &ing)
			if err != nil {
				return err
			}
		}
	}
	h.Log.Debug().Info("service and ingress in template", "service", svc, "ingress", ing)

	netRules := wsv1alpha1.NetworkRulesByServiceAndIngress(svc, ing)
	h.Log.Info("generated netrules by service and ingress in template", "netRules", netRules)

	// append network rules
	for _, netRule := range netRules {
		found := false
		for _, r := range ws.Spec.Network {
			if netRule.PortName == r.PortName {
				found = true
			}
		}
		if !found {
			ws.Spec.Network = append(ws.Spec.Network, netRule)
		}
	}
	return nil
}

func preTemplateBuild(ws wsv1alpha1.Workspace, rawTmpl string) ([]unstructured.Unstructured, error) {
	var inst cosmov1alpha1.Instance
	inst.SetName(ws.GetName())
	inst.SetNamespace(ws.GetNamespace())

	builder := template.NewRawYAMLBuilder(rawTmpl, &inst)
	return builder.Build()
}
