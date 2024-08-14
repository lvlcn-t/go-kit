package dependency

import (
	"reflect"
	"testing"
)

const something = "something"

type TestInterface interface {
	DoSomething() string
}

type TestImplementation struct{}

func (t *TestImplementation) DoSomething() string {
	return something
}

func TestInjector_Provide(t *testing.T) {
	tests := []struct {
		name      string
		iface     reflect.Type
		dep       any
		singleton bool
		factory   func() any
		depName   string
		wantErr   bool
	}{
		{
			name:      "Provide valid dependency",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       &TestImplementation{},
			singleton: false,
			factory:   nil,
			depName:   "",
			wantErr:   false,
		},
		{
			name:      "Provide with nil dependency and factory",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       nil,
			singleton: false,
			factory:   nil,
			depName:   "",
			wantErr:   true,
		},
		{
			name:      "Provide with factory function",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       nil,
			singleton: true,
			factory: func() any {
				return &TestImplementation{}
			},
			depName: "",
			wantErr: false,
		},
		{
			name:      "Provide with named dependency",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       &TestImplementation{},
			singleton: false,
			factory:   nil,
			depName:   "TestImplementation1",
			wantErr:   false,
		},
		{
			name:      "Provide with factory function that returns nil",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       nil,
			singleton: true,
			factory: func() any {
				return nil
			},
			depName: "",
			wantErr: true,
		},
		{
			name:      "Provide with nil interface",
			iface:     nil,
			dep:       &TestImplementation{},
			singleton: false,
			factory:   nil,
			depName:   "",
			wantErr:   false,
		},
		{
			name:      "Provide without dependency nor interface",
			iface:     nil,
			dep:       nil,
			singleton: false,
			factory:   nil,
			depName:   "",
			wantErr:   true,
		},
		{
			name:      "Provide with invalid interface",
			iface:     reflect.TypeOf("invalid"),
			dep:       &TestImplementation{},
			singleton: false,
			factory:   nil,
			depName:   "",
			wantErr:   true,
		},
		{
			name:      "Provided dependency does not implement interface",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       "invalid",
			singleton: false,
			factory:   nil,
			depName:   "",
			wantErr:   true,
		},
		{
			name:      "Provide with invalid name",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       &TestImplementation{},
			singleton: false,
			factory:   nil,
			depName:   "-1",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewDIContainer()
			err := c.Provide(tt.dep, tt.iface, tt.singleton, tt.factory, tt.depName)
			if (err != nil) != tt.wantErr {
				t.Errorf("injector.Provide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if _, ok := c.(*injector).dependencies[tt.iface]; !ok && tt.iface != nil {
					t.Errorf("injector.Provide() did not store dependency")
				}
				if tt.depName != "" {
					if _, ok := c.(*injector).registry[tt.depName]; !ok {
						t.Errorf("injector.Provide() did not store dependency name")
					}
				}
			}
		})
	}
}

func TestInjector_Resolve(t *testing.T) {
	tests := []struct {
		name      string
		iface     reflect.Type
		dep       any
		singleton bool
		factory   func() any
		depName   string
		wantErr   bool
	}{
		{
			name:      "Resolve existing dependency by type name",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       &TestImplementation{},
			singleton: false,
			factory:   nil,
			depName:   "",
			wantErr:   false,
		},
		{
			name:      "Resolve existing named dependency",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       &TestImplementation{},
			singleton: false,
			factory:   nil,
			depName:   "TestImplementation1",
			wantErr:   false,
		},
		{
			name:      "Resolve non-existing dependency",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       nil,
			singleton: false,
			factory:   nil,
			depName:   "",
			wantErr:   true,
		},
		{
			name:      "Resolve singleton dependency with factory",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       nil,
			singleton: true,
			factory: func() any {
				return &TestImplementation{}
			},
			depName: "",
			wantErr: false,
		},
		{
			name:      "Resolve singleton dependency without factory",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       &TestImplementation{},
			singleton: true,
			factory:   nil,
			depName:   "",
			wantErr:   false,
		},
		{
			name:      "Resolve first dependency with empty name",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       &TestImplementation{},
			singleton: false,
			factory:   nil,
			depName:   "",
			wantErr:   false,
		},
		{
			name:      "Resolve last dependency with '-1' name",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       &TestImplementation{},
			singleton: false,
			factory:   nil,
			depName:   "-1",
			wantErr:   false,
		},
		{
			name:      "Resolve with invalid name",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       &TestImplementation{},
			singleton: false,
			factory:   nil,
			depName:   "nonexistent",
			wantErr:   true,
		},
		{
			name:      "Resolve non-singleton dependency with factory function",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       nil,
			singleton: false,
			factory: func() any {
				return &TestImplementation{}
			},
			depName: "",
			wantErr: false,
		},
		{
			name:      "Resolve with invalid type",
			iface:     nil,
			dep:       nil,
			singleton: false,
			factory:   nil,
			depName:   "",
			wantErr:   true,
		},
		{
			name:      "Resolve with type mismatch between given type and dependency name",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       &TestImplementation{},
			singleton: false,
			factory:   nil,
			depName:   "mismatch",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewDIContainer()
			dep := tt.dep
			iface := tt.iface
			name := tt.depName
			switch {
			case tt.dep == nil && tt.factory == nil:
				c.(*injector).dependencies[tt.iface] = []dependency{{
					value:     reflect.ValueOf(tt.dep),
					factory:   tt.factory,
					singleton: false,
				}}
			case tt.depName == "mismatch":
				c.(*injector).dependencies[tt.iface] = []dependency{{
					value:     reflect.ValueOf(tt.dep),
					factory:   tt.factory,
					singleton: tt.singleton,
				}}
				c.(*injector).registry[tt.depName] = dependencyInfo{
					iface: reflect.TypeOf((*string)(nil)).Elem(),
					dep:   reflect.TypeOf((*TestImplementation)(nil)).Elem(),
				}
			case tt.depName == "-1" || tt.depName == "nonexistent":
				name = ""
				fallthrough
			default:
				err := c.Provide(dep, iface, tt.singleton, tt.factory, name)
				if err != nil {
					t.Fatalf("injector.Provide() setup failed: %v", err)
				}
			}

			val, err := c.Resolve(tt.iface, tt.depName)
			if (err != nil) != tt.wantErr {
				t.Errorf("injector.Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if val == nil {
					t.Errorf("injector.Resolve() returned nil, want non-nil")
				}
				iface, ok := val.(TestInterface)
				if !ok {
					t.Errorf("injector.Resolve() returned wrong type")
				}

				if iface.DoSomething() != something {
					t.Errorf("injector.Resolve().DoSomething() = %s, want %s", iface.DoSomething(), something)
				}
			}
		})
	}
}

func TestInjector_ResolveAll(t *testing.T) {
	tests := []struct {
		name    string
		iface   reflect.Type
		deps    []any
		wantErr bool
	}{
		{
			name:  "ResolveAll with multiple dependencies",
			iface: reflect.TypeOf((*TestInterface)(nil)).Elem(),
			deps: []any{
				&TestImplementation{},
				&TestImplementation{},
			},
			wantErr: false,
		},
		{
			name:    "ResolveAll with no dependencies",
			iface:   reflect.TypeOf((*TestInterface)(nil)).Elem(),
			deps:    nil,
			wantErr: true,
		},
		{
			name:    "ResolveAll with invalid type",
			iface:   nil,
			deps:    nil,
			wantErr: true,
		},
		{
			name:  "ResolveAll with invalid dependency",
			iface: reflect.TypeOf((*TestInterface)(nil)).Elem(),
			deps: []any{
				nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewDIContainer()
			if len(tt.deps) != 0 && tt.deps[0] != nil {
				for _, dep := range tt.deps {
					err := c.Provide(dep, tt.iface, false, nil)
					if err != nil {
						t.Fatalf("injector.Provide() setup failed: %v", err)
					}
				}
			} else {
				c.(*injector).dependencies[tt.iface] = []dependency{}
			}

			results, err := c.ResolveAll(tt.iface)
			if (err != nil) != tt.wantErr {
				t.Errorf("injector.ResolveAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if len(results) != len(tt.deps) {
					t.Errorf("injector.ResolveAll() returned %d dependencies, want %d", len(results), len(tt.deps))
				}

				for i, dep := range results {
					if dep == nil && tt.deps[i] != nil {
						t.Errorf("injector.ResolveAll() returned nil at index %d, want non-nil", i)
					}
					iface, ok := dep.(TestInterface)
					if !ok {
						t.Errorf("injector.ResolveAll() returned wrong type")
					}

					if iface.DoSomething() != something {
						t.Errorf("injector.ResolveAll().DoSomething() = %s, want %s", iface.DoSomething(), something)
					}
				}
			}
		})
	}
}

func TestInjector_Delete(t *testing.T) {
	tests := []struct {
		name      string
		iface     reflect.Type
		dep       any
		wantExist bool
	}{
		{
			name:      "Delete existing dependency",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       &TestImplementation{},
			wantExist: false,
		},
		{
			name:      "Delete non-existing dependency",
			iface:     reflect.TypeOf((*TestInterface)(nil)).Elem(),
			dep:       nil,
			wantExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewDIContainer()
			if tt.dep != nil {
				err := c.Provide(tt.dep, tt.iface, false, nil)
				if err != nil {
					t.Fatalf("injector.Provide() setup failed: %v", err)
				}
			}

			c.Delete(tt.iface)

			_, err := c.Resolve(tt.iface)
			if (err == nil) != tt.wantExist {
				t.Errorf("injector.Delete() unexpected dependency presence, got error: %v", err)
			}
		})
	}
}
