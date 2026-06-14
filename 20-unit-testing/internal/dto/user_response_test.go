package dto_test

import (
	"testing"
	"workshop/internal/dto"
	"workshop/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestUserResponse_Transform(t *testing.T) {
	user := model.User{
		ID:       "uuid-1",
		Name:     "admin",
		Username: "admin",
		Password: "secret",
		Email:    "admin@example.com",
		IsActive: true,
		Roles: []model.Role{
			{ID: 1, Name: "admin"},
		},
	}
	req := dto.UserResponse{}
	req.Transform(user)

	assert.Equal(t, req.Name, user.Name)
	assert.Equal(t, req.Roles[0].ID, user.Roles[0].ID)
}
