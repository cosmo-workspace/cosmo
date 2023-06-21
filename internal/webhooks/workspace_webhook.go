package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	"github.com/cosmo-workspace/cosmo/pkg/workspace"
)

type WorkspaceMutationWebhookHandler struct {
	Client  client.Client
	Log     *clog.Logger
	URLBase workspace.URLBase
	decoder *admission.Decoder
}

//+kubebuilder:webhook:path=/mutate-cosmo-workspace-github-io-v1alpha1-workspace,mutating=true,failurePolicy=fail,sideEffects=None,groups=cosmo-workspace.github.io,resources=workspaces,verbs=create;update,versions=v1alpha1,name=mworkspace.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *WorkspaceMutationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/mutate-cosmo-workspace-github-io-v1alpha1-workspace",
		&webhook.Admission{Handler: h},
	)
}

// Handle mutates the fields in workspace
func (h *WorkspaceMutationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := h.Log.WithValues("UID", req.UID, "GroupVersionKind", req.Kind.String(), "Name", req.Name, "Namespace", req.Namespace)
	ctx = clog.IntoContext(ctx, log)

	ws := &cosmov1alpha1.Workspace{}
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

	cfg, err := workspace.ConfigFromTemplateAnnotations(tmpl)
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

	// migrate template service to network rule
	if err := h.migrateTmplServiceToNetworkRule(ctx, ws, tmpl.Spec.RawYaml, cfg); err != nil {
		log.Error(err, "failed to migrate service to network rule")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// fill default value in network rules
	h.defaultNetworkRules(ws.Spec.Network, ws.GetName(), ws.GetNamespace(), h.URLBase)

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

//+kubebuilder:webhook:path=/validate-cosmo-workspace-github-io-v1alpha1-workspace,mutating=false,failurePolicy=fail,sideEffects=None,groups=cosmo-workspace.github.io,resources=workspaces,verbs=create;update,versions=v1alpha1,name=vworkspace.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *WorkspaceValidationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/validate-cosmo-workspace-github-io-v1alpha1-workspace",
		&webhook.Admission{Handler: h},
	)
}

// Handle validates the fields in Workspace
func (h *WorkspaceValidationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := h.Log.WithValues("UID", req.UID, "GroupVersionKind", req.Kind.String(), "Name", req.Name, "Namespace", req.Namespace)

	ws := &cosmov1alpha1.Workspace{}
	err := h.decoder.Decode(req, ws)
	if err != nil {
		log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.DumpObject(h.Client.Scheme(), ws, "request workspace")

	// check namespace for Workspace
	username := cosmov1alpha1.UserNameByNamespace(ws.GetNamespace())
	if username == "" {
		return admission.Denied(fmt.Sprintf("namespace '%s' is not cosmo user's namespace", ws.GetNamespace()))
	}

	// check netrules
	if err := checkNetworkRules(ws.Spec.Network); err != nil {
		log.Error(err, "network rules check failed")
		return admission.Errored(http.StatusBadRequest, err)
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

func sortNetworkRule(netRules []cosmov1alpha1.NetworkRule) []cosmov1alpha1.NetworkRule {
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

func (h *WorkspaceMutationWebhookHandler) defaultNetworkRules(netRules []cosmov1alpha1.NetworkRule, name, namespace string, urlBase workspace.URLBase) {
	for i := range netRules {
		netRules[i].Default()
		netRules[i].Host = pointer.String(workspace.GenerateIngressHost(netRules[i], name, namespace, urlBase))
	}
}

func checkNetworkRules(netRules []cosmov1alpha1.NetworkRule) error {
	for i, netRule := range netRules {
		if errs := validation.IsValidPortName(netRule.Name); len(errs) > 0 {
			return errors.New(errs[0])
		}
		if errs := validation.IsValidPortNum(int(netRule.PortNumber)); len(errs) > 0 {
			return errors.New(errs[0])
		}
		for j, v := range netRules {
			if i == j {
				continue
			}
			if netRule.Name == v.Name {
				return errors.New("duplicate network rule name")
			}
			if reflect.DeepEqual(netRule.Group, v.Group) &&
				netRule.HTTPPath == v.HTTPPath {
				return fmt.Errorf("duplicate group and path. group=%s,path=%s", *v.Group, v.HTTPPath)
			}
			if reflect.DeepEqual(netRule.Host, v.Host) &&
				netRule.HTTPPath == v.HTTPPath {
				return fmt.Errorf("duplicate host and path. host=%s,path=%s", *v.Host, v.HTTPPath)
			}
		}
	}
	return nil
}

func (h *WorkspaceMutationWebhookHandler) migrateTmplServiceToNetworkRule(ctx context.Context, ws *cosmov1alpha1.Workspace, rawTmpl string, cfg cosmov1alpha1.Config) error {
	log := clog.FromContext(ctx).WithCaller()

	unst, err := preTemplateBuild(*ws, rawTmpl)
	if err != nil {
		return err
	}

	var svc corev1.Service
	for _, u := range unst {
		log.Debug().Info(fmt.Sprintf("template resources: %v", u), "resourceGVK", u.GroupVersionKind(), "resourceName", u.GetName())

		log.DebugAll().Info(fmt.Sprintf("workspace config in template: %v", cfg),
			"gvk", u.GroupVersionKind(),
			"cfgServiceName", cfg.ServiceName,
			"instFixedName", instance.InstanceResourceName(template.DefaultVarsInstance, u.GetName()),
			"svcGvkEqual", kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.ServiceGVK),
			"svcNameEqual", instance.EqualInstanceResourceName(template.DefaultVarsInstance, u.GetName(), cfg.ServiceName),
		)

		if kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.ServiceGVK) &&
			instance.EqualInstanceResourceName(template.DefaultVarsInstance, u.GetName(), cfg.ServiceName) {
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &svc)
			if err != nil {
				return err
			}

		}
	}
	log.DebugAll().Info("service in template", "service", svc)

	netRules := cosmov1alpha1.NetworkRulesByService(svc)

	// append network rules
	for _, netRule := range netRules {
		found := false
		for _, r := range ws.Spec.Network {
			if netRule.Name == r.Name {
				found = true
			}
		}
		if !found {
			log.Info("generated netrules by service in template", "netRule", netRule)
			ws.Spec.Network = append(ws.Spec.Network, netRule)
		}
	}
	return nil
}

func preTemplateBuild(ws cosmov1alpha1.Workspace, rawTmpl string) ([]unstructured.Unstructured, error) {
	var inst cosmov1alpha1.Instance
	inst.SetName(ws.GetName())
	inst.SetNamespace(ws.GetNamespace())

	builder := template.NewRawYAMLBuilder(rawTmpl, &inst)
	return builder.Build()
}
