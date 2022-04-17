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
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	"github.com/cosmo-workspace/cosmo/pkg/transformer"
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
	before := inst.DeepCopy()
	h.Log.DebugAll().DumpObject(h.Client.Scheme(), before, "request instance")

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
				netSpec.Ingress[i].TargetName = instance.InstanceResourceName(inst.GetName(), ingSpec.TargetName)
				fixIngressBackendName(netSpec.Ingress[i].Rules, inst.GetName())
			}
		}

		for i, svcSpec := range netSpec.Service {
			if svcSpec.TargetName != "" {
				netSpec.Service[i].TargetName = instance.InstanceResourceName(inst.GetName(), svcSpec.TargetName)
			}
		}
	}

	scaleSpecs := inst.Spec.Override.Scale
	for i, scaleSpec := range scaleSpecs {
		if scaleSpec.Target.Name != "" {
			scaleSpecs[i].Target.Name = instance.InstanceResourceName(inst.GetName(), scaleSpec.Target.Name)
		}
	}

	patchSpec := inst.Spec.Override.PatchesJson6902
	for i, p := range patchSpec {
		if p.Target.Name != "" {
			patchSpec[i].Target.Name = instance.InstanceResourceName(inst.GetName(), p.Target.Name)
		}
	}

	h.Log.Debug().PrintObjectDiff(before, inst)

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
				path.Backend.Service.Name = instance.InstanceResourceName(instName, path.Backend.Service.Name)
			}
			if path.Backend.Resource != nil {
				path.Backend.Resource.Name = instance.InstanceResourceName(instName, path.Backend.Resource.Name)
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
	h.Log.DebugAll().DumpObject(h.Client.Scheme(), inst, "request instance")

	// validate the fields in instance

	// whether template exist
	tmpl, err := h.Client.GetTemplate(ctx, inst.Spec.Template.Name)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return admission.Denied(fmt.Sprintf("Template %s not found", inst.Spec.Template.Name))
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

	// whether valid apiVersion
	scaleSpecs := inst.Spec.Override.Scale
	for _, scaleSpec := range scaleSpecs {
		if _, err := schema.ParseGroupVersion(scaleSpec.Target.APIVersion); err != nil {
			return admission.Denied(fmt.Sprintf("APIVersion '%s' is invalid: %v", scaleSpec.Target.APIVersion, err))
		}
	}

	// whether valid apiVersion
	patchSpecs := inst.Spec.Override.PatchesJson6902
	for _, patchSpec := range patchSpecs {
		if _, err := schema.ParseGroupVersion(patchSpec.Target.APIVersion); err != nil {
			return admission.Denied(fmt.Sprintf("APIVersion '%s' is invalid: %v", patchSpec.Target.APIVersion, err))
		}
	}

	// dryrun apply
	if err := h.dryrunApply(ctx, tmpl, *inst); err != nil {
		return admission.Denied(err.Error())
	}

	return admission.Allowed("OK")
}

func (h *InstanceValidationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *InstanceValidationWebhookHandler) dryrunApply(ctx context.Context, tmpl *cosmov1alpha1.Template, inst cosmov1alpha1.Instance) error {
	builts, err := template.NewRawYAMLBuilder(tmpl.Spec.RawYaml, &inst).
		ReplaceDefaultVars().
		ReplaceCustomVars().
		Build()

	if err != nil {
		return fmt.Errorf("failed to build template: %w", err)
	}

	// Transform
	ts := []transformer.Transformer{
		// MetadataTransformer perform update each object's metadata
		transformer.NewMetadataTransformer(&inst, tmpl, h.Client.Scheme()),
		// NetworkTransformer perform update ingresses and services by network override
		transformer.NewNetworkTransformer(inst.Spec.Override.Network, inst.Name),
		// JSONPatchTransformer perform JSONPatch
		transformer.NewJSONPatchTransformer(inst.Spec.Override.PatchesJson6902, inst.Name),
		// ScalingTransformer perform override replicas
		transformer.NewScalingTransformer(inst.Spec.Override.Scale, inst.Name),
	}
	builts, err = transformer.ApplyTransformers(ctx, ts, builts)
	if err != nil {
		return fmt.Errorf("failed to transform objects: %w", err)
	}

	for _, built := range builts {
		if _, err := kubeutil.Apply(ctx, h.Client, &built, "instance-webhook", true, true); err != nil {
			// ignore NotFound in case the template contains a dependency resource that was not found.
			if !apierrs.IsNotFound(err) {
				return fmt.Errorf("dryrun failed: kind=%s name=%s: %w", built.GetKind(), built.GetName(), err)
			}
		}
	}
	return nil
}
