package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/instance"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
	"github.com/cosmo-workspace/cosmo/pkg/template"
	"github.com/cosmo-workspace/cosmo/pkg/transformer"
)

type InstanceMutationWebhookHandler struct {
	Client  client.Client
	Log     *clog.Logger
	decoder *admission.Decoder
}

//+kubebuilder:webhook:path=/mutate-cosmo-workspace-github-io-v1alpha1-instance,mutating=true,failurePolicy=fail,sideEffects=None,groups=cosmo-workspace.github.io,resources=instances,verbs=create;update,versions=v1alpha1,name=minstance.kb.io,admissionReviewVersions={v1,v1alpha1}
//+kubebuilder:webhook:path=/mutate-cosmo-workspace-github-io-v1alpha1-instance,mutating=true,failurePolicy=fail,sideEffects=None,groups=cosmo-workspace.github.io,resources=clusterinstances,verbs=create;update,versions=v1alpha1,name=mclusterinstance.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *InstanceMutationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/mutate-cosmo-workspace-github-io-v1alpha1-instance",
		&webhook.Admission{Handler: h},
	)
}

func (h *InstanceMutationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := h.Log.WithValues("UID", req.UID, "GroupVersionKind", req.Kind.String(), "Name", req.Name, "Namespace", req.Namespace)
	ctx = clog.IntoContext(ctx, log)

	var inst cosmov1alpha1.InstanceObject
	var tmpl cosmov1alpha1.TemplateObject

	switch req.RequestKind.Kind {
	case "Instance":
		inst = &cosmov1alpha1.Instance{}
		err := h.decoder.Decode(req, inst)
		if err != nil {
			log.Error(err, "failed to decode request")
			return admission.Errored(http.StatusBadRequest, err)
		}
		log.DumpObject(h.Client.Scheme(), inst, "request instance")

		tmpl = &cosmov1alpha1.Template{}
		err = h.Client.Get(ctx, types.NamespacedName{Name: inst.GetSpec().Template.Name}, tmpl)
		if err != nil {
			log.Error(err, "failed to get template")
			return admission.Errored(http.StatusBadRequest, err)
		}

	case "ClusterInstance":
		inst = &cosmov1alpha1.ClusterInstance{}
		err := h.decoder.Decode(req, inst)
		if err != nil {
			log.Error(err, "failed to decode request")
			return admission.Errored(http.StatusBadRequest, err)
		}
		log.DumpObject(h.Client.Scheme(), inst, "request cluster instance")

		tmpl = &cosmov1alpha1.ClusterTemplate{}
		err = h.Client.Get(ctx, types.NamespacedName{Name: inst.GetSpec().Template.Name}, tmpl)
		if err != nil {
			log.Error(err, "failed to get cluster template")
			return admission.Errored(http.StatusBadRequest, err)
		}

	default:
		err := fmt.Errorf("invalid kind: %v", req.RequestKind)
		log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}

	before := inst.DeepCopyObject().(cosmov1alpha1.InstanceObject)

	mutateInstanceObject(inst, tmpl)

	log.Debug().PrintObjectDiff(before, inst)

	marshaled, err := json.Marshal(inst)
	if err != nil {
		log.Error(err, "failed to marshal response")
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaled)
}

func mutateInstanceObject(inst cosmov1alpha1.InstanceObject, tmpl cosmov1alpha1.TemplateObject) {
	instSpec := inst.GetSpec()
	tmplSpec := tmpl.GetSpec()

	// mutate the fields in instance
	// propagate template type annotation to instance annotation
	if tmplType, ok := template.GetTemplateType(tmpl); ok {
		template.SetTemplateType(inst, tmplType)
	}

	// defaulting required vars
	for _, v := range tmplSpec.RequiredVars {
		found := false
		for key := range instSpec.Vars {
			if template.FixupTemplateVarKey(key) == template.FixupTemplateVarKey(v.Var) {
				found = true
			}
		}
		if !found && v.Default != "" {
			if instSpec.Vars == nil {
				instSpec.Vars = make(map[string]string)
			}
			instSpec.Vars[v.Var] = v.Default
		}
	}

	// update name to instance fixed resource name
	patchSpec := instSpec.Override.PatchesJson6902
	for i, p := range patchSpec {
		if p.Target.Name != "" && !template.IsDisableNamePrefix(tmpl) {
			patchSpec[i].Target.Name = instance.InstanceResourceName(inst.GetName(), p.Target.Name)
		}
	}
}

func (h *InstanceMutationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

type InstanceValidationWebhookHandler struct {
	Client  client.Client
	Log     *clog.Logger
	decoder *admission.Decoder

	FieldManager string
}

//+kubebuilder:webhook:path=/validate-cosmo-workspace-github-io-v1alpha1-instance,mutating=false,failurePolicy=fail,sideEffects=None,groups=cosmo-workspace.github.io,resources=clusterinstances,verbs=create;update,versions=v1alpha1,name=vclusterinstance.kb.io,admissionReviewVersions={v1,v1alpha1}
//+kubebuilder:webhook:path=/validate-cosmo-workspace-github-io-v1alpha1-instance,mutating=false,failurePolicy=fail,sideEffects=None,groups=cosmo-workspace.github.io,resources=instances,verbs=create;update,versions=v1alpha1,name=vinstance.kb.io,admissionReviewVersions={v1,v1alpha1}

func (h *InstanceValidationWebhookHandler) SetupWebhookWithManager(mgr ctrl.Manager) {
	mgr.GetWebhookServer().Register(
		"/validate-cosmo-workspace-github-io-v1alpha1-instance",
		&webhook.Admission{Handler: h},
	)
}

func (h *InstanceValidationWebhookHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := h.Log.WithValues("UID", req.UID, "GroupVersionKind", req.Kind.String(), "Name", req.Name, "Namespace", req.Namespace)
	ctx = clog.IntoContext(ctx, log)

	var inst cosmov1alpha1.InstanceObject
	var tmpl cosmov1alpha1.TemplateObject

	switch req.RequestKind.Kind {
	case "Instance":
		inst = &cosmov1alpha1.Instance{}
		err := h.decoder.Decode(req, inst)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		log.DumpObject(h.Client.Scheme(), inst, "request instance")

		// whether template exist
		tmpl = &cosmov1alpha1.Template{}
		err = h.Client.Get(ctx, types.NamespacedName{Name: inst.GetSpec().Template.Name}, tmpl)
		if err != nil {
			log.Error(err, "failed to get template")
			return admission.Errored(http.StatusBadRequest, err)
		}

	case "ClusterInstance":
		inst = &cosmov1alpha1.ClusterInstance{}
		err := h.decoder.Decode(req, inst)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		log.DumpObject(h.Client.Scheme(), inst, "request cluster instance")

		// whether template exist
		tmpl = &cosmov1alpha1.ClusterTemplate{}
		err = h.Client.Get(ctx, types.NamespacedName{Name: inst.GetSpec().Template.Name}, tmpl)
		if err != nil {
			log.Error(err, "failed to get clustertemplate")
			return admission.Errored(http.StatusBadRequest, err)
		}

	default:
		err := fmt.Errorf("invalid kind: %v", req.RequestKind)
		log.Error(err, "failed to decode request")
		return admission.Errored(http.StatusBadRequest, err)
	}

	// validate if satisfy template's required vars
	for _, v := range tmpl.GetSpec().RequiredVars {
		ok := false
		for key := range inst.GetSpec().Vars {
			if template.FixupTemplateVarKey(key) == template.FixupTemplateVarKey(v.Var) {
				ok = true
				break
			}
		}
		if !ok {
			return admission.Denied(fmt.Sprintf("Insufficient vars override: Var %v is required", v.Var))
		}
	}

	// validate patch
	patchSpecs := inst.GetSpec().Override.PatchesJson6902
	for _, patchSpec := range patchSpecs {
		if _, err := schema.ParseGroupVersion(patchSpec.Target.APIVersion); err != nil {
			return admission.Denied(fmt.Sprintf("APIVersion '%s' is invalid: %v", patchSpec.Target.APIVersion, err))
		}
		if _, err := json.Marshal(patchSpec.Patch); err != nil {
			return admission.Denied(fmt.Sprintf("JSON Patch format is invalid: %v", err))
		}
	}

	// dryrun apply
	if errs := dryrunReconcile(ctx, h.Client, h.FieldManager, inst, tmpl); len(errs) > 0 {
		msg := make([]string, len(errs))
		for i := range errs {
			msg[i] = errs[i].Error()
		}
		return admission.Denied(strings.Join(msg, ": "))
	}

	return admission.Allowed("OK")
}

func (h *InstanceValidationWebhookHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func dryrunReconcile(ctx context.Context, c client.Client, fieldManager string, inst cosmov1alpha1.InstanceObject, tmpl cosmov1alpha1.TemplateObject) []error {
	log := clog.FromContext(ctx).WithCaller()

	objects, err := template.BuildObjects(*tmpl.GetSpec(), inst)
	if err != nil {
		return []error{err}
	}

	objects, err = transformer.ApplyTransformers(ctx, transformer.AllTransformers(inst, c.Scheme(), tmpl), objects)
	if err != nil {
		return []error{fmt.Errorf("failed to transform objects: %w", err)}
	}

	errs := make([]error, 0)
	for _, built := range objects {
		// in webhook, ownerReference should not be set because the error occuerd
		// err -> metadata.ownerReferences.uid: Invalid value: "": uid must not be empty
		built.SetOwnerReferences(nil)

		log.Debug().Info(fmt.Sprintf("Validate instance's object by dry-run applying... %v\n", built))
		out, err := kubeutil.Apply(ctx, c, &built, fieldManager, true, true)
		log.DebugAll().Info(fmt.Sprintf("Applied object dump %v\n", out))

		if err != nil {
			// ignore NotFound in case the template contains a dependency resource that was not found.
			if !apierrs.IsNotFound(err) {
				errs = append(errs, fmt.Errorf("dryrun failed: kind=%s name=%s: %w", built.GetKind(), built.GetName(), err))
			}
		}
	}
	return errs
}
