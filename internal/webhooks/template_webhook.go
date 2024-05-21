package webhooks

import (
	"context"
	"fmt"
	"net/http"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

type TemplateValidationWebhookHandler struct {
	Client  client.Client
	Log     *clog.Logger
	Decoder admission.Decoder
}

//+kubebuilder:webhook:path=/validate-cosmo-workspace-github-io-v1alpha1-template,mutating=false,failurePolicy=fail,sideEffects=None,groups=cosmo-workspace.github.io,resources=templates,verbs=create;update,versions=v1alpha1,name=vtemplate.kb.io,admissionReviewVersions={v1,v1alpha1}
//+kubebuilder:webhook:path=/validate-cosmo-workspace-github-io-v1alpha1-template,mutating=false,failurePolicy=fail,sideEffects=None,groups=cosmo-workspace.github.io,resources=clustertemplates,verbs=create;update,versions=v1alpha1,name=vclustertemplate.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *TemplateValidationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/validate-cosmo-workspace-github-io-v1alpha1-template",
		&webhook.Admission{Handler: h},
	)
}

// Handle validates the fields in Template
func (h *TemplateValidationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := h.Log.WithValues("UID", req.UID, "GroupVersionKind", req.Kind.String(), "Name", req.Name, "Namespace", req.Namespace)

	switch req.RequestKind.Kind {
	case "Template":
		tmpl := &cosmov1alpha1.Template{}
		err := h.Decoder.Decode(req, tmpl)
		if err != nil {
			log.Error(err, "failed to decode request")
			return admission.Errored(http.StatusBadRequest, err)
		}
		log.DebugAll().DumpObject(h.Client.Scheme(), tmpl, "request template")

		clusterTmpl := &cosmov1alpha1.ClusterTemplate{}
		err = h.Client.Get(ctx, types.NamespacedName{Name: tmpl.Name}, clusterTmpl)
		if err == nil {
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("ClusterTemplate: %s already exists", tmpl.Name))
		} else {
			if !apierrs.IsNotFound(err) {
				return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to get ClusterTemplate: %w", err))
			}
		}

	case "ClusterTemplate":
		clusterTmpl := &cosmov1alpha1.ClusterTemplate{}
		err := h.Decoder.Decode(req, clusterTmpl)
		if err != nil {
			log.Error(err, "failed to decode request")
			return admission.Errored(http.StatusBadRequest, err)
		}
		log.DebugAll().DumpObject(h.Client.Scheme(), clusterTmpl, "request cluster template")

		tmpl := &cosmov1alpha1.Template{}
		err = h.Client.Get(ctx, types.NamespacedName{Name: clusterTmpl.Name}, tmpl)
		if err == nil {
			return admission.Errored(http.StatusBadRequest, fmt.Errorf("Template: %s already exists", clusterTmpl.Name))
		} else {
			if !apierrs.IsNotFound(err) {
				return admission.Errored(http.StatusInternalServerError, fmt.Errorf("failed to get Template: %w", err))
			}
		}

	default:
		err := fmt.Errorf("invalid kind: %v", req.RequestKind)
		log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.Allowed("Validation OK")
}
