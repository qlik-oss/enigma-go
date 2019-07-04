// Enums in this context are not really enumeration but rather valid values for certain arguments
// It should be viewed as argument restriction
// It will provide feedback during runtime, specifically when trying to Marshal a struct.
//
// The goal with this implementation is to provide helpful feedback to the users of enigma.go

package enigma

import (
  "fmt"
  "runtime/debug"
  "strings"
)

type argSet map[string]bool

func (as argSet) String() string {
  args := make([]string, len(as))
  i := 0
  for k := range as {
    args[i] = "\"" + k + "\""
    i++
  }
  s := "[" + strings.Join(args, ", ") + "]"
  return s
}

// validateArg is called by the MarshalText() method of an autogenerate type
// with the type itself as argument.
//
// Even though the type is defined as a string we cannot cast it from interface{} to a string easily
// so we might as well just use the fmt.Stringer interface.
func validateArg(t fmt.Stringer) error {
  sType := fmt.Sprintf("%T", t)
  val := t.String()
  set := typeArgSetMap[sType]
  if !set[val] {
    method, line := trace()
    // TODO shorten this line
    return fmt.Errorf("\n    In function %s at %s\n    \"%s\" is not a valid %s, must be one of: %s", method, line, val, sType, set.String())
  }
  return nil
}

// typeArgSetMap should not be accessed directly.
// This map maps a type to a set of valid arguments for that type.
// As types are not expressions in go we instead have to rely on
// the string representation of a type, meaning:
//
//   fmt.Sprintf("%T", t)
//
// which prints what go thinks the name of the type is.
var typeArgSetMap = map[string]argSet{}

// AddArgumentsForType should be used for adding additional valid arguments for a given type.
// TODO document this function for Andrée's use case.
func AddArgumentsForType(t interface{}, args []string) {
  sType := fmt.Sprintf("%T", t)
  if typeArgSetMap[sType] == nil {
    typeArgSetMap[sType] = map[string]bool{}
  }
  set := typeArgSetMap[sType]
  for _, arg := range args {
    set[arg] = true
  }
}

// trace uses the stack trace to return
// the method and line in a file that caused an error
func trace() (method, line string) {
  // stack has trailing new line
  // method is the third last line
  // file and line is the second last
  stack := strings.Split(string(debug.Stack()), "\n")
  method = stack[len(stack)-3]
  line = stack[len(stack)-2]
  line = strings.Trim(line, "\t ")
  line = strings.Split(line, " ")[0]
  return
}
