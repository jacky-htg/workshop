package dto_test

import (
	"testing"
	"workshop/internal/dto"
	"workshop/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestUserRequest_Transform(t *testing.T) {
	user := &model.User{}
	req := dto.UserRequest{
		Name:     "admin",
		Username: "admin",
		Password: "secret",
		Email:    "admin@example.com",
		IsActive: true,
		Roles:    []int{1},
	}
	req.Transform(user)

	assert.Equal(t, req.Name, user.Name)
	assert.Equal(t, req.Roles[0], user.Roles[0].ID)
}

func TestUserUpdateRequest_Transform(t *testing.T) {
	user := &model.User{}
	req := dto.UserUpdateRequest{
		Name:     "admin",
		IsActive: true,
		Roles:    []int{1},
	}
	req.Transform(user)

	assert.Equal(t, req.Name, user.Name)
	assert.Equal(t, req.Roles[0], user.Roles[0].ID)
}
