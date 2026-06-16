package dto_test

import (
	"testing"
	"workshop/internal/dto"
	"workshop/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestRoleResponse_Transform(t *testing.T) {
	parentID := 1
	role := model.Role{
		ID:   1,
		Name: "admin",
		Accesses: []model.Access{
			model.Access{
				ID:       11,
				ParentID: &parentID,
				Name:     "GET /users",
				Alias:    "users:list",
			},
		},
	}
	var req dto.RoleResponse
	req.Transform(role)

	assert.Equal(t, req.Name, role.Name)
}
