package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/useraddon"
)

type UserMutationWebhookHandler struct {
	Client  client.Client
	Log     *clog.Logger
	decoder *admission.Decoder
}

//+kubebuilder:webhook:path=/mutate-cosmo-workspace-github-io-v1alpha1-user,mutating=true,failurePolicy=fail,sideEffects=None,groups=cosmo-workspace.github.io,resources=users,verbs=create;update,versions=v1alpha1,name=muser.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *UserMutationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/mutate-cosmo-workspace-github-io-v1alpha1-user",
		&webhook.Admission{Handler: h},
	)
}

// Handle mutates the fields in user
func (h *UserMutationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := h.Log.WithValues("UID", req.UID, "GroupVersionKind", req.Kind.String(), "Name", req.Name, "Namespace", req.Namespace)
	ctx = clog.IntoContext(ctx, log)

	user := &cosmov1alpha1.User{}
	err := h.decoder.Decode(req, user)
	if err != nil {
		log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}
	before := user.DeepCopy()
	log.DumpObject(h.Client.Scheme(), before, "request user")

	addonTmpls, err := kubeutil.ListTemplateObjectsByType(ctx, h.Client, []string{cosmov1alpha1.TemplateLabelEnumTypeUserAddon})
	if err != nil {
		log.Error(err, "failed to list templates")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// defaulting auth type
	if user.Spec.AuthType == "" {
		user.Spec.AuthType = cosmov1alpha1.UserAuthTypePasswordSecert
	}

	// add default user addon
	for _, addonTmpl := range addonTmpls {
		ann := addonTmpl.GetAnnotations()
		if ann == nil {
			continue
		}
		val, ok := ann[cosmov1alpha1.UserAddonTemplateAnnKeyDefaultUserAddon]
		if !ok {
			continue
		}
		isDefaultUserAddon, err := strconv.ParseBool(val)
		if err != nil {
			log.Error(err, "failed to parse default-user-addon annotation value: %s: %w", val, err)
			continue
		}
		log.Debug().Info("defaulting user addon", "name", addonTmpl.GetName())

		if isDefaultUserAddon {
			var defaultAddon cosmov1alpha1.UserAddon
			defaultAddon.Template.Name = addonTmpl.GetName()
			defaultAddon.Template.ClusterScoped = addonTmpl.GetScope() == meta.RESTScopeRoot

			var found bool
			for _, v := range user.Spec.Addons {
				if reflect.DeepEqual(v.Template, defaultAddon.Template) {
					found = true
				}
			}

			if !found {
				log.Info("appended default addon", "user", user.Name, "addon", defaultAddon)
				if len(user.Spec.Addons) == 0 {
					user.Spec.Addons = []cosmov1alpha1.UserAddon{defaultAddon}
				} else {
					user.Spec.Addons = append(user.Spec.Addons, defaultAddon)
				}
			} else {
				log.Info("default addon is already defined", "user", user.Name, "addon", defaultAddon)
			}
		}
	}

	log.Debug().PrintObjectDiff(before, user)

	marshaled, err := json.Marshal(user)
	if err != nil {
		log.Error(err, "failed to marshal resoponse")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func (h *UserMutationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

type UserValidationWebhookHandler struct {
	Client  client.Client
	Log     *clog.Logger
	decoder *admission.Decoder
}

//+kubebuilder:webhook:path=/validate-cosmo-workspace-github-io-v1alpha1-user,mutating=false,failurePolicy=fail,sideEffects=None,groups=cosmo-workspace.github.io,resources=users,verbs=create;update,versions=v1alpha1,name=vuser.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *UserValidationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/validate-cosmo-workspace-github-io-v1alpha1-user",
		&webhook.Admission{Handler: h},
	)
}

// Handle validates the fields in User
func (h *UserValidationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := h.Log.WithValues("UID", req.UID, "GroupVersionKind", req.Kind.String(), "Name", req.Name, "Namespace", req.Namespace)
	ctx = clog.IntoContext(ctx, log)

	user := &cosmov1alpha1.User{}
	err := h.decoder.Decode(req, user)
	if err != nil {
		log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.DumpObject(h.Client.Scheme(), user, "request user")

	// check user name is valid for namespace
	if !validName(user.Name) {
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("metadata.name: Invalid value: '%s': a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')", user.Name))
	}

	// check auth type is valid
	if !user.Spec.AuthType.IsValid() {
		log.Info("invalid auth type", "user", user.Name, "authType", user.Spec.AuthType)
		return admission.Denied("invalid auth type")
	}

	// check addon template is labeled as useraddon
	if len(user.Spec.Addons) > 0 {
		for _, addon := range user.Spec.Addons {
			tmpl := useraddon.EmptyTemplateObject(addon)
			if tmpl == nil {
				continue
			}

			err = h.Client.Get(ctx, types.NamespacedName{Name: tmpl.GetName()}, tmpl)
			if err != nil {
				log.Error(err, "failed to create addon", "user", user.Name, "addon", tmpl.GetName())
				return admission.Denied(fmt.Sprintf("failed to create addon %s :%v", tmpl.GetName(), err))
			}

			// check label
			label := tmpl.GetLabels()
			if label == nil {
				log.Info("template is not labeled as useraddon", "user", user.Name, "addon", tmpl.GetName())
				return admission.Denied(fmt.Sprintf("failed to create addon %s: template is not labeled as useraddon", tmpl.GetName()))
			}
			if t, ok := label[cosmov1alpha1.TemplateLabelKeyType]; !ok || t != cosmov1alpha1.TemplateLabelEnumTypeUserAddon {
				log.Info("template is not labeled as useraddon", "user", user.Name, "addon", tmpl.GetName())
				return admission.Denied(fmt.Sprintf("failed to create addon %s: template is not labeled as useraddon", tmpl.GetName()))
			}

			// TODO
			// // dryrun create or update addon
			// inst := useraddon.EmptyInstanceObject(addon, user.GetName())
			// if _, err := kubeutil.DryrunCreateOrUpdate(ctx, h.Client, inst, func() error {
			// 	return useraddon.PatchUserAddonInstanceAsDesired(inst, addon, *user, nil)
			// }); err != nil {
			// 	return admission.Denied(fmt.Sprintf("failed to create or update addon %v", err))
			// }
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
