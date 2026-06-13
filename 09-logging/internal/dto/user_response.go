package dto

import "workshop/internal/model"

type UserResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

func (u *UserResponse) Transform(user model.User) {
	u.ID = user.ID
	u.Name = user.Name
	u.Username = user.Username
	u.Email = user.Email
	u.IsActive = user.IsActive
}
