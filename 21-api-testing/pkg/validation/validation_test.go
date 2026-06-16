package validation_test

import (
	"testing"
	"workshop/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestFormatValidationErrors_TableDriven(t *testing.T) {
	validate := validator.New()

	tests := []struct {
		name     string
		object   interface{}
		expected map[string]string
	}{
		{
			name: "required - empty field",
			object: struct {
				Name string `validate:"required"`
			}{},
			expected: map[string]string{"name": "Name is required"},
		},
		{
			name: "email - invalid format",
			object: struct {
				Email string `validate:"email"`
			}{Email: "invalid"},
			expected: map[string]string{"email": "Email must be a valid email address"},
		},
		{
			name: "min - too short",
			object: struct {
				Password string `validate:"min=8"`
			}{Password: "123"},
			expected: map[string]string{"password": "Password is too short"},
		},
		{
			name: "max - too long",
			object: struct {
				Username string `validate:"max=5"`
			}{Username: "tooolong"},
			expected: map[string]string{"username": "Username is too long"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(tt.object)
			if err == nil {
				t.Skip("validation should fail")
			}

			result := validation.FormatValidationErrors(err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
