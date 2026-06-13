package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"workshop/internal/model"

	"github.com/jacky-htg/go-libs/logger"
	"github.com/lib/pq"
)

type RoleRepository interface {
	// Basic CRUD
	Create(ctx context.Context, role *model.Role) error
	FindByID(ctx context.Context, id int) (*model.Role, error)
	List(ctx context.Context) ([]model.Role, error)
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id int) error

	// Many-to-many dengan Access
	GrantAccess(ctx context.Context, roleID, accessID int) error
	RevokeAccess(ctx context.Context, roleID, accessID int) error
	GetAccessesByRoles(ctx context.Context, roleIDs []int) ([]model.Access, error)

	// Helper
	HasAccess(ctx context.Context, roleID, accessID int) (bool, error)
}

type roleRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewRoleRepository(db *sql.DB, log logger.Logger) RoleRepository {
	return &roleRepository{db: db, log: log}
}

func (u *roleRepository) Create(ctx context.Context, role *model.Role) error {
	query := `INSERT INTO roles (name) VALUES ($1) RETURNING id`
	err := u.db.QueryRowContext(ctx, query, role.Name).Scan(&role.ID)
	if err != nil {
		u.log.Error(ctx, "error: inserting role", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *roleRepository) FindByID(ctx context.Context, id int) (*model.Role, error) {
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

	row := u.db.QueryRowContext(ctx, query, id)

	var obj model.Role
	var accessesJSON []byte
	if err := row.Scan(&obj.ID, &obj.Name, &accessesJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		u.log.Error(ctx, "error: scanning role row", slog.Any("error", err))
		return nil, err
	}

	err := json.Unmarshal(accessesJSON, &obj.Accesses)
	if err != nil {
		u.log.Error(ctx, "error: unmarshall accesses", slog.Any("error", err))
		return nil, err
	}

	return &obj, nil
}

func (u *roleRepository) List(ctx context.Context) ([]model.Role, error) {
	query := `SELECT id, name FROM roles ORDER BY name`
	rows, err := u.db.QueryContext(ctx, query)
	if err != nil {
		u.log.Error(ctx, "error: querying roles", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	var list []model.Role = make([]model.Role, 0)
	for rows.Next() {
		var obj model.Role
		if err := rows.Scan(&obj.ID, &obj.Name); err != nil {
			u.log.Error(ctx, "error: scanning roles row", slog.Any("error", err))
			return nil, err
		}
		list = append(list, obj)
	}

	if err := rows.Err(); err != nil {
		u.log.Error(ctx, "error: iterating roles rows", slog.Any("error", err))
		return nil, err
	}

	return list, nil
}

func (u *roleRepository) Update(ctx context.Context, role *model.Role) error {
	query := `UPDATE roles SET name = $1 WHERE id = $2`
	_, err := u.db.ExecContext(ctx, query, role.Name, role.ID)
	if err != nil {
		u.log.Error(ctx, "error: updating role", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *roleRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM roles WHERE id = $1`
	_, err := u.db.ExecContext(ctx, query, id)
	if err != nil {
		u.log.Error(ctx, "error: delete role", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *roleRepository) GrantAccess(ctx context.Context, roleID, accessID int) error {
	query := `INSERT INTO access_roles (access_id, role_id) VALUES ($1, $2)`
	_, err := u.db.ExecContext(ctx, query, accessID, roleID)
	if err != nil {
		u.log.Error(ctx, "error: grant access", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *roleRepository) RevokeAccess(ctx context.Context, roleID, accessID int) error {
	query := `DELETE FROM access_roles WHERE access_id = $1 AND role_id = $2`
	_, err := u.db.ExecContext(ctx, query, accessID, roleID)
	if err != nil {
		u.log.Error(ctx, "error: grant access", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *roleRepository) GetAccessesByRoles(ctx context.Context, roleIDs []int) ([]model.Access, error) {
	var list []model.Access = make([]model.Access, 0)

	if len(roleIDs) == 0 {
		return list, nil
	}
	query := `
		SELECT DISTINCT a.id, a.parent_id, a.alias 
		FROM roles r
		JOIN access_roles ar ON (r.id = ar.role_id)
		JOIN access a ON (ar.access_id = a.id) 
		WHERE r.id = ANY($1)
		ORDER BY a.parent_id, a.alias`

	rows, err := u.db.QueryContext(ctx, query, pq.Array(roleIDs))
	if err != nil {
		u.log.Error(ctx, "error: querying get access by role", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var obj model.Access
		if err := rows.Scan(&obj.ID, &obj.ParentID, &obj.Alias); err != nil {
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

func (u *roleRepository) HasAccess(ctx context.Context, roleID, accessID int) (bool, error) {
	query := `SELECT true FROM access_roles WHERE role_id = $1 AND access_id = $2`
	row := u.db.QueryRowContext(ctx, query, roleID, accessID)

	var hasAccess bool
	if err := row.Scan(&hasAccess); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		u.log.Error(ctx, "error: scanning role row", slog.Any("error", err))
		return false, err
	}

	return hasAccess, nil
}
