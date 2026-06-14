package repository_test

import (
	"context"
	"database/sql"
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
	qAccessList   = `SELECT id, parent_id, name, alias FROM access WHERE alias != 'root' ORDER BY parent_id, name`
	qAccessCreate = `
					WITH inserted AS (
						INSERT INTO access (parent_id, name, alias) 
						VALUES ($1, $2, $3)
						ON CONFLICT (name) DO NOTHING
						RETURNING id
					)
					SELECT id FROM inserted
					UNION ALL
					SELECT id FROM access WHERE name = $2
					LIMIT 1`
)

func TestAccessRepository_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewAccessRepository(db, log)

	parentID := 1
	expected := []model.Access{
		{ID: 11, ParentID: &parentID, Name: "GET /users", Alias: "users:list"},
		{ID: 12, ParentID: &parentID, Name: "POST /users", Alias: "users:create"},
	}

	mock.ExpectQuery(regexp.QuoteMeta(qAccessList)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id", "name", "alias"}).
			AddRow(expected[0].ID, expected[0].ParentID, expected[0].Name, expected[0].Alias).
			AddRow(expected[1].ID, expected[1].ParentID, expected[1].Name, expected[1].Alias),
		)

	ctx := context.Background()
	accesss, err := repo.List(ctx)

	assert.NoError(t, err)
	assert.Equal(t, len(expected), len(accesss))
	assert.Equal(t, expected[0].Name, accesss[0].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAccessRepository_List_NotFoundError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewAccessRepository(db, log)

	mock.ExpectQuery(regexp.QuoteMeta(qAccessList)).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	accesss, err := repo.List(ctx)

	assert.Error(t, err)
	assert.Nil(t, accesss)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAccessRepository_List_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewAccessRepository(db, log)

	rows := sqlmock.NewRows([]string{"id", "parent_id", "name", "alias"}).
		AddRow("not_a_number", 1, "GET /users", "users:list") // ID seharusnya int, tapi dikirim string
	mock.ExpectQuery(regexp.QuoteMeta(qAccessList)).
		WillReturnRows(rows)

	ctx := context.Background()
	accesss, err := repo.List(ctx)

	assert.Error(t, err)
	assert.Nil(t, accesss)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAccessRepository_List_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewAccessRepository(db, log)

	// Simulasikan error saat iterasi row
	rows := sqlmock.NewRows([]string{"id", "parent_id", "name", "alias"}).
		AddRow(11, 1, "GET /users", "users:list").
		AddRow(12, 1, "POST /users", "users:create").
		RowError(1, errors.New("database connection lost")) // Error pada row ke-2

	mock.ExpectQuery(regexp.QuoteMeta(qAccessList)).WillReturnRows(rows)

	ctx := context.Background()
	accesss, err := repo.List(ctx)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, accesss)
	assert.Contains(t, err.Error(), "database connection lost")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAccessRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewAccessRepository(db, log)

	ctx := context.Background()
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	parentID := 1
	access := &model.Access{ParentID: &parentID, Name: "root", Alias: "root"}
	mock.ExpectQuery(regexp.QuoteMeta(qAccessCreate)).
		WithArgs(access.ParentID, access.Name, access.Alias).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	err = repo.Create(ctx, tx, access)

	assert.NoError(t, err)
	assert.Equal(t, 1, access.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAccessRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	log := mockpkg.NewMockLogger()
	repo := repository.NewAccessRepository(db, log)

	ctx := context.Background()
	mock.ExpectBegin()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	parentID := 1
	access := &model.Access{ParentID: &parentID, Name: "root", Alias: "root"}
	mock.ExpectQuery(regexp.QuoteMeta(qAccessCreate)).
		WithArgs(access.ParentID, access.Name, access.Alias).
		WillReturnError(errors.New("duplicate key value violates unique constraint"))

	err = repo.Create(ctx, tx, access)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
