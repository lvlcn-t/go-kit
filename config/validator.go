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
	// Validate validates the configuration.
	Validate() error
}

// Validate validates the provided configuration struct using the "validate" tag.
// If the configuration struct implements the [Validator] interface, its [Validate] method is called instead.
//
// The "validate" tag can contain the following rules:
//   - required: the field must not be nil or the zero value
//   - min=<value>: the field must be greater than or equal to the value
//   - max=<value>: the field must be less than or equal to the value
//   - len=<value>: the field must have the specified length
//   - eq=<value>: the field must be equal to the value
//   - ne=<value>: the field must not be equal to the value
//   - gt=<value>: the field must be greater than the value
//   - lt=<value>: the field must be less than the value
//   - gte=<value>: the field must be greater than or equal to the value
//   - lte=<value>: the field must be less than or equal to the value
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
		panic("value must be a struct or a pointer to a struct")
	}

	var errs []error
	v := reflect.ValueOf(cfg)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		typ := v.Type().Field(i)
		tag, ok := typ.Tag.Lookup("validate")
		if !ok || tag == "-" || !field.CanInterface() {
			continue
		}

		if field.Kind() == reflect.Struct {
			if err := Validate(field.Interface()); err != nil {
				errs = append(errs, fmt.Errorf("field %s: %w", typ.Name, err))
			}
			continue
		}

		root := newAST(tag)
		if err := root.validate(field.Interface()); err != nil {
			errs = append(errs, fmt.Errorf("field %s: %w", typ.Name, err))
		}
	}

	return errors.Join(errs...)
}

// newAST parses a validation tag and returns the corresponding AST
func newAST(tag string) *validationNode {
	root := newValidationNode("root", "")
	conditions := lists.Distinct(strings.Split(tag, ","))

	for _, c := range conditions {
		parts := strings.SplitN(c, "=", 2)
		if len(parts) == 1 {
			root.addChild(newValidationNode(parts[0], ""))
			continue
		}
		root.addChild(newValidationNode(parts[0], parts[1]))
	}

	return root
}

// validationNode represents a single validation rule node
type validationNode struct {
	Type     string
	Value    string
	Children []*validationNode
}

// newValidationNode creates a new validation node
func newValidationNode(t, val string) *validationNode {
	return &validationNode{Type: t, Value: val, Children: []*validationNode{}}
}

// addChild adds a child node to the current node
func (n *validationNode) addChild(child *validationNode) {
	n.Children = append(n.Children, child)
}

// validate traverses the AST and applies validation rules
func (n *validationNode) validate(val any) error {
	if validator, ok := validators[n.Type]; ok {
		return validator.Validate(val, n.Value)
	}

	var err error
	for _, child := range n.Children {
		if vErr := child.validate(val); vErr != nil {
			if errors.Is(vErr, newParserError("", "")) {
				panic(vErr)
			}
			err = errors.Join(err, vErr)
		}
	}
	return err
}

var validators = map[string]rule{
	"required": newRequiredValidator(),
	"min":      newComparisonValidator("min"),
	"max":      newComparisonValidator("max"),
	"len":      newLenValidator(),
	"eq":       newEqValidator(),
	"ne":       newNeValidator(),
	"gt":       newComparisonValidator("gt"),
	"lt":       newComparisonValidator("lt"),
	"gte":      newComparisonValidator("gte"),
	"lte":      newComparisonValidator("lte"),
}

type rule interface {
	Validate(input any, condition string) error
}

type requiredValidator struct{}

func newRequiredValidator() rule {
	return &requiredValidator{}
}

func (v *requiredValidator) Validate(input any, _ string) error {
	if input == nil || reflect.DeepEqual(input, reflect.Zero(reflect.TypeOf(input)).Interface()) {
		return errors.New("field is required")
	}
	return nil
}

type comparisonValidator struct {
	op string
}

func newComparisonValidator(op string) rule {
	return &comparisonValidator{op: op}
}

func (v *comparisonValidator) Validate(input any, condition string) error {
	if input == nil {
		return nil
	}
	switch input.(type) {
	case int, int8, int16, int32, int64:
		val, err := strconv.ParseInt(condition, 10, 64)
		if err != nil {
			return newParserError(v.op, condition)
		}
		return compare(reflect.ValueOf(input).Int(), val, v.op)
	case uint, uint8, uint16, uint32, uint64:
		val, err := strconv.ParseUint(condition, 10, 64)
		if err != nil {
			return newParserError(v.op, condition)
		}
		return compare(reflect.ValueOf(input).Uint(), val, v.op)
	case float32, float64:
		val, err := strconv.ParseFloat(condition, 64)
		if err != nil {
			return newParserError(v.op, condition)
		}
		return compare(reflect.ValueOf(input).Float(), val, v.op)
	}
	return nil
}

func compare[T cmp.Ordered](input, condition T, op string) error {
	switch op {
	case "gt":
		if input <= condition {
			return fmt.Errorf("value must be greater than %v", condition)
		}
	case "lt":
		if input >= condition {
			return fmt.Errorf("value must be less than %v", condition)
		}
	case "gte", "min":
		if input < condition {
			return fmt.Errorf("value must be at least %v", condition)
		}
	case "lte", "max":
		if input > condition {
			return fmt.Errorf("value must be at most %v", condition)
		}
	}
	return nil
}

type lenValidator struct{}

func newLenValidator() rule {
	return &lenValidator{}
}

func (v *lenValidator) Validate(input any, length string) error {
	l, err := strconv.Atoi(length)
	if err != nil {
		return newParserError("len", length)
	}

	switch v := input.(type) {
	case string:
		if len(v) != l {
			return fmt.Errorf("length must be %d", l)
		}
	case []any:
		if len(v) != l {
			return fmt.Errorf("length must be %d", l)
		}
	}
	return nil
}

// eqValidator checks for equality.
type eqValidator struct{}

func newEqValidator() rule {
	return &eqValidator{}
}

func (v *eqValidator) Validate(input any, eq string) error {
	return compareValues(input, eq, "eq")
}

// neValidator checks for inequality.
type neValidator struct{}

func newNeValidator() rule {
	return &neValidator{}
}

func (v *neValidator) Validate(input any, ne string) error {
	return compareValues(input, ne, "ne")
}

// compareValues compares input with a condition based on the operation.
func compareValues[T comparable](input T, condition, op string) error {
	val, err := parseCondition[T](condition)
	if err != nil {
		return err
	}

	switch op {
	case "eq":
		if input != val {
			return fmt.Errorf("value must be eq to %v", val)
		}
	case "ne":
		if input == val {
			return fmt.Errorf("value must be ne to %v", val)
		}
	default:
		return fmt.Errorf("unknown operation: %s", op)
	}

	return nil
}

// parseCondition parses the condition string to the type of input.
func parseCondition[T comparable](condition string) (T, error) {
	var zero T
	for reflect.TypeOf(zero).Kind() == reflect.Pointer {
		zero = reflect.ValueOf(zero).Elem().Interface().(T)
	}

	switch any(zero).(type) {
	case int, int8, int16, int32, int64:
		val, err := strconv.ParseInt(condition, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("parse error: %v", err)
		}
		return any(val).(T), nil
	case uint, uint8, uint16, uint32, uint64:
		val, err := strconv.ParseUint(condition, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("parse error: %v", err)
		}
		return any(val).(T), nil
	case float32, float64:
		val, err := strconv.ParseFloat(condition, 64)
		if err != nil {
			return zero, fmt.Errorf("parse error: %v", err)
		}
		return any(val).(T), nil
	case string:
		return any(condition).(T), nil
	default:
		return zero, fmt.Errorf("unsupported type: %s", reflect.TypeOf(zero).Name())
	}
}

type parserError struct {
	field string
	value string
}

func (e *parserError) Error() string {
	return fmt.Sprintf("invalid value %q for %s", e.value, e.field)
}

func newParserError(field, value string) error {
	return &parserError{field: field, value: value}
}
