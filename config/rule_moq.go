// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package config

import (
	"sync"
)

// Ensure, that RuleCheckerMock does implement RuleChecker.
// If this is not the case, regenerate this file with moq.
var _ RuleChecker = &RuleCheckerMock{}

// RuleCheckerMock is a mock implementation of RuleChecker.
//
//	func TestSomethingThatUsesRuleChecker(t *testing.T) {
//
//		// make and configure a mocked RuleChecker
//		mockedRuleChecker := &RuleCheckerMock{
//			ValidateFunc: func(value any, condition string) error {
//				panic("mock out the Validate method")
//			},
//		}
//
//		// use mockedRuleChecker in code that requires RuleChecker
//		// and then make assertions.
//
//	}
type RuleCheckerMock struct {
	// ValidateFunc mocks the Validate method.
	ValidateFunc func(value any, condition string) error

	// calls tracks calls to the methods.
	calls struct {
		// Validate holds details about calls to the Validate method.
		Validate []struct {
			// Value is the value argument value.
			Value any
			// Condition is the condition argument value.
			Condition string
		}
	}
	lockValidate sync.RWMutex
}

// Validate calls ValidateFunc.
func (mock *RuleCheckerMock) Validate(value any, condition string) error {
	if mock.ValidateFunc == nil {
		panic("RuleCheckerMock.ValidateFunc: method is nil but RuleChecker.Validate was just called")
	}
	callInfo := struct {
		Value     any
		Condition string
	}{
		Value:     value,
		Condition: condition,
	}
	mock.lockValidate.Lock()
	mock.calls.Validate = append(mock.calls.Validate, callInfo)
	mock.lockValidate.Unlock()
	return mock.ValidateFunc(value, condition)
}

// ValidateCalls gets all the calls that were made to Validate.
// Check the length with:
//
//	len(mockedRuleChecker.ValidateCalls())
func (mock *RuleCheckerMock) ValidateCalls() []struct {
	Value     any
	Condition string
} {
	var calls []struct {
		Value     any
		Condition string
	}
	mock.lockValidate.RLock()
	calls = mock.calls.Validate
	mock.lockValidate.RUnlock()
	return calls
}
