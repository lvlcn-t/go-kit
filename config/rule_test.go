package config

import (
	"errors"
	"testing"
)

func TestValidate_WithCustomRule(t *testing.T) {
	type testconfig struct {
		Name string `validate:"custom"`
	}

	tests := []struct {
		name      string
		config    any
		rule      Rule
		checker   RuleChecker
		wantErr   bool
		wantPanic bool
	}{
		{
			name:   "validate with custom rule",
			config: testconfig{Name: "test"},
			rule:   "custom",
			checker: &RuleCheckerMock{
				ValidateFunc: func(value any, condition string) error {
					v, ok := value.(string)
					if !ok {
						t.Errorf("RuleChecker.Validate() value = %T, want string", value)
					}
					if v != "test" {
						t.Errorf("RuleChecker.Validate() value = %q, want %q", v, "test")
					}
					return nil
				},
			},
			wantErr: false,
		},
		{
			name:   "validate with custom rule error",
			config: testconfig{Name: "test"},
			rule:   "custom",
			checker: &RuleCheckerMock{
				ValidateFunc: func(value any, condition string) error {
					v, ok := value.(string)
					if !ok {
						t.Errorf("RuleChecker.Validate() value = %T, want string", value)
					}
					if v != "test" {
						t.Errorf("RuleChecker.Validate() value = %q, want %q", v, "test")
					}
					return errors.New("error")
				},
			},
			wantErr: true,
		},
		{
			name:      "validate with custom rule not registered",
			rule:      "custom",
			wantErr:   true,
			wantPanic: true,
		},
		{
			name: "validate with custom rule with condition",
			config: struct {
				Name string `validate:"custom=hello"`
			}{
				Name: "hello",
			},
			rule: "custom",
			checker: &RuleCheckerMock{
				ValidateFunc: func(value any, condition string) error {
					v, ok := value.(string)
					if !ok {
						t.Errorf("RuleChecker.Validate() value = %T, want string", value)
					}
					if v != condition {
						t.Errorf("RuleChecker.Validate() value = %q, want %q", v, condition)
					}
					return nil
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Validate() did not panic")
					}
				}()
			}

			if err := tt.rule.Register(tt.checker); err != nil {
				t.Fatalf("Rule.Register() error = %v", err)
			}

			if err := Validate(tt.config); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := tt.rule.Unregister(); err != nil {
				t.Fatalf("Rule.Unregister() error = %v", err)
			}
		})
	}
}

func TestRule_Register(t *testing.T) {
	tests := []struct {
		name    string
		rule    Rule
		checker RuleChecker
		wantErr bool
	}{
		{
			name:    "register custom rule",
			rule:    "custom",
			checker: &RuleCheckerMock{},
			wantErr: false,
		},
		{
			name:    "register builtin rule",
			rule:    ruleRequired,
			checker: &RuleCheckerMock{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.rule.Register(tt.checker); (err != nil) != tt.wantErr {
				t.Errorf("Rule.Register() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if _, ok := validators[tt.rule]; !ok {
					t.Errorf("Rule.Register() failed to register the rule")
				}
			}
		})
	}
}

func TestRule_Unregister(t *testing.T) {
	tests := []struct {
		name    string
		rule    Rule
		wantErr bool
	}{
		{
			name:    "unregister custom rule",
			rule:    "custom",
			wantErr: false,
		},
		{
			name:    "unregister builtin rule",
			rule:    ruleRequired,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.rule.Unregister(); (err != nil) != tt.wantErr {
				t.Errorf("Rule.Unregister() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if _, ok := validators[tt.rule]; ok {
					t.Errorf("Rule.Unregister() failed to unregister the rule")
				}
			}
		})
	}
}

func TestRule_String(t *testing.T) {
	tests := []struct {
		name string
		rule Rule
		want string
	}{
		{
			name: "rule required",
			rule: ruleRequired,
			want: "required",
		},
		{
			name: "custom rule",
			rule: "custom",
			want: "custom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rule.String(); got != tt.want {
				t.Errorf("Rule.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
