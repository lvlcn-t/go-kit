package config

import "errors"

// Rule represents a validation Rule.
type Rule string

// Validation rules.
const (
	ruleRequired         Rule = "required"
	ruleMinimum          Rule = "min"
	ruleMaximum          Rule = "max"
	ruleLength           Rule = "len"
	ruleEqual            Rule = "eq"
	ruleNotEqual         Rule = "ne"
	ruleGreaterThan      Rule = "gt"
	ruleLessThan         Rule = "lt"
	ruleGreaterThanEqual Rule = "gte"
	ruleLessThanEqual    Rule = "lte"
)

// String returns the string representation of the rule.
func (r Rule) String() string {
	return string(r)
}

// Register registers the custom rule to be considered by [Validate].
// Returns an error if the rule's name is the same as a built-in rule.
func (r Rule) Register(v RuleChecker) error {
	if r.isBuiltin() {
		return errors.New("builtin rule cannot be registered")
	}
	validators[r] = v
	return nil
}

// Unregister unregisters a custom rule from the validation rules.
// Returns an error if the rule is a built-in rule.
func (r Rule) Unregister() error {
	if r.isBuiltin() {
		return errors.New("builtin rule cannot be unregistered")
	}
	delete(validators, r)
	return nil
}

// isValid checks if the rule is a valid validation rule.
func (r Rule) isValid() bool {
	_, ok := validators[r]
	return ok
}

// isBuiltin checks if the rule is a built-in validation rule.
func (r Rule) isBuiltin() bool {
	_, ok := builtinRules[r]
	return ok
}

// RuleChecker is an interface that defines a method to validate a value against a condition.
type RuleChecker interface {
	// Validate checks if the value satisfies the condition.
	// The condition is a string representation of the rule's value. (i.e., "min=10" -> "10")
	// Returns an error if the value does not satisfy the condition.
	Validate(value any, condition string) error
}

// builtinRules contains the built-in validation rules.
var builtinRules = map[Rule]RuleChecker{
	ruleRequired:         newRequiredRule(),
	ruleMinimum:          newComparisonRule(ruleMinimum),
	ruleMaximum:          newComparisonRule(ruleMaximum),
	ruleLength:           newLengthRule(),
	ruleEqual:            newEqualRule(),
	ruleNotEqual:         newUneqalRule(),
	ruleGreaterThan:      newComparisonRule(ruleGreaterThan),
	ruleLessThan:         newComparisonRule(ruleLessThan),
	ruleGreaterThanEqual: newComparisonRule(ruleGreaterThanEqual),
	ruleLessThanEqual:    newComparisonRule(ruleLessThanEqual),
}
