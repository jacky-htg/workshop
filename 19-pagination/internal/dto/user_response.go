package dto

import "workshop/internal/model"

type UserResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`

	Roles []RoleResponse `json:"roles,omitempty"`
}

func (u *UserResponse) Transform(user model.User) {
	u.ID = user.ID
	u.Name = user.Name
	u.Username = user.Username
	u.Email = user.Email
	u.IsActive = user.IsActive

	u.Roles = make([]RoleResponse, 0)
	for _, r := range user.Roles {
		var role RoleResponse
		role.Transform(r)
		u.Roles = append(u.Roles, role)
	}
}
