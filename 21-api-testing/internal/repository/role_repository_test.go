package repository_test

import (
	"context"
	"database/sql"
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

const (
	qRoleCreate   = `INSERT INTO roles (name) VALUES ($1) RETURNING id`
	qRoleFindByID = `
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
	qRoleList               = `SELECT id, name FROM roles ORDER BY name`
	qRoleUpdate             = `UPDATE roles SET name = $1 WHERE id = $2`
	qRoleDelete             = `DELETE FROM roles WHERE id = $1`
	qRoleGrant              = `INSERT INTO access_roles (access_id, role_id) VALUES ($1, $2)`
	qRoleRevoke             = `DELETE FROM access_roles WHERE access_id = $1 AND role_id = $2`
	qRoleGetAccessesByRoles = `
                SELECT DISTINCT a.id, a.parent_id, a.alias 
                FROM roles r
                JOIN access_roles ar ON (r.id = ar.role_id)
                JOIN access a ON (ar.access_id = a.id) 
                WHERE r.id = ANY($1)
                ORDER BY a.parent_id, a.alias`
	qRoleHasAccess = `SELECT true FROM access_roles WHERE role_id = $1 AND access_id = $2`
)

func TestRoleRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectQuery(regexp.QuoteMeta(qRoleCreate)).
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

	mock.ExpectQuery(regexp.QuoteMeta(qRoleCreate)).
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

	mock.ExpectQuery(regexp.QuoteMeta(qRoleFindByID)).
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

func TestRoleRepository_FindByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectQuery(regexp.QuoteMeta(qRoleFindByID)).
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	role, err := repo.FindByID(ctx, 1)

	assert.NoError(t, err)
	assert.Nil(t, role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_FindByID_ScanError(t *testing.T) {
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
		Name: "admin",
	}

	accessesJSON, err := json.Marshal(expectedAccesses)
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(qRoleFindByID)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "accesses"}).
			AddRow("invalid-ID", expectedRole.Name, accessesJSON))

	ctx := context.Background()
	role, err := repo.FindByID(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "sql: Scan error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_FindByID_UnmarshalError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	expectedRole := model.Role{
		ID:   1,
		Name: "admin",
	}

	accessesJSON := []byte(`{"id":1)`) // invalid closing bracket

	mock.ExpectQuery(regexp.QuoteMeta(qRoleFindByID)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "accesses"}).
			AddRow(expectedRole.ID, expectedRole.Name, accessesJSON))

	ctx := context.Background()
	role, err := repo.FindByID(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, role)
	assert.Contains(t, err.Error(), "invalid character ')'")
	assert.NoError(t, mock.ExpectationsWereMet())
}

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

	mock.ExpectQuery(regexp.QuoteMeta(qRoleList)).
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

func TestRoleRepository_List_NotFoundError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectQuery(regexp.QuoteMeta(qRoleList)).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	roles, err := repo.List(ctx)

	assert.Error(t, err)
	assert.Nil(t, roles)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_List_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow("not_a_number", "admin") // ID seharusnya int, tapi dikirim string
	mock.ExpectQuery(regexp.QuoteMeta(qRoleList)).
		WillReturnRows(rows)

	ctx := context.Background()
	roles, err := repo.List(ctx)

	assert.Error(t, err)
	assert.Nil(t, roles)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_List_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	// Simulasikan error saat iterasi row
	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "admin").
		AddRow(2, "manager").
		RowError(1, errors.New("database connection lost")) // Error pada row ke-2

	mock.ExpectQuery(regexp.QuoteMeta(qRoleList)).WillReturnRows(rows)

	ctx := context.Background()
	roles, err := repo.List(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, roles)
	assert.Contains(t, err.Error(), "database connection lost")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_Update_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qRoleUpdate)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	role := &model.Role{Name: "admin"}
	err = repo.Update(ctx, role)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_Update_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qRoleUpdate)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	role := &model.Role{Name: "admin"}
	err = repo.Update(ctx, role)

	assert.Error(t, err)
	assert.Equal(t, err, sql.ErrConnDone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qRoleDelete)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err = repo.Delete(ctx, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qRoleDelete)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	err = repo.Delete(ctx, 1)

	assert.Error(t, err)
	assert.Equal(t, err, sql.ErrConnDone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_GrantAccess_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qRoleGrant)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err = repo.GrantAccess(ctx, 1, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_GrantAccess_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qRoleGrant)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	err = repo.GrantAccess(ctx, 1, 1)

	assert.Error(t, err)
	assert.Equal(t, err, sql.ErrConnDone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_RevokeAccess_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qRoleRevoke)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err = repo.RevokeAccess(ctx, 1, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_RevokeAccess_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectExec(regexp.QuoteMeta(qRoleRevoke)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	err = repo.RevokeAccess(ctx, 1, 1)

	assert.Error(t, err)
	assert.Equal(t, err, sql.ErrConnDone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_GetAccessesByRoles_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	parentID := 1
	expected := []model.Access{
		{ID: 11, ParentID: &parentID, Alias: "users:list"},
		{ID: 12, ParentID: &parentID, Alias: "roles:create"},
	}

	mock.ExpectQuery(regexp.QuoteMeta(qRoleGetAccessesByRoles)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id", "name"}).
			AddRow(expected[0].ID, expected[0].ParentID, expected[0].Alias).
			AddRow(expected[1].ID, expected[0].ParentID, expected[0].Alias),
		)

	ctx := context.Background()
	accesses, err := repo.GetAccessesByRoles(ctx, []int{1})

	assert.NoError(t, err)
	assert.Equal(t, len(expected), len(accesses))
	assert.Equal(t, expected[0].Alias, accesses[0].Alias)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_GetAccessesByRoles_RoleIDsEmpty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	ctx := context.Background()
	accesses, err := repo.GetAccessesByRoles(ctx, []int{})

	assert.NoError(t, err)
	assert.Equal(t, []model.Access{}, accesses)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_GetAccessesByRoles_NotFoundError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectQuery(regexp.QuoteMeta(qRoleGetAccessesByRoles)).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	accesses, err := repo.GetAccessesByRoles(ctx, []int{1})

	assert.Error(t, err)
	assert.Nil(t, accesses)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_GetAccessesByRoles_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	rows := sqlmock.NewRows([]string{"id", "parent_id", "name"}).
		AddRow("not_a_number", 1, "admin") // ID seharusnya int, tapi dikirim string
	mock.ExpectQuery(regexp.QuoteMeta(qRoleGetAccessesByRoles)).
		WillReturnRows(rows)

	ctx := context.Background()
	accesses, err := repo.GetAccessesByRoles(ctx, []int{1})

	assert.Error(t, err)
	assert.Nil(t, accesses)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_GetAccessesByRoles_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	// Simulasikan error saat iterasi row
	rows := sqlmock.NewRows([]string{"id", "parent_id", "name"}).
		AddRow(11, 1, "users:view").
		AddRow(12, 1, "roles:create").
		RowError(1, errors.New("database connection lost")) // Error pada row ke-2

	mock.ExpectQuery(regexp.QuoteMeta(qRoleGetAccessesByRoles)).WillReturnRows(rows)

	ctx := context.Background()
	accesses, err := repo.GetAccessesByRoles(ctx, []int{1})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, accesses)
	assert.Contains(t, err.Error(), "database connection lost")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_HasAccess_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectQuery(regexp.QuoteMeta(qRoleHasAccess)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"has_access"}).AddRow(true))

	ctx := context.Background()
	hasAccess, err := repo.HasAccess(ctx, 1, 1)

	assert.NoError(t, err)
	assert.True(t, hasAccess)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_HasAccess_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectQuery(regexp.QuoteMeta(qRoleHasAccess)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	hasAccess, err := repo.HasAccess(ctx, 1, 1)

	assert.NoError(t, err)
	assert.False(t, hasAccess)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRoleRepository_HasAccess_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewRoleRepository(db, log)

	mock.ExpectQuery(regexp.QuoteMeta(qRoleHasAccess)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"has_access"}).AddRow("invalid-bool"))

	ctx := context.Background()
	hasAccess, err := repo.HasAccess(ctx, 1, 1)

	assert.Error(t, err)
	assert.False(t, hasAccess)
	assert.Contains(t, err.Error(), "sql: Scan error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
