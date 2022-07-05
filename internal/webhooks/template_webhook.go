package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	"github.com/cosmo-workspace/cosmo/pkg/wscfg"
)

type TemplateMutationWebhookHandler struct {
	Client  client.Client
	Log     *clog.Logger
	decoder *admission.Decoder

	DefaultURLBase string
}

//+kubebuilder:webhook:path=/mutate-cosmo-cosmo-workspace-github-io-v1alpha1-template,mutating=true,failurePolicy=fail,sideEffects=None,groups=cosmo.cosmo-workspace.github.io,resources=templates,verbs=create;update,versions=v1alpha1,name=mtemplate.kb.io,admissionReviewVersions={v1,v1alpha1}
//+kubebuilder:webhook:path=/mutate-cosmo-cosmo-workspace-github-io-v1alpha1-template,mutating=true,failurePolicy=fail,sideEffects=None,groups=cosmo.cosmo-workspace.github.io,resources=clustertemplates,verbs=create;update,versions=v1alpha1,name=mclustertemplate.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *TemplateMutationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/mutate-cosmo-cosmo-workspace-github-io-v1alpha1-template",
		&webhook.Admission{Handler: h},
	)
}

func (h *TemplateMutationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := h.Log.WithValues("UID", req.UID, "GroupVersionKind", req.Kind.String(), "Name", req.Name, "Namespace", req.Namespace)

	var tmpl cosmov1alpha1.TemplateObject

	switch req.RequestKind.Kind {
	case "Template":
		tmpl = &cosmov1alpha1.Template{}
		err := h.decoder.Decode(req, tmpl)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		log.DebugAll().DumpObject(h.Client.Scheme(), tmpl, "request template")

	case "ClusterTemplate":
		tmpl = &cosmov1alpha1.ClusterTemplate{}
		err := h.decoder.Decode(req, tmpl)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		log.DebugAll().DumpObject(h.Client.Scheme(), tmpl, "request cluster template")

	default:
		err := fmt.Errorf("invalid kind: %v", req.RequestKind)
		log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}

	before := tmpl.DeepCopyObject().(cosmov1alpha1.TemplateObject)

	// mutate the fields in template
	tmplType, _ := template.GetTemplateType(tmpl)

	switch tmplType {
	case wsv1alpha1.TemplateTypeWorkspace:
		t, ok := tmpl.(*cosmov1alpha1.Template)
		if ok {
			cfg, err := wscfg.ConfigFromTemplateAnnotations(t)
			if err != nil {
				log.Error(err, "failed to get workspace config")
				return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to get workspace config: %w", err))
			}
			if cfg.URLBase == "" {
				cfg.URLBase = h.DefaultURLBase
			}

			wscfg.SetConfigOnTemplateAnnotations(t, cfg)
		}
	}

	log.Debug().PrintObjectDiff(before, tmpl)

	marshaled, err := json.Marshal(tmpl)
	if err != nil {
		log.Error(err, "failed to marshal response")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func (h *TemplateMutationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

type TemplateValidationWebhookHandler struct {
	Client  client.Client
	Log     *clog.Logger
	decoder *admission.Decoder

	FieldManager string
}

//+kubebuilder:webhook:path=/validate-cosmo-cosmo-workspace-github-io-v1alpha1-template,mutating=false,failurePolicy=fail,sideEffects=None,groups=cosmo.cosmo-workspace.github.io,resources=templates,verbs=create;update,versions=v1alpha1,name=vtemplate.kb.io,admissionReviewVersions={v1,v1alpha1}
//+kubebuilder:webhook:path=/validate-cosmo-cosmo-workspace-github-io-v1alpha1-template,mutating=false,failurePolicy=fail,sideEffects=None,groups=cosmo.cosmo-workspace.github.io,resources=clustertemplates,verbs=create;update,versions=v1alpha1,name=vclustertemplate.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *TemplateValidationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/validate-cosmo-cosmo-workspace-github-io-v1alpha1-template",
		&webhook.Admission{Handler: h},
	)
}

// Handle validates the fields in Template
func (h *TemplateValidationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := h.Log.WithValues("UID", req.UID, "GroupVersionKind", req.Kind.String(), "Name", req.Name, "Namespace", req.Namespace)

	var tmpl cosmov1alpha1.TemplateObject
	var dummyInst cosmov1alpha1.InstanceObject

	switch req.RequestKind.Kind {
	case "Template":
		tmpl = &cosmov1alpha1.Template{}
		err := h.decoder.Decode(req, tmpl)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		log.DebugAll().DumpObject(h.Client.Scheme(), tmpl, "request template")

		dummyInst = &cosmov1alpha1.Instance{}
		dummyInst.SetName("dummy")
		dummyInst.SetNamespace("default")

	case "ClusterTemplate":
		tmpl = &cosmov1alpha1.ClusterTemplate{}
		err := h.decoder.Decode(req, tmpl)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		log.DebugAll().DumpObject(h.Client.Scheme(), tmpl, "request cluster template")

		dummyInst = &cosmov1alpha1.ClusterInstance{}
		dummyInst.SetName("dummy")

	default:
		err := fmt.Errorf("invalid kind: %v", req.RequestKind)
		log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}

	warnings := make([]string, 0)
	if !template.IsSkipValidation(tmpl) {
		// dryrun apply
		if errs := dryrunReconcile(ctx, h.Client, h.FieldManager, dummyInst, tmpl); len(errs) > 0 {
			for _, err := range errs {
				warnings = append(warnings, err.Error())
			}
		}
	} else {
		h.Log.Info("skip dryrun validation")
	}

	res := admission.Allowed("Validation OK")
	if len(warnings) > 0 {
		res.Warnings = warnings
	}
	return res
}

func (h *TemplateValidationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}
