package dto

import "workshop/internal/model"

type RoleRequest struct {
	Name string `json:"name" validate:"required,min=3,max=25"`
}

func (u *RoleRequest) Transform(role *model.Role) {
	role.Name = u.Name
}
