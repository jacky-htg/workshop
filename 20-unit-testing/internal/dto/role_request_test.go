package dto_test

import (
	"testing"
	"workshop/internal/dto"
	"workshop/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestRoleRequest_Transform(t *testing.T) {
	role := &model.Role{}
	req := dto.RoleRequest{
		Name: "admin",
	}
	req.Transform(role)

	assert.Equal(t, req.Name, role.Name)
}
