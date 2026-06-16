package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"workshop/internal/dto"
	"workshop/pkg/errors"

	"github.com/stretchr/testify/require"
)

func ListAccesses(t *testing.T, token string) []dto.AccessTreeResponse {
	var accesses []dto.AccessTreeResponse
	DoRequest(t, RequestConfig[any]{
		Method:         "GET",
		Path:           "/accesses",
		Token:          token,
		Body:           nil,
		ExpectedStatus: http.StatusOK,
		ExpectedCode:   "B1",
		ExpectedMsg:    "Success",
		Validate: func(t *testing.T, data json.RawMessage) *any {
			err := json.Unmarshal(data, &accesses)
			require.NoError(t, err)
			return nil
		},
	})
	return accesses
}

func ListAccessesExpectForbidden(t *testing.T, token string) {
	DoRequest(t, RequestConfig[any]{
		Method:         "GET",
		Path:           "/accesses",
		Token:          token,
		Body:           nil,
		ExpectedStatus: http.StatusForbidden,
		ExpectedCode:   errors.ForbiddenCode,
		ExpectedMsg:    "Forbidden",
		Validate:       nil,
	})
}

func GrantAccess(t *testing.T, token string, roleID, accessID int) {
	DoRequest(t, RequestConfig[any]{
		Method:         "POST",
		Path:           fmt.Sprintf("/roles/%d/access/%d", roleID, accessID),
		Token:          token,
		Body:           nil,
		ExpectedStatus: http.StatusOK,
		ExpectedCode:   "B1",
		ExpectedMsg:    "Success",
		Validate:       nil,
	})
}

func RevokeAccess(t *testing.T, token string, roleID, accessID int) {
	DoRequest(t, RequestConfig[any]{
		Method:         "DELETE",
		Path:           fmt.Sprintf("/roles/%d/access/%d", roleID, accessID),
		Token:          token,
		Body:           nil,
		ExpectedStatus: http.StatusOK,
		ExpectedCode:   "B1",
		ExpectedMsg:    "Success",
		Validate:       nil,
	})
}

func GetAccessIDs(t *testing.T, token string, permissions []string) []int {
	accessIDs := make([]int, 0)
	accesses := ListAccesses(t, token)

	for _, r := range accesses {
		for _, child := range r.Childrens {
			for _, permission := range permissions {
				if child.Alias == permission {
					accessIDs = append(accessIDs, child.ID)
				}
			}
		}
	}
	return accessIDs
}
