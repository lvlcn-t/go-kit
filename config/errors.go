package config

import (
	"errors"
	"fmt"
	"strings"
)

// ErrFieldRequired is an error type that indicates a required field is missing in the configuration.
type ErrFieldRequired struct {
	Field   string
	Type    string
	Message string
}

// Error returns the error message.
func (e ErrFieldRequired) Error() string {
	if e.Field == "" && e.Type == "" && e.Message == "" {
		return ""
	}

	if e.Message == "" {
		return fmt.Sprintf("field %q (%s) is required", e.Field, e.Type)
	}

	return fmt.Sprintf("field %q (%s) is required: %s", e.Field, e.Type, e.Message)
}

// Is returns true if the target error is an [ErrFieldRequired].
func (e ErrFieldRequired) Is(target error) bool {
	_, ok := target.(ErrFieldRequired)
	if !ok {
		_, ok = target.(*ErrFieldRequired)
	}
	return ok
}

// ErrFieldInvalid is an error type that indicates an invalid field in the configuration.
type ErrFieldInvalid struct {
	Field   string
	Type    string
	Message string
}

// Error returns the error message.
func (e ErrFieldInvalid) Error() string {
	if e.Field == "" && e.Type == "" && e.Message == "" {
		return ""
	}

	if e.Message == "" {
		return fmt.Sprintf("field %q (%s) is invalid", e.Field, e.Type)
	}

	return fmt.Sprintf("field %q (%s) is invalid: %s", e.Field, e.Type, e.Message)
}

// Is returns true if the target error is an [ErrFieldInvalid].
func (e ErrFieldInvalid) Is(target error) bool {
	_, ok := target.(ErrFieldInvalid)
	if !ok {
		_, ok = target.(*ErrFieldInvalid)
	}
	return ok
}

// newFieldError creates a new field error.
func newFieldError(name, typ string, err error) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case ErrFieldRequired, *ErrFieldRequired:
		return &ErrFieldRequired{Field: name, Type: typ, Message: err.Error()}
	case ErrFieldInvalid, *ErrFieldInvalid:
		return &ErrFieldInvalid{Field: name, Type: typ, Message: err.Error()}
	case *ruleError:
		for _, e := range err.(interface{ Unwrap() []error }).Unwrap() {
			e = newFieldError(name, typ, e)
			if e != nil {
				return e
			}
		}
		return nil
	default:
		return err
	}
}

// ErrConfigEmpty is an error type that indicates an empty configuration.
type ErrConfigEmpty struct{}

// Error returns the error message.
func (e ErrConfigEmpty) Error() string {
	return "you must provide a configuration"
}

// parserError is an error type that indicates a parsing error with the validation rule.
type parserError struct {
	rule  rule
	value string
}

// Error returns the error message.
func (e *parserError) Error() string {
	if e.rule == "" {
		return fmt.Sprintf("invalid value %q", e.value)
	}
	return fmt.Sprintf("invalid value %q for %s", e.value, e.rule)
}

// Is returns true if the target error is a [parserError].
func (e *parserError) Is(target error) bool {
	_, ok := target.(*parserError)
	return ok
}

// newParserError creates a new parser error.
func newParserError(rule rule, value string) error {
	return &parserError{rule: rule, value: value}
}

// ruleError is an error type that indicates a validation error with the configuration.
type ruleError struct {
	typ  rule
	errs []error
}

// Error returns the error message.
func (e *ruleError) Error() string {
	var b strings.Builder
	for i, err := range e.errs {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

// Unwrap returns the list of wrapped errors.
func (e *ruleError) Unwrap() []error {
	return e.errs
}

// Is returns true if the target error is a [ruleError].
func (e *ruleError) Is(target error) bool {
	_, ok := target.(*ruleError)
	return ok
}

// newRuleError creates a new rule error which wraps a list of errors.
func newRuleError(typ rule, errs ...error) error {
	err := errors.Join(errs...)
	if err == nil {
		return nil
	}

	return &ruleError{typ: typ, errs: err.(interface{ Unwrap() []error }).Unwrap()}
}

// comparisonError is an error type that indicates a comparison error with the configuration.
type comparisonError struct {
	value any
	cond  any
	rule  rule
}

// Error returns the error message.
func (e *comparisonError) Error() string {
	value := fmt.Sprintf("%v", e.value)
	condition := fmt.Sprintf("%v", e.cond)
	if _, ok := e.cond.(string); ok {
		value = fmt.Sprintf("%q", e.value)
		condition = fmt.Sprintf("%q", e.cond)
	}

	op := map[rule]string{
		ruleGreaterThan:      "greater than",
		ruleLessThan:         "less than",
		ruleGreaterThanEqual: "greater than or equal to",
		ruleLessThanEqual:    "less than or equal to",
		ruleEqual:            "equal to",
		ruleNotEqual:         "not equal to",
		ruleMinimum:          "greater than or equal to",
		ruleMaximum:          "less than or equal to",
		"":                   "",
	}[e.rule]
	if op != "" {
		return fmt.Sprintf("value = %s, want a value %s %s", value, op, condition)
	}

	return fmt.Sprintf("unknown operation: %q", e.rule)
}

// Is returns true if the target error is a [comparisonError].
func (e *comparisonError) Is(target error) bool {
	_, ok := target.(*comparisonError)
	return ok
}

// newComparisonError creates a new comparison error.
func newComparisonError(value, cond any, rule rule) *comparisonError {
	return &comparisonError{value: value, cond: cond, rule: rule}
}
