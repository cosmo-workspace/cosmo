package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	"github.com/cosmo-workspace/cosmo/pkg/wscfg"
)

type TemplateMutationWebhookHandler struct {
	Client  kosmo.Client
	Log     *clog.Logger
	decoder *admission.Decoder

	DefaultURLBase string
}

//+kubebuilder:webhook:path=/mutate-cosmo-cosmo-workspace-github-io-v1alpha1-template,mutating=true,failurePolicy=fail,sideEffects=None,groups=cosmo.cosmo-workspace.github.io,resources=templates,verbs=create;update,versions=v1alpha1,name=mtemplate.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *TemplateMutationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/mutate-cosmo-cosmo-workspace-github-io-v1alpha1-template",
		&webhook.Admission{Handler: h},
	)
}

func (h *TemplateMutationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	tmpl := &cosmov1alpha1.Template{}
	err := h.decoder.Decode(req, tmpl)
	if err != nil {
		h.Log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}
	before := tmpl.DeepCopy()
	h.Log.DebugAll().DumpObject(h.Client.Scheme(), before, "request template")

	// mutate the fields in template
	tmplType, _ := template.GetTemplateType(tmpl)
	switch tmplType {
	case wsv1alpha1.TemplateTypeWorkspace:
		cfg, err := wscfg.ConfigFromTemplateAnnotations(tmpl)
		if err != nil {
			h.Log.Error(err, "failed to get workspace config")
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to get workspace config: %w", err))
		}
		if cfg.URLBase == "" {
			cfg.URLBase = h.DefaultURLBase
		}

		wscfg.SetConfigOnTemplateAnnotations(tmpl, cfg)
	}

	h.Log.Debug().PrintObjectDiff(before, tmpl)

	marshaled, err := json.Marshal(tmpl)
	if err != nil {
		h.Log.Error(err, "failed to marshal response")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func (h *TemplateMutationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// type TemplateValidateHandler interface {
// 	Valudate(context.Context, cosmov1alpha1.Template) error
// }

// type TemplateValidationWebhookHandler struct {
// 	Client  kosmo.Client
// 	Log     *clog.Logger
// 	decoder *admission.Decoder

// 	WsTmplValidator TemplateValidateHandler
// }

// //+kubebuilder:webhook:path=/validate-cosmo-cosmo-workspace-github-io-v1alpha1-template,mutating=false,failurePolicy=fail,sideEffects=None,groups=cosmo.cosmo-workspace.github.io,resources=templates,verbs=create;update,versions=v1alpha1,name=vtemplate.kb.io,admissionReviewVersions={v1,v1alpha1}

// func (h *TemplateValidationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
// 	mgr.GetWebhookServer().Register(
// 		"/validate-cosmo-cosmo-workspace-github-io-v1alpha1-template",
// 		&webhook.Admission{Handler: h},
// 	)
// }

// // Handle validates the fields in Template
// func (h *TemplateValidationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
// 	tmpl := &cosmov1alpha1.Template{}
// 	err := h.decoder.Decode(req, tmpl)
// 	if err != nil {
// 		h.Log.Error(err, "failed to decode request")
// 		return admission.Errored(http.StatusBadRequest, err)
// 	}
// 	h.Log.DebugAll().DumpObject(h.Client.Scheme(), tmpl, "request template")

// 	tmplType, _ := template.GetTemplateType(tmpl)
// 	switch tmplType {
// 	case wsv1alpha1.TemplateTypeWorkspace:
// 		err := h.WsTmplValidator.Valudate(ctx, *tmpl)
// 		if err != nil {
// 			h.Log.Error(err, "invalid workspace template")
// 			return admission.Errored(http.StatusForbidden, fmt.Errorf("invalid workspace template: %w", err))
// 		}
// 	}

// 	return admission.Allowed("Validation OK")
// }

// func (h *TemplateValidationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
// 	h.decoder = d
// 	return nil
// }
