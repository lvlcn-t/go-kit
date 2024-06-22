package apimanager

import (
	"testing"
)

func TestErrAlreadyRunning(t *testing.T) {
	err := &ErrAlreadyRunning{}
	if err.Error() == "" {
		t.Error("No error message")
	}
	if !err.Is(&ErrAlreadyRunning{}) {
		t.Error("Is() should return true")
	}
}

func TestErrNotRunning(t *testing.T) {
	err := &ErrNotRunning{}
	if err.Error() == "" {
		t.Error("No error message")
	}
	if !err.Is(&ErrNotRunning{}) {
		t.Error("Is() should return true")
	}
}
