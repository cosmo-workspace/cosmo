package instance

import (
	cosmov1alpha1 "github.com/cosmo-workspace/cosmo/api/v1alpha1"
	"github.com/cosmo-workspace/cosmo/pkg/template"
)

func Mutate(inst cosmov1alpha1.InstanceObject, tmpl cosmov1alpha1.TemplateObject) {
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
			patchSpec[i].Target.Name = InstanceResourceName(inst.GetName(), p.Target.Name)
		}
	}
}
