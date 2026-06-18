package dto_test

import (
	"testing"
	"workshop/internal/dto"
	"workshop/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestAccessResponse_Transform(t *testing.T) {
	// Test with parent_id
	parentID := 10
	access := model.Access{
		ID:       1,
		ParentID: &parentID,
		Name:     "Create User",
		Alias:    "user:create",
	}

	var resp dto.AccessResponse
	resp.Transform(access)

	assert.Equal(t, 1, resp.ID)
	assert.Equal(t, &parentID, resp.ParentID)
	assert.Equal(t, "Create User", resp.Name)
	assert.Equal(t, "user:create", resp.Alias)

	// Test without parent_id
	access2 := model.Access{
		ID:       2,
		ParentID: nil,
		Name:     "Admin Access",
		Alias:    "admin:access",
	}

	var resp2 dto.AccessResponse
	resp2.Transform(access2)

	assert.Equal(t, 2, resp2.ID)
	assert.Nil(t, resp2.ParentID)
	assert.Equal(t, "Admin Access", resp2.Name)
	assert.Equal(t, "admin:access", resp2.Alias)
}

func TestAccessTreeResponse_Transform(t *testing.T) {
	parentID := 1

	// Create access tree with children
	accessTree := model.AccessTree{
		ID:    1,
		Name:  "User Management",
		Alias: "user:management",
		Childrens: []model.Access{
			{
				ID:       2,
				ParentID: &parentID,
				Name:     "Create User",
				Alias:    "user:create",
			},
			{
				ID:       3,
				ParentID: &parentID,
				Name:     "Delete User",
				Alias:    "user:delete",
			},
		},
	}

	var resp dto.AccessTreeResponse
	resp.Transform(accessTree)

	// Test parent
	assert.Equal(t, 1, resp.ID)
	assert.Equal(t, "User Management", resp.Name)
	assert.Equal(t, "user:management", resp.Alias)

	// Test children
	assert.Len(t, resp.Childrens, 2)

	assert.Equal(t, 2, resp.Childrens[0].ID)
	assert.Equal(t, &parentID, resp.Childrens[0].ParentID)
	assert.Equal(t, "Create User", resp.Childrens[0].Name)
	assert.Equal(t, "user:create", resp.Childrens[0].Alias)

	assert.Equal(t, 3, resp.Childrens[1].ID)
	assert.Equal(t, &parentID, resp.Childrens[1].ParentID)
	assert.Equal(t, "Delete User", resp.Childrens[1].Name)
	assert.Equal(t, "user:delete", resp.Childrens[1].Alias)
}

func TestAccessTreeResponse_Transform_NoChildren(t *testing.T) {
	accessTree := model.AccessTree{
		ID:        2,
		Name:      "Role Management",
		Alias:     "role:management",
		Childrens: []model.Access{},
	}

	var resp dto.AccessTreeResponse
	resp.Transform(accessTree)

	assert.Equal(t, 2, resp.ID)
	assert.Equal(t, "Role Management", resp.Name)
	assert.Equal(t, "role:management", resp.Alias)
	assert.Empty(t, resp.Childrens)
}
