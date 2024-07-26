package config

import (
	"errors"
	"testing"
)

type testConfig_CustomValidator struct {
	Host string `validate:"required"`
}

func (c testConfig_CustomValidator) Validate() error {
	if c.Host != "valid" {
		return errors.New("custom validation failed")
	}
	return nil
}

func TestValidate(t *testing.T) {
	type testConfig struct {
		Host string `validate:"required"`
		Port uint32 `validate:"required,min=1024,max=65535"`
	}

	type testConfig_NestedStruct struct {
		Inner testConfig `validate:"required"`
	}

	type testConfig_WithUnexportedField struct {
		Host string `validate:"required"`
		port int    `validate:"required,min=1024,max=65535"`
	}

	type testConfig_WithPointer struct {
		Host *string `validate:"required"`
		Port *int    `validate:"required,min=1024,max=65535"`
	}

	type testConfig_WithSkipField struct {
		Host string `validate:"required"`
		Port int    `validate:"-"`
	}

	type testConfig_WithInvalidTag struct {
		Host string `validate:"required,invalid"`
		Port int    `validate:"required,min=1024,max=65535"`
	}

	type testConfig_WithInvalidTagValue struct {
		Host string `validate:"required"`
		Port int    `validate:"required,min=1024,max=invalid"`
	}

	type testConfig_AllValidations struct {
		Host      string  `validate:"required,len=7,eq=example"`
		Port      int     `validate:"required,min=1024,max=65535"`
		Range     int     `validate:"required,gte=10,lte=20"`
		Exclusive int     `validate:"required,gt=10,lt=20"`
		Float     float64 `validate:"required,gte=10.5,lte=20.5"`
		Skip      int     `validate:"-"`
	}

	type testConfig_SliceMap struct {
		Hosts []string          `validate:"required,len=3"`
		Ports map[string]uint32 `validate:"required,min=1"`
	}

	type notStruct int

	type testConfig_EQNE struct {
		StrField   string  `validate:"eq=test,ne=not_test"`
		IntField   int     `validate:"eq=5,ne=6"`
		UIntField  uint    `validate:"eq=5,ne=6"`
		FloatField float64 `validate:"eq=1.23,ne=4.56"`
	}

	type testConfig_EQNE_Ptr struct {
		StrField   *string  `validate:"eq=test,ne=not_test"`
		IntField   *int     `validate:"eq=5,ne=6"`
		UIntField  *uint    `validate:"eq=5,ne=6"`
		FloatField *float64 `validate:"eq=1.23,ne=4.56"`
	}

	tests := []struct {
		name      string
		config    any
		wantErr   bool
		wantPanic bool
	}{
		{
			name: "valid config",
			config: testConfig{
				Host: "localhost",
				Port: 8080,
			},
			wantErr: false,
		},
		{
			name: "valid config as pointer",
			config: &testConfig{
				Host: "localhost",
				Port: 8080,
			},
			wantErr: false,
		},
		{
			name:    "missing required field",
			config:  testConfig{},
			wantErr: true,
		},
		{
			name: "invalid port, less than minimum",
			config: testConfig{
				Host: "localhost",
				Port: 1000,
			},
			wantErr: true,
		},
		{
			name: "invalid port, more than maximum",
			config: testConfig{
				Host: "localhost",
				Port: 70000,
			},
			wantErr: true,
		},
		{
			name: "valid nested config",
			config: testConfig_NestedStruct{
				Inner: testConfig{
					Host: "localhost",
					Port: 8080,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid nested config",
			config: testConfig_NestedStruct{
				Inner: testConfig{
					Host: "localhost",
					Port: 1000,
				},
			},
			wantErr: true,
		},
		{
			name: "valid pointer config",
			config: testConfig_WithPointer{
				Host: toPtr("localhost"),
				Port: toPtr(8080),
			},
			wantErr: false,
		},
		{
			name: "invalid pointer config",
			config: testConfig_WithPointer{
				Host: toPtr("localhost"),
				Port: toPtr(1000),
			},
			wantErr: true,
		},
		{
			name:    "nil pointer config",
			config:  testConfig_WithPointer{},
			wantErr: true,
		},
		{
			name: "valid unexported field config",
			config: testConfig_WithUnexportedField{
				Host: "localhost",
				port: 8080,
			},
			wantErr: false,
		},
		{
			name: "invalid unexported field config",
			config: testConfig_WithUnexportedField{
				Host: "localhost",
				port: 1000,
			},
			wantErr: false,
		},
		{
			name: "skip field config",
			config: testConfig_WithSkipField{
				Host: "localhost",
				Port: 0,
			},
			wantErr: false,
		},
		{
			name: "config with an invalid tag",
			config: testConfig_WithInvalidTag{
				Host: "localhost",
				Port: 8080,
			},
			wantErr:   true,
			wantPanic: true,
		},
		{
			name: "config with an invalid tag value",
			config: testConfig_WithInvalidTagValue{
				Host: "localhost",
				Port: 8080,
			},
			wantErr:   true,
			wantPanic: true,
		},
		{
			name: "config with all validations valid",
			config: testConfig_AllValidations{
				Host:      "example",
				Port:      8080,
				Range:     20,
				Exclusive: 15,
				Float:     15.5,
				Skip:      0,
			},
			wantErr: false,
		},
		{
			name: "config with all validations invalid",
			config: testConfig_AllValidations{
				Host:      "example",
				Port:      8080,
				Range:     25,
				Exclusive: 25,
				Float:     25.5,
				Skip:      0,
			},
			wantErr: true,
		},
		{
			name: "valid slice and map config",
			config: testConfig_SliceMap{
				Hosts: []string{"host1", "host2", "host3"},
				Ports: map[string]uint32{"port1": 1024},
			},
			wantErr: false,
		},
		{
			name: "invalid slice config",
			config: testConfig_SliceMap{
				Hosts: []string{"host1", "host2"},
				Ports: map[string]uint32{"port1": 1024},
			},
			wantErr: true,
		},
		{
			name: "invalid map config",
			config: testConfig_SliceMap{
				Hosts: []string{"host1", "host2", "host3"},
				Ports: map[string]uint32{},
			},
			wantErr: true,
		},
		{
			name: "custom validator valid",
			config: testConfig_CustomValidator{
				Host: "valid",
			},
			wantErr: false,
		},
		{
			name: "custom validator invalid",
			config: testConfig_CustomValidator{
				Host: "invalid",
			},
			wantErr: true,
		},
		{
			name: "unsupported type",
			config: struct {
				Channel chan int `validate:"required"`
			}{
				Channel: make(chan int),
			},
			wantErr: false,
		},
		{
			name:      "not a struct",
			config:    notStruct(0),
			wantErr:   true,
			wantPanic: true,
		},
		{
			name: "boundary value min valid",
			config: testConfig{
				Host: "localhost",
				Port: 1024,
			},
			wantErr: false,
		},
		{
			name: "boundary value max valid",
			config: testConfig{
				Host: "localhost",
				Port: 65535,
			},
			wantErr: false,
		},
		{
			name: "eq and ne validators valid",
			config: testConfig_EQNE{
				StrField:   "test",
				IntField:   5,
				UIntField:  5,
				FloatField: 1.23,
			},
			wantErr: false,
		},
		{
			name: "eq and ne validators invalid",
			config: testConfig_EQNE{
				StrField:   "not_test",
				IntField:   6,
				UIntField:  6,
				FloatField: 4.56,
			},
			wantErr: true,
		},
		{
			name: "eq and ne validators valid with pointers",
			config: testConfig_EQNE_Ptr{
				StrField:   toPtr("test"),
				IntField:   toPtr(5),
				UIntField:  toPtr[uint](5),
				FloatField: toPtr(1.23),
			},
			wantErr: false,
		},
		{
			name: "eq and ne validators invalid with pointers",
			config: testConfig_EQNE_Ptr{
				StrField:   toPtr("not_test"),
				IntField:   toPtr(6),
				UIntField:  toPtr[uint](6),
				FloatField: toPtr(4.56),
			},
			wantErr: true,
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

			err := Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				t.Logf("errors:\n%v", err)
			}
		})
	}
}

func toPtr[T any](v T) *T {
	return &v
}
