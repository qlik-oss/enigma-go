package enigma

import (
	"fmt"
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
	return fmt.Sprintf("%s: %s (%d)", err.ErrorParameter, err.ErrorMessage, err.ErrorCode)
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
