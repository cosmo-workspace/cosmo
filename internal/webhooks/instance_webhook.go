package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	netv1 "k8s.io/api/networking/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

type InstanceMutationWebhookHandler struct {
	Client  kosmo.Client
	Log     *clog.Logger
	decoder *admission.Decoder
}

//+kubebuilder:webhook:path=/mutate-cosmo-cosmo-workspace-github-io-v1alpha1-instance,mutating=true,failurePolicy=fail,sideEffects=None,groups=cosmo.cosmo-workspace.github.io,resources=instances,verbs=create;update,versions=v1alpha1,name=minstance.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *InstanceMutationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/mutate-cosmo-cosmo-workspace-github-io-v1alpha1-instance",
		&webhook.Admission{Handler: h},
	)
}

func (h *InstanceMutationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	inst := &cosmov1alpha1.Instance{}
	err := h.decoder.Decode(req, inst)
	if err != nil {
		h.Log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// mutate the fields in instance

	// fetch template
	tmpl, err := h.Client.GetTemplate(ctx, inst.Spec.Template.Name)
	if err != nil {
		h.Log.Error(err, "failed to get template")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// propagate template type annotation to instance annotation
	if tmplType, ok := template.GetTemplateType(tmpl); ok {
		template.SetTemplateType(inst, tmplType)
	}

	// defaulting required vars
	for _, v := range tmpl.Spec.RequiredVars {
		found := false
		for key := range inst.Spec.Vars {
			if template.FixupTemplateVarKey(key) == template.FixupTemplateVarKey(v.Var) {
				found = true
			}
		}
		if !found && v.Default != "" {
			if inst.Spec.Vars == nil {
				inst.Spec.Vars = make(map[string]string)
			}
			inst.Spec.Vars[v.Var] = v.Default
		}
	}

	// update name to instance fixed resource name
	netSpec := inst.Spec.Override.Network
	if netSpec != nil {
		for i, ingSpec := range netSpec.Ingress {
			if ingSpec.TargetName != "" {
				netSpec.Ingress[i].TargetName = cosmov1alpha1.InstanceResourceName(inst.GetName(), ingSpec.TargetName)
				fixIngressBackendName(netSpec.Ingress[i].Rules, inst.GetName())
			}
		}

		for i, svcSpec := range netSpec.Service {
			if svcSpec.TargetName != "" {
				netSpec.Service[i].TargetName = cosmov1alpha1.InstanceResourceName(inst.GetName(), svcSpec.TargetName)
			}
		}
	}

	scaleSpecs := inst.Spec.Override.Scale
	for i, scaleSpec := range scaleSpecs {
		if scaleSpec.Target.Name != "" {
			scaleSpecs[i].Target.Name = cosmov1alpha1.InstanceResourceName(inst.GetName(), scaleSpec.Target.Name)
		}
	}

	patchSpec := inst.Spec.Override.PatchesJson6902
	for i, p := range patchSpec {
		if p.Target.Name != "" {
			patchSpec[i].Target.Name = cosmov1alpha1.InstanceResourceName(inst.GetName(), p.Target.Name)
		}
	}

	marshaled, err := json.Marshal(inst)
	if err != nil {
		h.Log.Error(err, "failed to marshal response")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func fixIngressBackendName(ingRules []netv1.IngressRule, instName string) {
	for _, rule := range ingRules {
		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service != nil {
				path.Backend.Service.Name = cosmov1alpha1.InstanceResourceName(instName, path.Backend.Service.Name)
			}
			if path.Backend.Resource != nil {
				path.Backend.Resource.Name = cosmov1alpha1.InstanceResourceName(instName, path.Backend.Resource.Name)
			}
		}
	}
}

func (h *InstanceMutationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

type InstanceValidationWebhookHandler struct {
	Client  kosmo.Client
	Log     *clog.Logger
	decoder *admission.Decoder
}

//+kubebuilder:webhook:path=/validate-cosmo-cosmo-workspace-github-io-v1alpha1-instance,mutating=false,failurePolicy=fail,sideEffects=None,groups=cosmo.cosmo-workspace.github.io,resources=instances,verbs=create;update,versions=v1alpha1,name=vinstance.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *InstanceValidationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/validate-cosmo-cosmo-workspace-github-io-v1alpha1-instance",
		&webhook.Admission{Handler: h},
	)
}

func (h *InstanceValidationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	inst := &cosmov1alpha1.Instance{}
	err := h.decoder.Decode(req, inst)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// validate the fields in instance

	// whether template exist
	tmpl, err := h.Client.GetTemplate(ctx, inst.Spec.Template.Name)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return admission.Denied("Template not found")
		} else {
			h.Log.Error(err, "failed to get template")
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	// whether instance overrides template's required vars
	for _, v := range tmpl.Spec.RequiredVars {
		ok := false
		for key := range inst.Spec.Vars {
			if template.FixupTemplateVarKey(key) == template.FixupTemplateVarKey(v.Var) {
				ok = true
				break
			}
		}
		if !ok {
			return admission.Denied(fmt.Sprintf("Insufficient vars override: Var %v is required", v.Var))
		}
	}

	scaleSpecs := inst.Spec.Override.Scale
	for _, scaleSpec := range scaleSpecs {
		if _, err := schema.ParseGroupVersion(scaleSpec.Target.APIVersion); err != nil {
			return admission.Denied(fmt.Sprintf("APIVersion '%s' is invalid: %v", scaleSpec.Target.APIVersion, err))
		}
	}

	patchSpecs := inst.Spec.Override.PatchesJson6902
	for _, patchSpec := range patchSpecs {
		if _, err := schema.ParseGroupVersion(patchSpec.Target.APIVersion); err != nil {
			return admission.Denied(fmt.Sprintf("APIVersion '%s' is invalid: %v", patchSpec.Target.APIVersion, err))
		}
	}

	return admission.Allowed("OK")
}

func (h *InstanceValidationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}
