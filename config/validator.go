package config

import (
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
		if !ok || tag == "-" {
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
	root := newValidationNode("root", nil)
	conditions := lists.Distinct(strings.Split(tag, ","))

	for _, c := range conditions {
		switch {
		case c == "required":
			root.addChild(newValidationNode("required", nil))
		case strings.HasPrefix(c, "min="):
			root.addChild(newValidationNode("min", strings.TrimPrefix(c, "min=")))
		case strings.HasPrefix(c, "max="):
			root.addChild(newValidationNode("max", strings.TrimPrefix(c, "max=")))
		case strings.HasPrefix(c, "len="):
			root.addChild(newValidationNode("len", strings.TrimPrefix(c, "len=")))
		case strings.HasPrefix(c, "eq="):
			root.addChild(newValidationNode("eq", strings.TrimPrefix(c, "eq=")))
		case strings.HasPrefix(c, "ne="):
			root.addChild(newValidationNode("ne", strings.TrimPrefix(c, "ne=")))
		case strings.HasPrefix(c, "gt="):
			root.addChild(newValidationNode("gt", strings.TrimPrefix(c, "gt=")))
		case strings.HasPrefix(c, "lt="):
			root.addChild(newValidationNode("lt", strings.TrimPrefix(c, "lt=")))
		case strings.HasPrefix(c, "gte="):
			root.addChild(newValidationNode("gte", strings.TrimPrefix(c, "gte=")))
		case strings.HasPrefix(c, "lte="):
			root.addChild(newValidationNode("lte", strings.TrimPrefix(c, "lte=")))
		}
	}

	return root
}

// validationNode represents a single validation rule node
type validationNode struct {
	Type     string
	Value    any
	Children []*validationNode
}

// newValidationNode creates a new validation node
func newValidationNode(t string, val any) *validationNode {
	return &validationNode{Type: t, Value: val, Children: []*validationNode{}}
}

// addChild adds a child node to the current node
func (n *validationNode) addChild(child *validationNode) {
	n.Children = append(n.Children, child)
}

// validate traverses the AST and applies validation rules
func (n *validationNode) validate(val any) error { //nolint:gocyclo // TODO: refactor to validators map (map[string]Validator)
	switch n.Type {
	case "required":
		if val == nil || reflect.DeepEqual(val, reflect.Zero(reflect.TypeOf(val)).Interface()) {
			return errors.New("field is required")
		}
	case "min":
		return validateMin(val, n.Value.(string))
	case "max":
		return validateMax(val, n.Value.(string))
	case "len":
		return validateLen(val, n.Value.(string))
	case "eq":
		return validateEq(val, n.Value.(string))
	case "ne":
		return validateNe(val, n.Value.(string))
	case "gt":
		return validateGt(val, n.Value.(string))
	case "lt":
		return validateLt(val, n.Value.(string))
	case "gte":
		return validateGte(val, n.Value.(string))
	case "lte":
		return validateLte(val, n.Value.(string))
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

// validateLen validates the length of the value
func validateLen(val any, length string) error {
	l, err := strconv.Atoi(length)
	if err != nil {
		return newParserError("len", length)
	}

	switch v := val.(type) {
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

// validateMin validates the minimum value
func validateMin(val any, minimum string) error { //nolint:dupl // no need to refactor for now
	switch val.(type) {
	case int, int8, int16, int32, int64:
		m, err := strconv.ParseInt(minimum, 10, 64)
		if err != nil {
			return newParserError("min", minimum)
		}

		if reflect.ValueOf(val).Int() < m {
			return fmt.Errorf("value must be at least %d", m)
		}
	case uint, uint8, uint16, uint32, uint64:
		m, err := strconv.ParseUint(minimum, 10, 64)
		if err != nil {
			return newParserError("min", minimum)
		}

		if reflect.ValueOf(val).Uint() < m {
			return fmt.Errorf("value must be at least %d", m)
		}
	case float32, float64:
		m, err := strconv.ParseFloat(minimum, 64)
		if err != nil {
			return newParserError("min", minimum)
		}

		if reflect.ValueOf(val).Float() < m {
			return fmt.Errorf("value must be at least %f", m)
		}
	}
	return nil
}

// validateMax validates the maximum value
func validateMax(val any, maximum string) error { //nolint:dupl // no need to refactor for now
	switch val.(type) {
	case int, int8, int16, int32, int64:
		m, err := strconv.ParseInt(maximum, 10, 64)
		if err != nil {
			return newParserError("max", maximum)
		}

		if reflect.ValueOf(val).Int() > m {
			return fmt.Errorf("value must be at most %d", m)
		}
	case uint, uint8, uint16, uint32, uint64:
		m, err := strconv.ParseUint(maximum, 10, 64)
		if err != nil {
			return newParserError("max", maximum)
		}

		if reflect.ValueOf(val).Uint() > m {
			return fmt.Errorf("value must be at most %d", m)
		}
	case float32, float64:
		m, err := strconv.ParseFloat(maximum, 64)
		if err != nil {
			return newParserError("max", maximum)
		}

		if reflect.ValueOf(val).Float() > m {
			return fmt.Errorf("value must be at most %f", m)
		}
	}
	return nil
}

// validateEq validates the value to be equal to the specified value
func validateEq(val any, eq string) error { //nolint:dupl // no need to refactor for now
	switch v := val.(type) {
	case int, int8, int16, int32, int64:
		e, err := strconv.ParseInt(eq, 10, 64)
		if err != nil {
			return newParserError("eq", eq)
		}

		if reflect.ValueOf(val).Int() != e {
			return fmt.Errorf("value must be equal to %d", e)
		}
	case uint, uint8, uint16, uint32, uint64:
		e, err := strconv.ParseUint(eq, 10, 64)
		if err != nil {
			return newParserError("eq", eq)
		}

		if reflect.ValueOf(val).Uint() != e {
			return fmt.Errorf("value must be equal to %d", e)
		}
	case float32, float64:
		e, err := strconv.ParseFloat(eq, 64)
		if err != nil {
			return newParserError("eq", eq)
		}

		if reflect.ValueOf(val).Float() != e {
			return fmt.Errorf("value must be equal to %f", e)
		}
	case string:
		if v != eq {
			return fmt.Errorf("value must be equal to %s", eq)
		}
	}
	return nil
}

// validateNe validates the value to be not equal to the specified value
func validateNe(val any, ne string) error { //nolint:dupl // no need to refactor for now
	switch v := val.(type) {
	case int, int8, int16, int32, int64:
		e, err := strconv.ParseInt(ne, 10, 64)
		if err != nil {
			return newParserError("ne", ne)
		}

		if reflect.ValueOf(val).Int() == e {
			return fmt.Errorf("value must not be equal to %d", e)
		}
	case uint, uint8, uint16, uint32, uint64:
		e, err := strconv.ParseUint(ne, 10, 64)
		if err != nil {
			return newParserError("ne", ne)
		}

		if reflect.ValueOf(val).Uint() == e {
			return fmt.Errorf("value must not be equal to %d", e)
		}
	case float32, float64:
		e, err := strconv.ParseFloat(ne, 64)
		if err != nil {
			return newParserError("ne", ne)
		}

		if reflect.ValueOf(val).Float() == e {
			return fmt.Errorf("value must not be equal to %f", e)
		}
	case string:
		if v == ne {
			return fmt.Errorf("value must not be equal to %s", ne)
		}
	}
	return nil
}

// validateGt validates the value to be greater than the specified value
func validateGt(val any, gt string) error { //nolint:dupl // no need to refactor for now
	switch val.(type) {
	case int, int8, int16, int32, int64:
		g, err := strconv.ParseInt(gt, 10, 64)
		if err != nil {
			return newParserError("gt", gt)
		}

		if reflect.ValueOf(val).Int() <= g {
			return fmt.Errorf("value must be greater than %d", g)
		}
	case uint, uint8, uint16, uint32, uint64:
		g, err := strconv.ParseUint(gt, 10, 64)
		if err != nil {
			return newParserError("gt", gt)
		}

		if reflect.ValueOf(val).Uint() <= g {
			return fmt.Errorf("value must be greater than %d", g)
		}
	case float32, float64:
		g, err := strconv.ParseFloat(gt, 64)
		if err != nil {
			return newParserError("gt", gt)
		}

		if reflect.ValueOf(val).Float() <= g {
			return fmt.Errorf("value must be greater than %f", g)
		}
	}
	return nil
}

// validateLt validates the value to be less than the specified value
func validateLt(val any, lt string) error { //nolint:dupl // no need to refactor for now
	switch val.(type) {
	case int, int8, int16, int32, int64:
		l, err := strconv.ParseInt(lt, 10, 64)
		if err != nil {
			return newParserError("lt", lt)
		}

		if reflect.ValueOf(val).Int() >= l {
			return fmt.Errorf("value must be less than %d", l)
		}
	case uint, uint8, uint16, uint32, uint64:
		l, err := strconv.ParseUint(lt, 10, 64)
		if err != nil {
			return newParserError("lt", lt)
		}

		if reflect.ValueOf(val).Uint() >= l {
			return fmt.Errorf("value must be less than %d", l)
		}
	case float32, float64:
		l, err := strconv.ParseFloat(lt, 64)
		if err != nil {
			return newParserError("lt", lt)
		}

		if reflect.ValueOf(val).Float() >= l {
			return fmt.Errorf("value must be less than %f", l)
		}
	}
	return nil
}

// validateGte validates the value to be greater than or equal to the specified value
func validateGte(val any, gte string) error { //nolint:dupl // no need to refactor for now
	switch val.(type) {
	case int, int8, int16, int32, int64:
		g, err := strconv.ParseInt(gte, 10, 64)
		if err != nil {
			return newParserError("gte", gte)
		}

		if reflect.ValueOf(val).Int() < g {
			return fmt.Errorf("value must be greater than or equal to %d", g)
		}
	case uint, uint8, uint16, uint32, uint64:
		g, err := strconv.ParseUint(gte, 10, 64)
		if err != nil {
			return newParserError("gte", gte)
		}

		if reflect.ValueOf(val).Uint() < g {
			return fmt.Errorf("value must be greater than or equal to %d", g)
		}
	case float32, float64:
		g, err := strconv.ParseFloat(gte, 64)
		if err != nil {
			return newParserError("gte", gte)
		}

		if reflect.ValueOf(val).Float() < g {
			return fmt.Errorf("value must be greater than or equal to %f", g)
		}
	}
	return nil
}

// validateLte validates the value to be less than or equal to the specified value
func validateLte(val any, lte string) error { //nolint:dupl // no need to refactor for now
	switch val.(type) {
	case int, int8, int16, int32, int64:
		l, err := strconv.ParseInt(lte, 10, 64)
		if err != nil {
			return newParserError("lte", lte)
		}

		if reflect.ValueOf(val).Int() > l {
			return fmt.Errorf("value must be less than or equal to %d", l)
		}
	case uint, uint8, uint16, uint32, uint64:
		l, err := strconv.ParseUint(lte, 10, 64)
		if err != nil {
			return newParserError("lte", lte)
		}

		if reflect.ValueOf(val).Uint() > l {
			return fmt.Errorf("value must be less than or equal to %d", l)
		}
	case float32, float64:
		l, err := strconv.ParseFloat(lte, 64)
		if err != nil {
			return newParserError("lte", lte)
		}

		if reflect.ValueOf(val).Float() > l {
			return fmt.Errorf("value must be less than or equal to %f", l)
		}
	}
	return nil
}
