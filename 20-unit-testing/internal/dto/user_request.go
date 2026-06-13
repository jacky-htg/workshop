package dto

import (
	"workshop/internal/model"
)

type UserRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=100"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=10"`
	Email    string `json:"email" validate:"required,email"`
	IsActive bool   `json:"is_active"`

	Roles []int `json:"roles"`
}

func (u *UserRequest) Transform(user *model.User) {
	user.Name = u.Name
	user.Username = u.Username
	user.Password = u.Password
	user.Email = u.Email
	user.IsActive = u.IsActive

	user.Roles = make([]model.Role, 0)

	for _, v := range u.Roles {
		user.Roles = append(user.Roles, model.Role{ID: v})
	}
}

type UserUpdateRequest struct {
	Name     string `json:"name" validate:"required,min=3,max=100"`
	IsActive bool   `json:"is_active"`

	Roles []int `json:"roles"`
}

func (u *UserUpdateRequest) Transform(user *model.User) {
	user.Name = u.Name
	user.IsActive = u.IsActive

	user.Roles = make([]model.Role, 0)

	for _, v := range u.Roles {
		user.Roles = append(user.Roles, model.Role{ID: v})
	}
}
