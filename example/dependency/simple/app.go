package main

import (
	"fmt"

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
	// Provide a dependency.
	err := dependency.Provide[MyInterface](&MyFirstImplementation{}, false, nil)
	if err != nil {
		panic(err)
	}

	// Provide another dependency.
	err = dependency.Provide[MyInterface](&MySecondImplementation{}, false, nil)
	if err != nil {
		panic(err)
	}

	// Resolve the dependencies.
	first, err := dependency.Resolve[MyInterface]()
	if err != nil {
		panic(err)
	}

	second, err := dependency.Resolve[MyInterface](1)
	if err != nil {
		panic(err)
	}

	// Use the dependencies.
	first.DoSomething()
	second.DoSomething()
}
