package errors_test

import (
	"fmt"
	"net/http"
	"testing"
	"workshop/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestAllErrors(t *testing.T) {
	tests := []struct {
		name       string
		errFunc    func(...string) *errors.BusinessError
		code       string
		defaultMsg string
		httpStatus int
	}{
		{"InvalidInput", errors.InvalidInput, errors.InvalidInputCode, "Invalid input", http.StatusBadRequest},
		{"NotFound", errors.NotFound, errors.NotFoundCode, "Resource not found", http.StatusNotFound},
		{"Forbidden", errors.Forbidden, errors.ForbiddenCode, "Forbidden", http.StatusForbidden},
		{"Unauthorized", errors.Unauthorized, errors.UnauthorizedCode, "Unauthorized", http.StatusUnauthorized},
		{"InternalServerError", errors.InternalServerError, errors.InternalServerErrorCode, "Internal Server Error", http.StatusInternalServerError},
		{"GatewayTimeout", errors.GatewayTimeout, errors.GatewayTimeoutCode, "Gateway Timeout", http.StatusGatewayTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test default message
			err := tt.errFunc()
			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.defaultMsg, err.Message)
			assert.Equal(t, tt.httpStatus, err.HTTPStatus)

			// Test custom message
			customMsg := "custom message"
			err = tt.errFunc(customMsg)
			assert.Equal(t, customMsg, err.Message)
		})
	}
}

func TestWrapErrors(t *testing.T) {
	originalErr := fmt.Errorf("original error")

	wrappers := []struct {
		name string
		wrap func(error, ...string) *errors.BusinessError
	}{
		{"InvalidInputWrap", errors.InvalidInputWrap},
		{"NotFoundWrap", errors.NotFoundWrap},
		{"ForbiddenWrap", errors.ForbiddenWrap},
		{"UnauthorizedWrap", errors.UnauthorizedWrap},
		{"InternalServerErrorWrap", errors.InternalServerErrorWrap},
		{"GatewayTimeoutWrap", errors.GatewayTimeoutWrap},
	}

	for _, w := range wrappers {
		t.Run(w.name, func(t *testing.T) {
			err := w.wrap(originalErr)
			assert.ErrorIs(t, err.Err, originalErr)
		})
	}
}

func TestGetBusinessError(t *testing.T) {
	bizErr := errors.InvalidInput()
	extracted, ok := errors.GetBusinessError(bizErr)
	assert.True(t, ok)
	assert.Equal(t, bizErr, extracted)

	normalErr := fmt.Errorf("normal")
	_, ok = errors.GetBusinessError(normalErr)
	assert.False(t, ok)
}

func TestBusinessError_Error(t *testing.T) {
	// Without wrapped error
	err := errors.InvalidInput()
	assert.Equal(t, "[E001] Invalid input: Invalid input", err.Error())

	// With wrapped error
	err = errors.InvalidInputWrap(fmt.Errorf("test"))
	assert.Contains(t, err.Error(), "[E001] Invalid input: test")
}
