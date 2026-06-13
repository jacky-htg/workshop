package errors

import (
	"errors"
	"fmt"
	"net/http"
)

const (
	InternalServerErrorCode = "E000"
	InvalidInputCode        = "E001"
	NotFoundCode            = "E002"
	ForbiddenCode           = "E003"
	UnauthorizedCode        = "E004"
	GatewayTimeoutCode      = "E005"
)

// Default messages
const (
	InternalServerErrorMessage = "Internal Server Error"
	InvalidInputMessage        = "Invalid input"
	NotFoundMessage            = "Resource not found"
	ForbiddenMessage           = "Forbidden"
	UnauthorizedMessage        = "Unauthorized"
	GatewayTimeoutMessage      = "Gateway Timeout"
)

type BusinessError struct {
	Err        error
	Code       string
	Message    string
	HTTPStatus int
}

// Error implements error interface
func (err *BusinessError) Error() string {
	if err.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", err.Code, err.Message, err.Err)
	}
	return fmt.Sprintf("[%s] %s", err.Code, err.Message)
}

// Unwrap untuk error wrapping (Go 1.13+)
func (err *BusinessError) Unwrap() error {
	return err.Err
}

// Helper functions untuk create error
func ErrNew(code string, message string, httpStatus int) *BusinessError {
	return &BusinessError{
		Err:        fmt.Errorf("%s", message),
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

func ErrWrap(err error, bErr *BusinessError) {
	bErr.Err = err
}

// Quick constructors tanpa wrap
func InvalidInput(message ...string) *BusinessError {
	finalMessage := InvalidInputMessage
	if len(message) > 0 && len(message[0]) > 0 {
		finalMessage = message[0]
	}
	return ErrNew(InvalidInputCode, finalMessage, http.StatusBadRequest)
}

func NotFound(message ...string) *BusinessError {
	finalMessage := NotFoundMessage
	if len(message) > 0 && len(message[0]) > 0 {
		finalMessage = message[0]
	}
	return ErrNew(NotFoundCode, finalMessage, http.StatusNotFound)
}

func Forbidden(message ...string) *BusinessError {
	finalMessage := ForbiddenMessage
	if len(message) > 0 && len(message[0]) > 0 {
		finalMessage = message[0]
	}
	return ErrNew(ForbiddenCode, finalMessage, http.StatusForbidden)
}

func Unauthorized(message ...string) *BusinessError {
	finalMessage := UnauthorizedMessage
	if len(message) > 0 && len(message[0]) > 0 {
		finalMessage = message[0]
	}
	return ErrNew(UnauthorizedCode, finalMessage, http.StatusUnauthorized)
}

func InternalServerError(message ...string) *BusinessError {
	finalMessage := InternalServerErrorMessage
	if len(message) > 0 && len(message[0]) > 0 {
		finalMessage = message[0]
	}
	return ErrNew(InternalServerErrorCode, finalMessage, http.StatusInternalServerError)
}

func GatewayTimeout(message ...string) *BusinessError {
	finalMessage := GatewayTimeoutMessage
	if len(message) > 0 && len(message[0]) > 0 {
		finalMessage = message[0]
	}
	return ErrNew(GatewayTimeoutCode, finalMessage, http.StatusGatewayTimeout)
}

// Wrapped constructors
func InvalidInputWrap(err error, message ...string) *BusinessError {
	bErr := InvalidInput(message...)
	ErrWrap(err, bErr)
	return bErr
}

func NotFoundWrap(err error, message ...string) *BusinessError {
	bErr := NotFound(message...)
	ErrWrap(err, bErr)
	return bErr
}

func ForbiddenWrap(err error, message ...string) *BusinessError {
	bErr := Forbidden(message...)
	ErrWrap(err, bErr)
	return bErr
}

func UnauthorizedWrap(err error, message ...string) *BusinessError {
	bErr := Unauthorized(message...)
	ErrWrap(err, bErr)
	return bErr
}

func InternalServerErrorWrap(err error, message ...string) *BusinessError {
	bErr := InternalServerError(message...)
	ErrWrap(err, bErr)
	return bErr
}

func GatewayTimeoutWrap(err error, message ...string) *BusinessError {
	bErr := GatewayTimeout(message...)
	ErrWrap(err, bErr)
	return bErr
}

// GetBusinessError extracts BusinessError from error chain
func GetBusinessError(err error) (*BusinessError, bool) {
	var bizErr *BusinessError
	if errors.As(err, &bizErr) {
		return bizErr, true
	}
	return nil, false
}
