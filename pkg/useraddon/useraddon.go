package useraddon

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	ctrl "sigs.k8s.io/controller-runtime"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/core/v1alpha1"
	wsv1alpha1 "github.com/cosmo-workspace/cosmo/api/workspace/v1alpha1"
)

func EmptyTemplateObject(addon wsv1alpha1.UserAddon) cosmov1alpha1.TemplateObject {
	if addon.Template.Name != "" {
		return &cosmov1alpha1.Template{ObjectMeta: v1.ObjectMeta{Name: addon.Template.Name}}
	} else if addon.ClusterTemplate.Name != "" {
		return &cosmov1alpha1.ClusterTemplate{ObjectMeta: v1.ObjectMeta{Name: addon.ClusterTemplate.Name}}
	} else {
		return nil
	}
}

func EmptyInstanceObject(addon wsv1alpha1.UserAddon, userid string) cosmov1alpha1.InstanceObject {
	if addon.Template.Name != "" {
		return &cosmov1alpha1.Instance{
			ObjectMeta: v1.ObjectMeta{
				Name:      InstanceName(addon.Template.Name, ""),
				Namespace: wsv1alpha1.UserNamespace(userid),
			},
		}
	} else if addon.ClusterTemplate.Name != "" {
		return &cosmov1alpha1.ClusterInstance{
			ObjectMeta: v1.ObjectMeta{
				Name: InstanceName(addon.ClusterTemplate.Name, userid),
			},
		}
	} else {
		return nil
	}
}

func InstanceName(addonTmplName, userName string) (name string) {
	if userName != "" {
		name = fmt.Sprintf("useraddon-%s-%s", userName, addonTmplName)
	} else {
		name = fmt.Sprintf("useraddon-%s", addonTmplName)
	}

	// truncate name
	if len(name) > validation.DNS1123LabelMaxLength {
		return name[:validation.DNS1123LabelMaxLength]
	}
	return name
}

func PatchUserAddonInstanceAsDesired(inst cosmov1alpha1.InstanceObject, addon wsv1alpha1.UserAddon, user wsv1alpha1.User, scheme *runtime.Scheme) error {

	// set label
	label := inst.GetLabels()
	if label == nil {
		label = make(map[string]string)
	}
	label[cosmov1alpha1.TemplateLabelKeyType] = wsv1alpha1.TemplateTypeUserAddon
	inst.SetLabels(label)

	// set template name
	inst.GetSpec().Template = cosmov1alpha1.TemplateRef{Name: EmptyTemplateObject(addon).GetName()}

	// add default vars
	if addon.Vars == nil {
		addon.Vars = make(map[string]string)
	}
	addon.Vars[wsv1alpha1.TemplateVarUserNamespace] = wsv1alpha1.UserNamespace(user.Name)
	addon.Vars[wsv1alpha1.TemplateVarUserID] = user.Name
	inst.GetSpec().Vars = addon.Vars

	// set owner reference if scheme is not nil
	if scheme != nil {
		err := ctrl.SetControllerReference(&user, inst, scheme)
		if err != nil {
			return fmt.Errorf("failed to set controller reference: %w", err)
		}
	}

	return nil
}
