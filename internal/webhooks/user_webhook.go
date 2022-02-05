package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kosmo"
)

type UserMutationWebhookHandler struct {
	Client  kosmo.Client
	Log     *clog.Logger
	decoder *admission.Decoder
}

//+kubebuilder:webhook:path=/mutate-workspace-cosmo-workspace-github-io-v1alpha1-user,mutating=true,failurePolicy=fail,sideEffects=None,groups=workspace.cosmo-workspace.github.io,resources=users,verbs=create;update,versions=v1alpha1,name=muser.kb.io,admissionReviewVersions={v1,v1beta1}

func (h *UserMutationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/mutate-workspace-cosmo-workspace-github-io-v1alpha1-user",
		&webhook.Admission{Handler: h},
	)
}

// Handle mutates the fields in user
func (h *UserMutationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	user := &wsv1alpha1.User{}
	err := h.decoder.Decode(req, user)
	if err != nil {
		h.Log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}
	before := user.DeepCopy()
	h.Log.DebugAll().DumpObject(h.Client.Scheme(), before, "request user")

	addonTmpls, err := h.Client.ListTemplatesByType(ctx, []string{wsv1alpha1.TemplateTypeUserAddon})
	if err != nil {
		h.Log.Error(err, "failed to list templates")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// defaulting auth type
	if user.Spec.AuthType == "" {
		user.Spec.AuthType = wsv1alpha1.UserAuthTypeKosmoSecert
	}

	// add default user addon
	for _, v := range addonTmpls {
		h.Log.DebugAll().Info("user addon template", "name", v.Name)

		ann := v.GetAnnotations()
		if ann == nil {
			continue
		}
		val, ok := ann[wsv1alpha1.TemplateAnnKeyDefaultUserAddon]
		if !ok {
			continue
		}
		isDefaultUserAddon, err := strconv.ParseBool(val)
		if err != nil {
			h.Log.Error(err, "failed to parse default-user-addon annotation value: %s: %w", val, err)
			continue
		}
		h.Log.Debug().Info("defaulting user addon", "name", v.Name)

		if isDefaultUserAddon {
			addon := wsv1alpha1.UserAddon{Template: cosmov1alpha1.TemplateRef{Name: v.GetName()}}

			var found bool
			for _, v := range user.Spec.Addons {
				if v.Template.Name == addon.Template.Name {
					h.Log.Info("default addon is already defined", "user", user.Name, "addon", addon.Template.Name)
					found = true
				}
			}
			if !found {
				h.Log.Info("appended default addon", "user", user.Name, "addon", addon.Template.Name)
				if len(user.Spec.Addons) == 0 {
					user.Spec.Addons = []wsv1alpha1.UserAddon{addon}
				} else {
					user.Spec.Addons = append(user.Spec.Addons, addon)
				}
			}
		}
	}

	h.Log.Debug().PrintObjectDiff(before, user)

	marshaled, err := json.Marshal(user)
	if err != nil {
		h.Log.Error(err, "failed to marshal resoponse")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func (h *UserMutationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

type UserValidationWebhookHandler struct {
	Client  kosmo.Client
	Log     *clog.Logger
	decoder *admission.Decoder
}

//+kubebuilder:webhook:path=/validate-workspace-cosmo-workspace-github-io-v1alpha1-user,mutating=false,failurePolicy=fail,sideEffects=None,groups=workspace.cosmo-workspace.github.io,resources=users,verbs=create;update,versions=v1alpha1,name=vuser.kb.io,admissionReviewVersions={v1,v1beta1}

func (h *UserValidationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/validate-workspace-cosmo-workspace-github-io-v1alpha1-user",
		&webhook.Admission{Handler: h},
	)
}

// Handle validates the fields in User
func (h *UserValidationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	user := &wsv1alpha1.User{}
	err := h.decoder.Decode(req, user)
	if err != nil {
		h.Log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}
	h.Log.DebugAll().DumpObject(h.Client.Scheme(), user, "request user")

	// check user name is valid for namespace
	if !validName(user.Name) {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("metadata.name: Invalid value: '%s': a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')", user.Name))
	}

	// check role is valid
	if !user.Spec.Role.IsValid() {
		h.Log.Info("invalid user role", "user", user.Name, "role", user.Spec.Role)
		return admission.Denied("invalid user role")
	}

	// check auth type is valid
	if !user.Spec.AuthType.IsValid() {
		h.Log.Info("invalid auth type", "user", user.Name, "authType", user.Spec.AuthType)
		return admission.Denied("invalid auth type")
	}

	// check addon template is labeled as user-addon
	if len(user.Spec.Addons) > 0 {
		for _, addon := range user.Spec.Addons {
			tmpl, err := h.Client.GetTemplate(ctx, addon.Template.Name)
			if err != nil {
				h.Log.Error(err, "failed to create addon", "user", user.Name, "addon", addon.Template.Name)
				return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to create addon %s :%v", addon.Template.Name, err))
			}

			label := tmpl.GetLabels()
			if label == nil {
				h.Log.Info("template is not labeled as user-addon", "user", user.Name, "addon", addon.Template.Name)
				return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to create addon %s: template is not labeled as user-addon", addon.Template.Name))
			}

			if t, ok := label[cosmov1alpha1.TemplateLabelKeyType]; !ok || t != wsv1alpha1.TemplateTypeUserAddon {
				h.Log.Info("template is not labeled as user-addon", "user", user.Name, "addon", addon.Template.Name)
				return admission.Errored(http.StatusBadRequest, fmt.Errorf("failed to create addon %s: template is not labeled as user-addon", addon.Template.Name))
			}
		}
	}

	return admission.Allowed("Validation OK")
}

func (h *UserValidationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func validName(v string) bool {
	r, _ := regexp.Compile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	return r.MatchString(v)
}
