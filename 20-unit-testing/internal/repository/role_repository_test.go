package repository_test

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"testing"
	"workshop/internal/model"
	"workshop/internal/repository"
	"workshop/mock/mockpkg"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	expectedRoles := []model.Role{
		{ID: 1, Name: "admin"},
		{ID: 2, Name: "manager"},
	}

	query := `SELECT id, name FROM roles ORDER BY name`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).
			AddRow(expectedRoles[0].ID, expectedRoles[0].Name).
			AddRow(expectedRoles[1].ID, expectedRoles[1].Name),
		)

	ctx := context.Background()
	roles, err := repo.List(ctx)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedRoles), len(roles))
	assert.Equal(t, expectedRoles[0].Name, roles[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	query := `INSERT INTO roles (name) VALUES ($1) RETURNING id`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("admin").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	ctx := context.Background()
	role := &model.Role{Name: "admin"}
	err = repo.Create(ctx, role)

	assert.NoError(t, err)
	assert.Equal(t, 1, role.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	query := `INSERT INTO roles (name) VALUES ($1) RETURNING id`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs("admin").
		WillReturnError(errors.New("duplicate key value violates unique constraint"))

	ctx := context.Background()
	role := &model.Role{Name: "admin"}
	err = repo.Create(ctx, role)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_FindByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	expectedAccesses := []model.Access{
		{ID: 1, Name: "GET /roles", Alias: "roles::list"},
		{ID: 1, Name: "POST /roles", Alias: "roles::create"},
	}
	expectedRole := model.Role{
		ID:       1,
		Name:     "admin",
		Accesses: expectedAccesses,
	}

	accessesJSON, err := json.Marshal(expectedAccesses)
	require.NoError(t, err)

	query := `
		SELECT r.id, r.name, 
			    COALESCE(
					json_agg(
						json_build_object(
							'id', a.id,
							'name', a.name,
							'alias', a.alias
						)
					) FILTER (WHERE a.id IS NOT NULL),
					'[]'::json
				)  AS accesses
		FROM roles r
		LEFT JOIN access_roles ar ON (r.id = ar.role_id)
		LEFT JOIN access a ON (ar.access_id = a.id) 
		WHERE r.id = $1 GROUP BY r.id, r.name`

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "accesses"}).
			AddRow(expectedRole.ID, expectedRole.Name, accessesJSON))

	ctx := context.Background()
	role, err := repo.FindByID(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedRole.ID, role.ID)
	assert.Equal(t, expectedRole.Name, role.Name)
	assert.Equal(t, len(expectedRole.Accesses), len(role.Accesses))
	assert.Equal(t, expectedRole.Accesses[0].Alias, role.Accesses[0].Alias)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	query := `UPDATE roles SET name = $1 WHERE id = $2`
	mock.ExpectExec(regexp.QuoteMeta(query)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	role := &model.Role{Name: "admin"}
	err = repo.Update(ctx, role)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
