package dto

import "workshop/internal/model"

type RoleResponse struct {
	ID       int              `json:"id"`
	Name     string           `json:"name,omitempty"`
	Accesses []AccessResponse `json:"accesses,omitempty"`
}

func (u *RoleResponse) Transform(role model.Role) {
	u.ID = role.ID
	u.Name = role.Name
	u.Accesses = make([]AccessResponse, 0)

	for _, a := range role.Accesses {
		var access AccessResponse
		access.Transform(a)
		u.Accesses = append(u.Accesses, access)
	}
}
