package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

	err = h.mutateWorkspace(ctx, ws)
	if err != nil {
		log.Error(err, "muration error")
		return admission.Errored(http.StatusBadRequest, err)
	}

	log.Debug().PrintObjectDiff(before, ws)

	marshaled, err := json.Marshal(ws)
	if err != nil {
		log.Error(err, "failed to marshal resoponse")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func (h *WorkspaceMutationWebhookHandler) mutateWorkspace(ctx context.Context, ws *cosmov1alpha1.Workspace) error {
	tmpl := &cosmov1alpha1.Template{}
	err := h.Client.Get(ctx, types.NamespacedName{Name: ws.Spec.Template.Name}, tmpl)
	if err != nil {
		return fmt.Errorf("failed to fetch template '%s': %w", ws.Spec.Template.Name, err)
	}

	// default replica 1
	if ws.Spec.Replicas == nil {
		var rep int64 = 1
		ws.Spec.Replicas = &rep
	}

	// migrate template service to network rule
	cfg, err := workspace.ConfigFromTemplateAnnotations(tmpl)
	if err != nil {
		return fmt.Errorf("failed to get config from template: %w", err)
	}
	if err := h.migrateTmplServiceToNetworkRule(ctx, ws, tmpl.Spec.RawYaml, cfg); err != nil {
		return fmt.Errorf("failed to migrate service to network rule: %w", err)
	}

	// fill default value in network rules
	for i := range ws.Spec.Network {
		ws.Spec.Network[i].Default()
	}

	// sort network rules
	ws.Spec.Network = sortNetworkRule(ws.Spec.Network, ws.Status.Config)

	return nil
}

func (h *WorkspaceMutationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *WorkspaceMutationWebhookHandler) migrateTmplServiceToNetworkRule(ctx context.Context, ws *cosmov1alpha1.Workspace, rawTmpl string, cfg cosmov1alpha1.Config) error {
	unst, err := preTemplateBuild(*ws, rawTmpl)
	if err != nil {
		return err
	}

	svc, err := pickServiceInUnstructureds(unst, cfg.ServiceName)
	if err != nil {
		return err
	}

	// append network rules
	for _, netRule := range networkRulesByServicePorts(svc.Spec.Ports) {
		r := *netRule.DeepCopy()
		appendNetworkRuleIfNotExist(ws, r)
	}
	return nil
}

func appendNetworkRuleIfNotExist(ws *cosmov1alpha1.Workspace, netRule cosmov1alpha1.NetworkRule) {
	netRule.Default()
	for _, r := range ws.Spec.Network {
		if netRule.UniqueKey() == r.UniqueKey() {
			return
		}
	}
	ws.Spec.Network = append(ws.Spec.Network, netRule)
}

func preTemplateBuild(ws cosmov1alpha1.Workspace, rawTmpl string) ([]unstructured.Unstructured, error) {
	var inst cosmov1alpha1.Instance
	inst.SetName(ws.GetName())
	inst.SetNamespace(ws.GetNamespace())

	builder := template.NewRawYAMLBuilder(rawTmpl, &inst)
	return builder.Build()
}

func pickServiceInUnstructureds(objects []unstructured.Unstructured, serviceName string) (*corev1.Service, error) {
	var svc corev1.Service
	found := false
	for _, u := range objects {
		if kubeutil.IsGVKEqual(u.GroupVersionKind(), kubeutil.ServiceGVK) &&
			instance.EqualInstanceResourceName(template.DefaultVarsInstance, u.GetName(), serviceName) {

			found = true
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &svc)
			if err != nil {
				return nil, err
			}
		}
	}
	if !found {
		return nil, fmt.Errorf("Service '%s' not found", serviceName)
	}
	return &svc, nil
}

func networkRulesByServicePorts(svcPorts []corev1.ServicePort) []cosmov1alpha1.NetworkRule {
	netRules := make([]cosmov1alpha1.NetworkRule, 0, len(svcPorts))
	for _, p := range svcPorts {
		var netRule cosmov1alpha1.NetworkRule
		netRule.CustomHostPrefix = p.Name
		netRule.PortNumber = p.Port

		if p.TargetPort.IntValue() != 0 {
			netRule.TargetPortNumber = pointer.Int32(int32(p.TargetPort.IntValue()))
		}

		netRules = append(netRules, netRule)
	}
	return netRules
}

func sortNetworkRule(netRules []cosmov1alpha1.NetworkRule, cfg cosmov1alpha1.Config) []cosmov1alpha1.NetworkRule {
	sort.SliceStable(netRules, func(i, j int) bool {
		// move main rule to the top of netrules
		if cosmov1alpha1.MainRuleKey(cfg) == netRules[i].UniqueKey() {
			return true
		} else if cosmov1alpha1.MainRuleKey(cfg) == netRules[j].UniqueKey() {
			return false
		} else if netRules[i].CustomHostPrefix < netRules[j].CustomHostPrefix {
			// sort by CustomHostPrefix
			return true
		} else if netRules[i].CustomHostPrefix == netRules[j].CustomHostPrefix {
			// if CustomHostPrefix are the same, place in order of longer path
			return netRules[i].HTTPPath > netRules[j].HTTPPath
		}
		return false
	})
	return netRules
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
	log.DebugAll().DumpObject(h.Client.Scheme(), ws, "request workspace")

	err = h.validateWorkspace(ctx, ws)
	if err != nil {
		log.Error(err, "validation failed")
		return admission.Errored(http.StatusForbidden, err)
	}

	return admission.Allowed("Validation OK")
}

func (h *WorkspaceValidationWebhookHandler) validateWorkspace(ctx context.Context, ws *cosmov1alpha1.Workspace) error {
	// check namespace for Workspace
	username := cosmov1alpha1.UserNameByNamespace(ws.GetNamespace())
	if username == "" {
		return fmt.Errorf("namespace '%s' is not cosmo user's namespace", ws.GetNamespace())
	}

	// check netrules
	if err := checkNetworkRules(ws.Spec.Network); err != nil {
		return fmt.Errorf("network rules check failed: %w", err)
	}

	return nil
}

func (h *WorkspaceValidationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func checkNetworkRules(netRules []cosmov1alpha1.NetworkRule) error {
	for i, netRule := range netRules {
		if errs := validation.IsValidPortNum(int(netRule.PortNumber)); len(errs) > 0 {
			return fmt.Errorf("port validation failed: port=%d", netRule.PortNumber)
		}
		for j, v := range netRules {
			if i == j {
				continue
			}
			if netRule.UniqueKey() == v.UniqueKey() {
				r, _ := json.Marshal(v)
				return fmt.Errorf("duplicate network rules: %s", string(r))
			}
		}
	}
	return nil
}
