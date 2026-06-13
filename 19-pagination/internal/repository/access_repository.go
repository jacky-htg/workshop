package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"workshop/internal/model"

	"github.com/jacky-htg/go-libs/logger"
)

type AccessRepository interface {
	List(ctx context.Context) ([]model.Access, error)
	Create(ctx context.Context, tx *sql.Tx, access *model.Access) error
}

type accessRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewAccessRepository(db *sql.DB, log logger.Logger) AccessRepository {
	return &accessRepository{db: db, log: log}
}

func (u *accessRepository) List(ctx context.Context) ([]model.Access, error) {
	query := `SELECT id, parent_id, name, alias FROM access WHERE alias != 'root' ORDER BY parent_id, name`
	rows, err := u.db.QueryContext(ctx, query)
	if err != nil {
		u.log.Error(ctx, "error: querying access", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	var list []model.Access = make([]model.Access, 0)
	for rows.Next() {
		var obj model.Access
		if err := rows.Scan(&obj.ID, &obj.ParentID, &obj.Name, &obj.Alias); err != nil {
			u.log.Error(ctx, "error: scanning access row", slog.Any("error", err))
			return nil, err
		}
		list = append(list, obj)
	}

	if err := rows.Err(); err != nil {
		u.log.Error(ctx, "error: iterating access rows", slog.Any("error", err))
		return nil, err
	}

	return list, nil
}

func (u *accessRepository) Create(ctx context.Context, tx *sql.Tx, access *model.Access) error {
	query := `
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
	err := tx.QueryRowContext(ctx, query, access.ParentID, access.Name, access.Alias).Scan(&access.ID)
	if err != nil {
		fmt.Println(query, query, *access.ParentID, access.Name, access.Alias)
		u.log.Error(ctx, "error: inserting access", slog.Any("error", err))
		return err
	}

	return nil
}
