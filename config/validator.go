package config

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/lvlcn-t/go-kit/lists"
)

// Validator is an interface that can be implemented by a [Loadable] to be validated with custom rules.
type Validator interface {
	// Validate checks the configuration and returns an error if it is invalid.
	Validate() error
}

// Validate checks the provided configuration struct using the "validate" tag.
// If the configuration's type implements the [Validator] interface, its [Validate] method is called instead.
//
// The "validate" tag can contain the following rules:
//   - required: the field must not be nil or the zero value (available for any type)
//   - min=<value>: the field must be greater than or equal to the value (available for [cmp.Ordered], slice, and map types)
//   - max=<value>: the field must be less than or equal to the value (available for [cmp.Ordered], slice, and map types)
//   - len=<value>: the field must have the specified length (available for string, slice, map, and chan types)
//   - eq=<value>: the field must be equal to the value (available for [cmp.Ordered] types)
//   - ne=<value>: the field must not be equal to the value (available for [cmp.Ordered] types)
//   - gt=<value>: the field must be greater than the value (available for [cmp.Ordered] types)
//   - lt=<value>: the field must be less than the value (available for [cmp.Ordered] types)
//   - gte=<value>: the field must be greater than or equal to the value (available for [cmp.Ordered] types)
//   - lte=<value>: the field must be less than or equal to the value (available for [cmp.Ordered] types)
//
// Example:
//
//	type Config struct {
//		Host string `validate:"required"`
//		Port int `validate:"required,min=1024,max=65535"`
//	}
//
//	cfg := Config{Host: "localhost", Port: 8080}
//	if err := config.Validate(cfg); err != nil {
//		// Handle error
//	}
func Validate(cfg any) error {
	if c, ok := cfg.(Validator); ok {
		return c.Validate()
	}

	for reflect.TypeOf(cfg).Kind() == reflect.Pointer {
		cfg = reflect.ValueOf(cfg).Elem().Interface()
	}

	if reflect.TypeOf(cfg).Kind() != reflect.Struct {
		panic("value must be a struct or a pointer to a struct to use validation without implementing the Validator interface")
	}

	var errs []error
	val := reflect.ValueOf(cfg)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typ := val.Type().Field(i)
		tag, ok := typ.Tag.Lookup("validate")
		if !ok || tag == "-" || !field.CanInterface() {
			continue
		}

		if field.Kind() == reflect.Struct {
			if err := Validate(field.Interface()); err != nil {
				errs = append(errs, newFieldError(typ.Name, typ.Type.String(), err))
			}
			continue
		}

		root := newAST(tag)
		if err := root.apply(field.Interface()); err != nil {
			errs = append(errs, newFieldError(typ.Name, typ.Type.String(), err))
		}
	}

	return errors.Join(errs...)
}

// newAST parses a validation tag and returns the corresponding AST node.
func newAST(tag string) *node {
	root := newNode("root", "")
	conditions := lists.Distinct(strings.Split(tag, ","))

	for _, cond := range conditions {
		parts := strings.SplitN(cond, "=", 2)
		if len(parts) == 1 {
			root.addChild(newNode(parts[0], ""))
			continue
		}
		root.addChild(newNode(parts[0], parts[1]))
	}

	return root
}

// node represents a single validation rule in the abstract syntax tree.
// Each node can have multiple children, which represent additional rules.
type node struct {
	typ      rule
	value    string
	children []*node
}

// newNode creates a new validation rule node.
func newNode(typ, val string) *node {
	return &node{typ: rule(typ), value: val, children: []*node{}}
}

// addChild adds a child node to the current node.
func (n *node) addChild(child *node) {
	n.children = append(n.children, child)
}

// apply traverses the AST and applies validation rules.
func (n *node) apply(value any) error {
	if validator, ok := validators[n.typ]; ok {
		return validator.Validate(value, n.value)
	}

	var err error
	for _, child := range n.children {
		if vErr := child.apply(value); vErr != nil {
			if errors.Is(vErr, newParserError("", "")) {
				panic(vErr)
			}
			err = newRuleError(child.typ, err, vErr)
		}
	}
	return err
}

// rule represents a validation rule.
type rule string

// String returns the string representation of the rule.
func (a rule) String() string {
	return string(a)
}

// Validation rules.
const (
	ruleRequired         rule = "required"
	ruleMinimum          rule = "min"
	ruleMaximum          rule = "max"
	ruleLength           rule = "len"
	ruleEqual            rule = "eq"
	ruleNotEqual         rule = "ne"
	ruleGreaterThan      rule = "gt"
	ruleLessThan         rule = "lt"
	ruleGreaterThanEqual rule = "gte"
	ruleLessThanEqual    rule = "lte"
)

// validators contains all available validation rules.
var validators = map[rule]interface {
	Validate(value any, condition string) error
}{
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

// requiredRule is a validation rule that checks if a field is not nil or the zero value.
type requiredRule struct{}

// newRequiredRule creates a new required rule.
func newRequiredRule() *requiredRule {
	return &requiredRule{}
}

// Validate checks if the value is not nil or the zero value.
func (v *requiredRule) Validate(value any, _ string) error {
	if value == nil || reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface()) {
		return &ErrFieldRequired{}
	}
	return nil
}

// lengthRule is a validation rule that checks if a field has the specified length.
type lengthRule struct{}

// newLengthRule creates a new length rule.
func newLengthRule() *lengthRule {
	return &lengthRule{}
}

// Validate checks if the value has the specified length.
func (v *lengthRule) Validate(value any, length string) error {
	l, err := strconv.Atoi(length)
	if err != nil {
		return newParserError(ruleLength, length)
	}

	val, null := getValue(value)
	if null {
		return nil
	}

	var size int
	switch val.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		size = val.Len()
	default:
		return fmt.Errorf("unsupported type: %s", val.Type())
	}

	if size != l {
		return fmt.Errorf("length must be %d; got %d", l, size)
	}

	return nil
}

// comparisonRule is a validation rule that checks if a field satisfies a comparison rule.
type comparisonRule struct {
	rule rule
}

// newComparisonRule creates a new comparison rule.
func newComparisonRule(rule rule) *comparisonRule {
	return &comparisonRule{rule: rule}
}

// Validate checks if the value satisfies the comparison rule.
func (v *comparisonRule) Validate(value any, condition string) error {
	val, null := getValue(value)
	if null {
		return nil
	}

	switch val.Kind() {
	case reflect.String:
		return compare(val.String(), condition, v.rule)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		condVal, err := strconv.ParseInt(condition, 10, 64)
		if err != nil {
			return newParserError(v.rule, condition)
		}
		return compare(val.Int(), condVal, v.rule)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		condVal, err := strconv.ParseUint(condition, 10, 64)
		if err != nil {
			return newParserError(v.rule, condition)
		}
		return compare(val.Uint(), condVal, v.rule)
	case reflect.Float32, reflect.Float64:
		condVal, err := strconv.ParseFloat(condition, 64)
		if err != nil {
			return newParserError(v.rule, condition)
		}
		return compare(val.Float(), condVal, v.rule)
	case reflect.Slice, reflect.Array, reflect.Map:
		if v.rule != ruleMinimum && v.rule != ruleMaximum {
			return newParserError(v.rule, condition)
		}
		condVal, err := strconv.Atoi(condition)
		if err != nil {
			return newParserError(v.rule, condition)
		}
		return compare(val.Len(), condVal, v.rule)
	}

	return nil
}

// compare checks if the value satisfies the comparison rule.
func compare[T cmp.Ordered](value, condition T, rule rule) error {
	switch rule {
	case ruleGreaterThan:
		if cmp.Compare(value, condition) <= 0 {
			return newComparisonError(value, condition, ruleGreaterThan)
		}
	case ruleLessThan:
		if cmp.Compare(value, condition) >= 0 {
			return newComparisonError(value, condition, ruleLessThan)
		}
	case ruleGreaterThanEqual, ruleMinimum:
		if cmp.Compare(value, condition) < 0 {
			return newComparisonError(value, condition, ruleGreaterThanEqual)
		}
	case ruleLessThanEqual, ruleMaximum:
		if cmp.Compare(value, condition) > 0 {
			return newComparisonError(value, condition, ruleLessThanEqual)
		}
	}
	return nil
}

// equalRule is a validation rule that checks if a field is equal to a value.
type equalRule struct{}

// newEqualRule creates a new equal rule.
func newEqualRule() *equalRule {
	return &equalRule{}
}

// Validate checks if the value is equal to the condition.
func (v *equalRule) Validate(value any, condition string) error {
	return validateEquality(value, condition, ruleEqual)
}

// unequalRule is a validation rule that checks if a field is not equal to a value.
type unequalRule struct{}

// newUneqalRule creates a new unequal rule.
func newUneqalRule() *unequalRule {
	return &unequalRule{}
}

// Validate checks if the value is not equal to the condition.
func (v *unequalRule) Validate(value any, condition string) error {
	return validateEquality(value, condition, ruleNotEqual)
}

// validateEquality checks if the value is either equal or not equal to the condition depending on the operation.
func validateEquality(value any, condition string, op rule) error {
	val, null := getValue(value)
	if null {
		return nil
	}

	var err error
	switch val.Kind() {
	case reflect.String:
		err = compareEquality(val.String(), condition, op)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = compareEquality(val.Int(), condition, op)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = compareEquality(val.Uint(), condition, op)
	case reflect.Float32, reflect.Float64:
		err = compareEquality(val.Float(), condition, op)
	default:
		err = fmt.Errorf("unsupported type: %s", val.Type().Name())
	}

	return err
}

// compareEquality checks if the value is either equal or not equal to the condition.
// The operation is determined by the provided rule.
func compareEquality[T cmp.Ordered](value T, condition string, op rule) error {
	condVal, err := parseValue[T](condition)
	if err != nil {
		if errors.Is(err, newParserError("", "")) {
			err = newParserError(op, condition)
		}
		return err
	}

	switch op {
	case ruleEqual:
		if value != condVal {
			return newComparisonError(value, condVal, ruleEqual)
		}
	case ruleNotEqual:
		if value == condVal {
			return newComparisonError(value, condVal, ruleNotEqual)
		}
	default:
		return newComparisonError(value, condVal, op)
	}

	return nil
}

// parseValue parses a string value into the specified type.
func parseValue[T cmp.Ordered](condition string) (T, error) {
	var zero T
	typ := reflect.TypeOf(zero)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	switch typ.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(condition, 10, 64)
		if err != nil {
			return zero, newParserError("", condition)
		}
		return reflect.ValueOf(val).Convert(typ).Interface().(T), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(condition, 10, 64)
		if err != nil {
			return zero, newParserError("", condition)
		}
		return reflect.ValueOf(val).Convert(typ).Interface().(T), nil
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(condition, 64)
		if err != nil {
			return zero, newParserError("", condition)
		}
		return reflect.ValueOf(val).Convert(typ).Interface().(T), nil
	case reflect.String:
		return reflect.ValueOf(condition).Convert(typ).Interface().(T), nil
	default:
		return zero, fmt.Errorf("unsupported type for validation tag: %s", typ.Name())
	}
}

// getValue returns the reflect value of the provided value.
// If the value is a pointer, it is dereferenced.
// Returns the reflect value and a boolean indicating if the value is nil.
func getValue(value any) (val reflect.Value, isNil bool) {
	if value == nil {
		return reflect.Value{}, true
	}

	val = reflect.ValueOf(value)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return reflect.Value{}, true
		}
		val = val.Elem()
	}

	return val, false
}
