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

type CreateUserResponse struct {
	Message string `json:"message"`

	User *User `json:"user"`
}

// AssertCreateUserResponseRequired checks if the required fields are not zero-ed
func AssertCreateUserResponseRequired(obj CreateUserResponse) error {
	elements := map[string]interface{}{
		"message": obj.Message,
		"user":    obj.User,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	if obj.User != nil {
		if err := AssertUserRequired(*obj.User); err != nil {
			return err
		}
	}
	return nil
}

// AssertRecurseCreateUserResponseRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of CreateUserResponse (e.g. [][]CreateUserResponse), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseCreateUserResponseRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aCreateUserResponse, ok := obj.(CreateUserResponse)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertCreateUserResponseRequired(aCreateUserResponse)
	})
}
