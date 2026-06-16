//go:build integration

package rbac_test

import (
	"net/http"
	"testing"
	"workshop/test/e2e/helper"

	"github.com/stretchr/testify/require"
)

func TestRBAC(t *testing.T) {
	const (
		adminEmail = "admin@example.com"
		adminPass  = "1234"
		userEmail  = "manager@example.com"
		userPass   = "1234567890"
		userName   = "manger"
		roleName   = "manager"
	)

	var (
		permissions = []string{
			"accesses::list",
			"roles::list",
			"roles::view",
		}
	)

	// Step 1: Login with non-existent user (negative test)
	helper.LoginExpectError(t, userEmail, userPass, http.StatusBadRequest, "E001", "Invalid username/password")

	// Step 2: Login as admin
	tokenAdmin := helper.Login(t, adminEmail, adminPass)

	// Step 3: Verify role does NOT exist
	require.False(t, helper.RoleExists(t, tokenAdmin, roleName), "Role should not exist yet")

	// Step 4: Create role
	role := helper.CreateRole(t, tokenAdmin, roleName)
	t.Logf("✅ Role created: ID=%d, Name=%s", role.ID, role.Name)

	// Step 5: Verify role by ID
	gotRole := helper.GetRole(t, tokenAdmin, role.ID)
	require.Equal(t, role.Name, gotRole.Name)

	// Step 6: Create user
	user := helper.CreateUser(t, tokenAdmin, userName, userEmail, userPass, []int{role.ID})
	t.Logf("✅ User created: ID=%s", user.ID)

	// Step 7: Verify user by ID
	gotUser := helper.GetUser(t, tokenAdmin, user.ID)
	require.Equal(t, user.Email, gotUser.Email)

	// Step 8: Login as new user (no permissions yet)
	tokenUser := helper.Login(t, userEmail, userPass)

	// Step 9: Access /accesses without permission (should be forbidden)
	helper.ListAccessesExpectForbidden(t, tokenUser)

	// Step 10: Get permission IDs
	permissionIDs := helper.GetAccessIDs(t, tokenAdmin, permissions)
	require.Equal(t, len(permissions), len(permissionIDs), "Length of Permission IDs should be 3")

	// Step 11: Grant access to role
	for _, accessID := range permissionIDs {
		helper.GrantAccess(t, tokenAdmin, role.ID, accessID)
	}

	// Step 12: Access /accesses with permission (should succeed)
	accesses := helper.ListAccesses(t, tokenUser)
	require.NotEmpty(t, accesses, "User should have permissions after grant")
	t.Logf("✅ User now has %d permissions", len(accesses))

	// Cleanup (auto cleanup dengan t.Cleanup)
	t.Cleanup(func() {
		helper.DeleteUser(t, tokenAdmin, user.ID)
		helper.DeleteRole(t, tokenAdmin, role.ID)
		t.Log("✅ Cleanup completed")
	})
}
