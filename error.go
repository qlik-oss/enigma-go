package enigma

import (
	"fmt"
  "encoding/json"
  "reflect"
  "runtime/debug"
  "strings"
)

type (
	// Error extends the built in error type with extra error information provided by the Qlik Associative Engine.
  // Code returns the provided error code from engine.
  // Parameter gives some context for the error.
  // Message returns the error message from Engine.
  // CallParams returns contains the parameters used when calling engine if any.
  // Trace returns the stack trace of the error, omitting some superfluous lines.
  // Caller returns the first call in the stack, containing function and file.
	Error interface {
		error
		Code() int
		Parameter() string
		Message() string
    CallParams() string
    Trace() string
    Caller() string
	}

  //qixError contains the stacktrace leading to the error and may contain the parameters used when
  //the error occurred as well as the underlying error informat from engine.
	qixError struct {
		code      int    `json:"code"`
		parameter string `json:"parameter"`
		message   string `json:"message"`
    stack []byte
    callParams []interface{}
	}
)

func (e *qixError) Error() string {
	return fmt.Sprintf("%s: %s (%d)", e.parameter, e.message, e.code)
}

func (e *qixError) Code() int {
	return e.code
}

func (e *qixError) Parameter() string {
	return e.parameter
}

func (e *qixError) Message() string {
	return e.message
}

func (e *qixError) UnmarshalJSON(text []byte) error {
  m := map[string]interface{}{}
  if err := json.Unmarshal(text, &m); err != nil {
    return err
  }
  e.unmarshalMap(m)
  e.stack = debug.Stack()
  return nil
}

func (e *qixError) unmarshalMap(m map[string]interface{}) error {
  e.parameter = m["parameter"].(string)
  e.message = m["message"].(string)
  defer func () {
    if r := recover(); r != nil {
      f:= m["code"].(float64)
      e.code = int(f)
    }
  }()
  code := m["code"].(int)
  e.code = code
  return nil
}


// Trace returns the full stack trace leading to the error
func (e *qixError) Trace() string {
  // Exclude the five first lines as they contain cluttering implementation
  // specific information.
  s := strings.Split(string(e.stack), "\n")
  msg := "Stacktrace:\n"
  msg += strings.Join(s[:len(s)-1], "\n")
  return msg
}

// Caller returns a string containing the calling method along with the file it's in.
func (e *qixError) Caller() string {
  method, line := e.trace()
  return fmt.Sprintf("Calling func: %s\n\tat: %s", method, line)
}

// Params returns a string representation of the parameters used when this error occured.
func (e *qixError) CallParams() string {
  str := "Parameters:\n"
  for _, p := range e.callParams {
    str += pp(p)
  }
  return str
}

// pp does the pretty printing work for VerboseError.Params
// if we want to expand non-nil fields we have to go deeper into
// the reflections
func pp(i interface{}) string {
  var s string
  switch v := reflect.ValueOf(i); v.Kind() {
  case reflect.Struct:
    s = structString(v, 0)
  case reflect.Ptr:
    s = "*" + structString(reflect.Indirect(v), 0)
  default:
    s = fmt.Sprintf("%#v", i)
  }
  return s
}

// structString prretty prints a struct with non-nil field names
func structString(v reflect.Value, d int) string {
  d++
  ind := strings.Repeat("  ", d)
  vt := v.Type()
  s := vt.String() + "{\n"
  for i := 0; i < v.NumField(); i++ {
    f := v.Field(i)
    fName := vt.Field(i).Name
    switch f.Kind() {
    case reflect.Struct:
      s += fName + ": " + structString(f, d) + "\n"
    case reflect.Ptr:
      if !f.IsNil() {
        s += ind + fName + ": *" + structString(reflect.Indirect(f), d) + "\n"
      }
    default:
      s += fieldString(ind, vt.Field(i).Name, f)
    }
  }
  s += ind[:(d-1)*2] + "}"
  if d > 1 {
    s += ","
  }
  return s
}

// fieldString pretty prints a struct field that is not a (*)struct
func fieldString(ind, fName string, f reflect.Value) string {
  k := f.Kind()
  // ignore nil values
  switch {
  case reflect.Chan <= k && reflect.Slice >= k:
    if f.IsNil() {
      return ""
    }
  case reflect.Bool == k:
    if !f.Bool() {
      return ""
    }
  case reflect.Int <= k && reflect.Int64 >= k:
    if f.Int() == 0 {
      return ""
    }
  case reflect.Uint <= k && reflect.Uint64 >= k:
    if f.Uint() == 0 {
      return ""
    }
  case reflect.Float32 == k || reflect.Float64 == k:
    if f.Float() == 0.0 {
      return ""
    }
  case reflect.String == k:
    if f.String() == "" {
      return ""
    }
  }
  s := ind + fName + ": "
  switch k {
  case reflect.Slice:
    s += f.Type().String() + "{"
    if f.Len() > 0 {
      s += "..."
    }
    s += "}"
  case reflect.Ptr:
    s += f.Type().String() + "{...}"
  default:
    s += fmt.Sprintf("%v", f.Interface())
  }
  return s + ",\n"
}

// trace uses the stack trace to return
// the method and line in a file that caused an error
func (e *qixError) trace() (method, line string) {
	// stack has trailing new line
	// method is the third last line
	// file and line is the second last
	stack := strings.Split(string(e.stack), "\n")
	method = stack[len(stack)-3]
	line = stack[len(stack)-2]
	line = strings.Trim(line, "\t ")
	line = strings.Split(line, " ")[0]
	return
}
