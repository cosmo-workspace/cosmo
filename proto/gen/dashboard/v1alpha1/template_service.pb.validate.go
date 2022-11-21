// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: dashboard/v1alpha1/template_service.proto

package dashboardv1alpha1

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on GetUserAddonTemplatesResponse with the
// rules defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetUserAddonTemplatesResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetUserAddonTemplatesResponse with
// the rules defined in the proto definition for this message. If any rules
// are violated, the result is a list of violation errors wrapped in
// GetUserAddonTemplatesResponseMultiError, or nil if none found.
func (m *GetUserAddonTemplatesResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *GetUserAddonTemplatesResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Message

	for idx, item := range m.GetItems() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, GetUserAddonTemplatesResponseValidationError{
						field:  fmt.Sprintf("Items[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, GetUserAddonTemplatesResponseValidationError{
						field:  fmt.Sprintf("Items[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GetUserAddonTemplatesResponseValidationError{
					field:  fmt.Sprintf("Items[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return GetUserAddonTemplatesResponseMultiError(errors)
	}

	return nil
}

// GetUserAddonTemplatesResponseMultiError is an error wrapping multiple
// validation errors returned by GetUserAddonTemplatesResponse.ValidateAll()
// if the designated constraints aren't met.
type GetUserAddonTemplatesResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetUserAddonTemplatesResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetUserAddonTemplatesResponseMultiError) AllErrors() []error { return m }

// GetUserAddonTemplatesResponseValidationError is the validation error
// returned by GetUserAddonTemplatesResponse.Validate if the designated
// constraints aren't met.
type GetUserAddonTemplatesResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetUserAddonTemplatesResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetUserAddonTemplatesResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetUserAddonTemplatesResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetUserAddonTemplatesResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetUserAddonTemplatesResponseValidationError) ErrorName() string {
	return "GetUserAddonTemplatesResponseValidationError"
}

// Error satisfies the builtin error interface
func (e GetUserAddonTemplatesResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetUserAddonTemplatesResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetUserAddonTemplatesResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetUserAddonTemplatesResponseValidationError{}

// Validate checks the field values on GetWorkspaceTemplatesResponse with the
// rules defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetWorkspaceTemplatesResponse) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetWorkspaceTemplatesResponse with
// the rules defined in the proto definition for this message. If any rules
// are violated, the result is a list of violation errors wrapped in
// GetWorkspaceTemplatesResponseMultiError, or nil if none found.
func (m *GetWorkspaceTemplatesResponse) ValidateAll() error {
	return m.validate(true)
}

func (m *GetWorkspaceTemplatesResponse) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Message

	for idx, item := range m.GetItems() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, GetWorkspaceTemplatesResponseValidationError{
						field:  fmt.Sprintf("Items[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, GetWorkspaceTemplatesResponseValidationError{
						field:  fmt.Sprintf("Items[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return GetWorkspaceTemplatesResponseValidationError{
					field:  fmt.Sprintf("Items[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return GetWorkspaceTemplatesResponseMultiError(errors)
	}

	return nil
}

// GetWorkspaceTemplatesResponseMultiError is an error wrapping multiple
// validation errors returned by GetWorkspaceTemplatesResponse.ValidateAll()
// if the designated constraints aren't met.
type GetWorkspaceTemplatesResponseMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetWorkspaceTemplatesResponseMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetWorkspaceTemplatesResponseMultiError) AllErrors() []error { return m }

// GetWorkspaceTemplatesResponseValidationError is the validation error
// returned by GetWorkspaceTemplatesResponse.Validate if the designated
// constraints aren't met.
type GetWorkspaceTemplatesResponseValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetWorkspaceTemplatesResponseValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetWorkspaceTemplatesResponseValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetWorkspaceTemplatesResponseValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetWorkspaceTemplatesResponseValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetWorkspaceTemplatesResponseValidationError) ErrorName() string {
	return "GetWorkspaceTemplatesResponseValidationError"
}

// Error satisfies the builtin error interface
func (e GetWorkspaceTemplatesResponseValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetWorkspaceTemplatesResponse.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetWorkspaceTemplatesResponseValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetWorkspaceTemplatesResponseValidationError{}