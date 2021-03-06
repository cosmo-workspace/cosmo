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

type ListUsersResponse struct {
	Message string `json:"message,omitempty"`

	Items []User `json:"items"`
}

// AssertListUsersResponseRequired checks if the required fields are not zero-ed
func AssertListUsersResponseRequired(obj ListUsersResponse) error {
	elements := map[string]interface{}{
		"items": obj.Items,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.Items {
		if err := AssertUserRequired(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertRecurseListUsersResponseRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of ListUsersResponse (e.g. [][]ListUsersResponse), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseListUsersResponseRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aListUsersResponse, ok := obj.(ListUsersResponse)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertListUsersResponseRequired(aListUsersResponse)
	})
}
