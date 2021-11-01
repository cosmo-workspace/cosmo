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

type User struct {
	Id string `json:"id"`

	DisplayName string `json:"displayName,omitempty"`

	Role string `json:"role,omitempty"`

	AuthType string `json:"authType,omitempty"`

	Addons []ApiV1alpha1UserAddons `json:"addons,omitempty"`

	DefaultPassword string `json:"defaultPassword,omitempty"`

	Status string `json:"status,omitempty"`
}

// AssertUserRequired checks if the required fields are not zero-ed
func AssertUserRequired(obj User) error {
	elements := map[string]interface{}{
		"id": obj.Id,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.Addons {
		if err := AssertApiV1alpha1UserAddonsRequired(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertRecurseUserRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of User (e.g. [][]User), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseUserRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aUser, ok := obj.(User)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertUserRequired(aUser)
	})
}
