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

type UserAddons struct {
	Template string `json:"template"`

	Vars map[string]string `json:"vars,omitempty"`
}

// AssertUserAddonsRequired checks if the required fields are not zero-ed
func AssertUserAddonsRequired(obj UserAddons) error {
	elements := map[string]interface{}{
		"template": obj.Template,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertRecurseUserAddonsRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of UserAddons (e.g. [][]UserAddons), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseUserAddonsRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aUserAddons, ok := obj.(UserAddons)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertUserAddonsRequired(aUserAddons)
	})
}
