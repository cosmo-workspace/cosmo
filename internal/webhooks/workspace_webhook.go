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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	"github.com/cosmo-workspace/cosmo/pkg/wscfg"
	"github.com/cosmo-workspace/cosmo/pkg/wsnet"
)

type WorkspaceMutationWebhookHandler struct {
	Client  client.Client
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
	log := h.Log.WithValues("UID", req.UID, "GroupVersionKind", req.Kind.String(), "Name", req.Name, "Namespace", req.Namespace)
	ctx = clog.IntoContext(ctx, log)

	ws := &wsv1alpha1.Workspace{}
	err := h.decoder.Decode(req, ws)
	if err != nil {
		log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}
	before := ws.DeepCopy()
	log.DumpObject(h.Client.Scheme(), before, "request workspace")

	tmpl := &cosmov1alpha1.Template{}
	err = h.Client.Get(ctx, types.NamespacedName{Name: ws.Spec.Template.Name}, tmpl)
	if err != nil {
		log.Error(err, "failed to get template")
		return admission.Errored(http.StatusBadRequest, err)
	}

	cfg, err := wscfg.ConfigFromTemplateAnnotations(tmpl)
	if err != nil {
		log.Error(err, "failed to get config")
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.Debug().Info(fmt.Sprintf("workspace config in template %s", cfg))

	// default replica 1
	if ws.Spec.Replicas == nil {
		var rep int64 = 1
		ws.Spec.Replicas = &rep
	}

	// migrate template service and ingress to network rule
	if err := h.migrateTmplServiceAndIngressToNetworkRule(ctx, ws, tmpl.Spec.RawYaml, cfg); err != nil {
		log.Error(err, "failed to migrate service and ingress to network rule")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// fill default value in network rules
	h.defaultNetworkRules(ws.Spec.Network, ws.GetName(), ws.GetNamespace(), wsnet.URLBase(cfg.URLBase))

	// sort network rules
	ws.Spec.Network = sortNetworkRule(ws.Spec.Network)

	log.Debug().PrintObjectDiff(before, ws)

	marshaled, err := json.Marshal(ws)
	if err != nil {
		log.Error(err, "failed to marshal resoponse")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func (h *WorkspaceMutationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

type WorkspaceValidationWebhookHandler struct {
	Client  client.Client
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
	log := h.Log.WithValues("UID", req.UID, "GroupVersionKind", req.Kind.String(), "Name", req.Name, "Namespace", req.Namespace)
	ctx = clog.IntoContext(ctx, log)

	ws := &wsv1alpha1.Workspace{}
	err := h.decoder.Decode(req, ws)
	if err != nil {
		log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.DumpObject(h.Client.Scheme(), ws, "request workspace")

	// check namespace for Workspace
	userid := wsv1alpha1.UserIDByNamespace(ws.GetNamespace())
	if userid == "" {
		return admission.Denied(fmt.Sprintf("namespace '%s' is not cosmo user's namespace", ws.GetNamespace()))
	}

	// check netrules
	if err := checkNetworkRules(ws.Spec.Network); err != nil {
		log.Error(err, "network rules check failed")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// check netrule ports duplication
	dupPort := duplicatedPort(ws.Spec.Network)
	if dupPort > 0 {
		return admission.Denied(fmt.Sprintf("port '%d' is duplicated", dupPort))
	}

	// TODO
	// // dryrun
	// inst := &cosmov1alpha1.Instance{}
	// inst.SetName(ws.Name)
	// inst.SetNamespace(ws.Namespace)

	// _, err = kubeutil.DryrunCreateOrUpdate(ctx, h.Client, inst, func() error {
	// 	return workspace.PatchWorkspaceInstanceAsDesired(inst, *ws, nil)
	// })
	// if err != nil {
	// 	return admission.Denied(fmt.Sprintf("failed to dryrun create or update workspace instance: %v", err))
	// }

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
	}
}

func checkNetworkRules(netRules []wsv1alpha1.NetworkRule) error {
	for _, netRule := range netRules {
		if errs := validation.IsValidPortName(netRule.Name); len(errs) > 0 {
			return errors.New(errs[0])
		}
		if errs := validation.IsValidPortNum(netRule.PortNumber); len(errs) > 0 {
			return errors.New(errs[0])
		}
	}
	return nil
}

func duplicatedPort(netRules []wsv1alpha1.NetworkRule) int {
	for _, netRule := range netRules {
		for _, v := range netRules {
			if netRule.Name != v.Name && netRule.PortNumber == v.PortNumber {
				return netRule.PortNumber
			}
		}
	}
	return 0
}

func (h *WorkspaceMutationWebhookHandler) migrateTmplServiceAndIngressToNetworkRule(ctx context.Context, ws *wsv1alpha1.Workspace, rawTmpl string, cfg wsv1alpha1.Config) error {
	log := clog.FromContext(ctx).WithCaller()

	unst, err := preTemplateBuild(*ws, rawTmpl)
	if err != nil {
		return err
	}

	var svc corev1.Service
	var ing netv1.Ingress
	for _, u := range unst {
		log.Debug().Info(fmt.Sprintf("template resources: %v", u), "resourceGVK", u.GroupVersionKind(), "resourceName", u.GetName())

		log.DebugAll().Info(fmt.Sprintf("workspace config in template: %v", cfg),
			"gvk", u.GroupVersionKind(),
			"cfgServiceName", cfg.ServiceName, "cfgIngressName", cfg.IngressName,
			"instFixedName", instance.InstanceResourceName(template.DefaultVarsInstance, u.GetName()),
			"svcGvkEqual", kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.ServiceGVK),
			"ingGvkEqual", kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.IngressGVK),
			"svcNameEqual", instance.EqualInstanceResourceName(template.DefaultVarsInstance, u.GetName(), cfg.ServiceName),
			"ingNameEqual", instance.EqualInstanceResourceName(template.DefaultVarsInstance, u.GetName(), cfg.IngressName),
		)

		if kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.ServiceGVK) &&
			instance.EqualInstanceResourceName(template.DefaultVarsInstance, u.GetName(), cfg.ServiceName) {
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &svc)
			if err != nil {
				return err
			}

		} else if kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.IngressGVK) &&
			instance.EqualInstanceResourceName(template.DefaultVarsInstance, u.GetName(), cfg.IngressName) {
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &ing)
			if err != nil {
				return err
			}
		}
	}
	log.Debug().Info("service and ingress in template", "service", svc, "ingress", ing)

	netRules := wsv1alpha1.NetworkRulesByServiceAndIngress(svc, ing)
	log.Info("generated netrules by service and ingress in template", "netRules", netRules)

	// append network rules
	for _, netRule := range netRules {
		found := false
		for _, r := range ws.Spec.Network {
			if netRule.Name == r.Name {
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
