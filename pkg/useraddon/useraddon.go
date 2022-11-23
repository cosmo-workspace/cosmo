package useraddon

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	ctrl "sigs.k8s.io/controller-runtime"

	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

func EmptyTemplateObject(addon cosmov1alpha1.UserAddon) cosmov1alpha1.TemplateObject {
	if addon.Template.Name == "" {
		return nil
	}
	if addon.Template.ClusterScoped {
		return &cosmov1alpha1.ClusterTemplate{ObjectMeta: v1.ObjectMeta{Name: addon.Template.Name}}
	}
	return &cosmov1alpha1.Template{ObjectMeta: v1.ObjectMeta{Name: addon.Template.Name}}
}

func EmptyInstanceObject(addon cosmov1alpha1.UserAddon, username string) cosmov1alpha1.InstanceObject {
	if addon.Template.Name == "" {
		return nil
	}

	if addon.Template.ClusterScoped {
		return &cosmov1alpha1.ClusterInstance{
			ObjectMeta: v1.ObjectMeta{
				Name: InstanceName(addon.Template.Name, username),
			},
		}
	}
	return &cosmov1alpha1.Instance{
		ObjectMeta: v1.ObjectMeta{
			Name:      InstanceName(addon.Template.Name, ""),
			Namespace: cosmov1alpha1.UserNamespace(username),
		},
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

func PatchUserAddonInstanceAsDesired(inst cosmov1alpha1.InstanceObject, addon cosmov1alpha1.UserAddon, user cosmov1alpha1.User, scheme *runtime.Scheme) error {

	// set label
	label := inst.GetLabels()
	if label == nil {
		label = make(map[string]string)
	}
	label[cosmov1alpha1.TemplateLabelKeyType] = cosmov1alpha1.TemplateLabelEnumTypeUserAddon
	inst.SetLabels(label)

	// set template name
	inst.GetSpec().Template = cosmov1alpha1.TemplateRef{Name: EmptyTemplateObject(addon).GetName()}

	// add default vars
	if addon.Vars == nil {
		addon.Vars = make(map[string]string)
	}
	addon.Vars[template.DefaultVarsNamespace] = cosmov1alpha1.UserNamespace(user.Name)
	addon.Vars[cosmov1alpha1.TemplateVarUserName] = user.Name
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
