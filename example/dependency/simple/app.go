package main

import (
	"fmt"
	"reflect"

	"github.com/lvlcn-t/go-kit/dependency"
)

type MyInterface interface {
	DoSomething()
}

type MyFirstImplementation struct{}

func (m *MyFirstImplementation) DoSomething() {
	fmt.Println("MyFirstImplementation.DoSomething")
}

type MySecondImplementation struct{}

func (m *MySecondImplementation) DoSomething() {
	fmt.Println("MySecondImplementation.DoSomething")
}

func main() {
	// Create a new dependency container.
	c := dependency.NewContainer()

	// Provide some dependencies.
	c.Provide(
		dependency.NewSingleton[MyInterface](&MyFirstImplementation{}).Named("my-first-implementation"),
		dependency.NewFactory(func() MyInterface { return &MySecondImplementation{} }).Named("my-second-implementation"),
	)

	// Resolve the dependencies by type or name.
	first := c.Resolve(reflect.TypeFor[MyInterface]()).(MyInterface)
	second := c.ResolveNamed("my-second-implementation").(MyInterface)

	// Use the dependencies.
	first.DoSomething()
	second.DoSomething()
}
