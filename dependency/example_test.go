package dependency_test

import (
	"bytes"
	"errors"
	"fmt"
	"iter"
	"reflect"
	"time"

	"github.com/lvlcn-t/go-kit/dependency"
)

func Example() {
	// Create a new dependency container.
	c := dependency.NewContainer()

	// Provide several dependencies.
	c.Provide(
		dependency.NewSingleton(errors.New("error 1")),
		dependency.NewSingleton[fmt.Stringer](bytes.NewBufferString("buffer")),
		// Note that [time.Now] does implement [fmt.Stringer] but the container doesn't recognize it because we didn't provide it explicitly.
		dependency.NewSingletonFunc(time.Now),
		dependency.NewFactory(func() fmt.Stringer { return bytes.NewBufferString("I am a factory") }).Named("factory"),
	)

	// Resolve a dependency.
	err := c.Resolve(reflect.TypeFor[error]()).(error)
	fmt.Println(err)

	// Resolve another dependency.
	buf := c.Resolve(reflect.TypeFor[fmt.Stringer]()).(fmt.Stringer)
	fmt.Println(buf)

	// Resolve the factory dependency by name.
	factory := c.ResolveNamed("factory").(fmt.Stringer)
	fmt.Println(factory)

	// Resolve all dependencies of a specific type.
	next, stop := iter.Pull2(c.ResolveAll(reflect.TypeFor[fmt.Stringer]()))
	defer stop()

	var stringers []fmt.Stringer
	for {
		_, v, ok := next()
		if !ok {
			break
		}
		stringers = append(stringers, v.(fmt.Stringer))
	}

	// Print the resolved dependencies.
	for _, s := range stringers {
		fmt.Println(s)
	}

	// Output:
	// error 1
	// buffer
	// I am a factory
	// buffer
	// I am a factory
}
