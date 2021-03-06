/*
 * Cosmo Dashboard API
 *
 * Manipulate cosmo dashboard resource API
 *
 * API version: v1alpha1
 * Contact: jlandowner8@gmail.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package v1alpha1

type TemplateRequiredVars struct {
	VarName string `json:"varName"`

	DefaultValue string `json:"defaultValue,omitempty"`
}

// AssertTemplateRequiredVarsRequired checks if the required fields are not zero-ed
func AssertTemplateRequiredVarsRequired(obj TemplateRequiredVars) error {
	elements := map[string]interface{}{
		"varName": obj.VarName,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertRecurseTemplateRequiredVarsRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of TemplateRequiredVars (e.g. [][]TemplateRequiredVars), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseTemplateRequiredVarsRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aTemplateRequiredVars, ok := obj.(TemplateRequiredVars)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertTemplateRequiredVarsRequired(aTemplateRequiredVars)
	})
}
