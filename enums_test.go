package enigma

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAddArgumentsForType tests if valid arguments can be added for types
func TestAddArgumentsForType(t *testing.T) {
	validArgs := []string{"foo", "bar"}
	AddArgumentsForType(testType(""), validArgs)
	strType := fmt.Sprintf("%T", testType(""))
	for _, arg := range validArgs {
		assert.True(t, typeArgSetMap[strType][arg])
	}
	validArgs = append(validArgs, "kiwi", "pear", "peach")
	AddArgumentsForType(testType(""), validArgs)
	for _, arg := range validArgs {
		assert.True(t, typeArgSetMap[strType][arg])
	}
}

// TestMarshalTextError tests if validateArgs during MarshalText calls works as expected
func TestMarshalTextError(t *testing.T) {
	validArgs := []string{"foo", "bar"}
	// Add validArgs to the set of valid arguments for testType
	AddArgumentsForType(testType(""), validArgs)
	// Valid args:
	assert.NoError(t, testMarshal("foo"))
	assert.NoError(t, testMarshal("bar"))
	// Invalid args:
	assert.Error(t, testMarshal("Foo"))
	assert.Error(t, testMarshal("Fizzyfuzzy"))
	err := testMarshal("Banana")
	assert.Error(t, err)
	msg := err.Error()
	// msg should contain the erroneous argument
	assert.Contains(t, msg, "Banana")
	// msg should list the valid arguments for the type
	for _, arg := range validArgs {
		assert.Contains(t, msg, arg)
	}
}

func testMarshal(t testType) error {
	_, err := json.Marshal(t)
	return err
}

type testType string

func (f testType) String() string {
	return string(f)
}

func (f testType) MarshalText() ([]byte, error) {
	err := validateArg(f)
	return []byte(f), err
}
