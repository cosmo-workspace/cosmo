package kosmo

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	apierrs "k8s.io/apimachinery/pkg/api/errors"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/cosmo-workspace/cosmo/pkg/kubeutil"
)

func FilterTemplates(ctx context.Context, tmpls []cosmov1alpha1.TemplateObject, u *cosmov1alpha1.User) []cosmov1alpha1.TemplateObject {
	filteredTmpls := make([]cosmov1alpha1.TemplateObject, 0, len(tmpls))
	for _, v := range tmpls {
		if IsAllowedToUseTemplate(ctx, u, v) {
			filteredTmpls = append(filteredTmpls, v)
		}
	}
	return filteredTmpls
}

func IsAllowedToUseTemplate(ctx context.Context, u *cosmov1alpha1.User, tmpl cosmov1alpha1.TemplateObject) bool {
	debugAll := clog.FromContext(ctx).DebugAll()

	ann := tmpl.GetAnnotations()
	if ann == nil || cosmov1alpha1.HasPrivilegedRole(u.Spec.Roles) {
		// all allowed
		debugAll.Info("all allowed", "tmpl", tmpl.GetName())
		return true
	}

	forRoles := ann[cosmov1alpha1.TemplateAnnKeyUserRoles]
	if forRoles == "" {
		// all allowed
		debugAll.Info("allowed: roles does not matched all forbiddenRoles and NO forRoles", "forRoles", forRoles, "tmpl", tmpl.GetName())
		return true
	}
	for _, forRole := range strings.Split(forRoles, ",") {
		for _, role := range u.Spec.Roles {
			debugAll.Info("matching to forRole...", "forRoles", forRoles, "role", role.Name, "tmpl", tmpl.GetName())
			if matched, err := filepath.Match(forRole, role.Name); err == nil && matched {
				debugAll.Info("allowed: roles matched to forRole", "forRoles", forRoles, "role", role.Name, "tmpl", tmpl.GetName())
				return true
			}
		}
	}
	// the role does not match the specified roles
	debugAll.Info("forbidden: roles does not match forRoles", forRoles)
	return false
}

func HasRequiredAddons(ctx context.Context, u *cosmov1alpha1.User, tmpl cosmov1alpha1.TemplateObject) bool {
	debugAll := clog.FromContext(ctx).DebugAll()

	reqAddons := kubeutil.GetAnnotation(tmpl, cosmov1alpha1.TemplateAnnKeyRequiredAddons)
	if reqAddons == "" {
		return true
	}
	for _, requiredAddon := range strings.Split(reqAddons, ",") {
		for _, addon := range u.Spec.Addons {
			if requiredAddon == addon.Template.Name {
				return true
			}
		}
	}
	debugAll.Info("user does not have required addon for template", "requiredAddons", reqAddons)
	return false
}

func (c *Client) ListWorkspaceTemplates(ctx context.Context) ([]cosmov1alpha1.TemplateObject, error) {
	log := clog.FromContext(ctx).WithCaller()
	if tmpls, err := kubeutil.ListTemplateObjectsByType(ctx, c, []string{cosmov1alpha1.TemplateLabelEnumTypeWorkspace}); err != nil {
		log.Error(err, "failed to list WorkspaceTemplates")
		return nil, apierrs.NewInternalError(fmt.Errorf("failed to list WorkspaceTemplates: %w", err))
	} else {
		return tmpls, nil
	}
}

func (c *Client) ListUserAddonTemplates(ctx context.Context) ([]cosmov1alpha1.TemplateObject, error) {
	log := clog.FromContext(ctx).WithCaller()
	if tmpls, err := kubeutil.ListTemplateObjectsByType(ctx, c, []string{cosmov1alpha1.TemplateLabelEnumTypeUserAddon}); err != nil {
		log.Error(err, "failed to list UserAddon Templates")
		return nil, apierrs.NewInternalError(fmt.Errorf("failed to list UserAddon Templates: %w", err))
	} else {
		return tmpls, nil
	}
}
