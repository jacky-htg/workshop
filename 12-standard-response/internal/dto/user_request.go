package dto

import "workshop/internal/model"

type UserRequest struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

func (u *UserRequest) Transform(user *model.User) {
	user.Name = u.Name
	user.Username = u.Username
	user.Password = u.Password
	user.Email = u.Email
	user.IsActive = u.IsActive
}

type UserUpdateRequest struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

func (u *UserUpdateRequest) Transform(user *model.User) {
	user.Name = u.Name
	user.IsActive = u.IsActive
}
