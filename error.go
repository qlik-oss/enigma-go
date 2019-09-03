package enigma

import (
	"fmt"
  "strings"
)

type (
	// Error extends the built in error type with extra error information provided by the Qlik Associative Engine
	Error interface {
		error
		Code() int
		Parameter() string
		Message() string
	}

	qixError struct {
		ErrorCode      int    `json:"code"`
		ErrorParameter string `json:"parameter"`
		ErrorMessage   string `json:"message"`
	}
)

func (err *qixError) Error() string {
  errorType := ErrorCodeLookup(err.ErrorCode)
  if errorType[:7] == "LOCERR_" {
    errorType = errorType[7:]
  }
  errorType = strings.ReplaceAll(errorType, "_", " ")
	return fmt.Sprintf("%s: %s (%d %s)", err.ErrorParameter, err.ErrorMessage, err.ErrorCode, errorType)
}

func (err *qixError) Code() int {
	return err.ErrorCode
}

func (err *qixError) Parameter() string {
	return err.ErrorParameter
}

func (err *qixError) Message() string {
	return err.ErrorMessage
}
