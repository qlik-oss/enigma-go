package enigma

import (
	"testing"
)

func TestErrorCodeLookup(t *testing.T) {
	err := qixError{}
	// Error code -128 is an internal error.
	err.ErrorCode = -128
	expected := "INTERNAL ERROR"
	if actual := err.codeLookup(); actual != expected {
		t.Errorf("Expected: '%s', actual: '%s'", expected, actual)
	}
	// Error code -100 is not a valid error code.
	err.ErrorCode = -100
	if err.codeLookup() != "" {
		t.Error("Expected empty string for invalid error code.")
	}
}
