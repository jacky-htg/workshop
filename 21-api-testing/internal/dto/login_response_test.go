package dto_test

import (
	"testing"
	"workshop/internal/dto"
	"workshop/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestLoginResponse_Transform(t *testing.T) {
	// Setup
	token := "test-token-123"
	user := model.User{
		ID:       "uuid-123",
		Name:     "John Doe",
		Username: "johndoe",
		Email:    "john@example.com",
		IsActive: true,
	}
	accesses := []string{"read", "write", "delete"}

	// Execute
	var resp dto.LoginResponse
	resp.Transform(token, user, accesses)

	// Assert
	assert.Equal(t, token, resp.Token)
	assert.Equal(t, accesses, resp.Accesses)
	assert.Equal(t, user.ID, resp.User.ID)
	assert.Equal(t, user.Name, resp.User.Name)
	assert.Equal(t, user.Username, resp.User.Username)
	assert.Equal(t, user.Email, resp.User.Email)
	assert.Equal(t, user.IsActive, resp.User.IsActive)
}
