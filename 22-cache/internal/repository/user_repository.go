package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"workshop/internal/model"

	"github.com/jacky-htg/go-libs/logger"
)

type UserRepository interface {
	List(ctx context.Context, search, order, sort string, limit, offset int) ([]model.User, int, error)
	Create(ctx context.Context, tx *sql.Tx, user *model.User) error
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Update(ctx context.Context, tx *sql.Tx, user *model.User) error
	Delete(ctx context.Context, id string) error

	// Manage roles untuk user
	AssignRole(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error
	RemoveRole(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error

	// Check permission
	HasPermission(ctx context.Context, email, routePath, routeGroup string) bool
}

type userRepository struct {
	db  *sql.DB
	log logger.Logger
}

func NewUserRepository(db *sql.DB, log logger.Logger) UserRepository {
	return &userRepository{db: db, log: log}
}

// List : http handler for returning list of users
func (u *userRepository) List(ctx context.Context, search, order, sort string, limit, offset int) ([]model.User, int, error) {
	conditions := []string{"deleted_at IS NULL"}
	args := []any{}

	if len(search) > 0 {
		conditions = append(conditions, fmt.Sprintf(`(name ILIKE $%d)`, len(args)+1))
		args = append(args, "%"+search+"%")
	}

	conditionStr := strings.Join(conditions, " AND ")

	var count int
	err := u.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE `+conditionStr, args...).Scan(&count)
	if err != nil {
		u.log.Error(ctx, "error: querying count users", slog.Any("error", err))
		return nil, count, err
	}

	orderByMap := map[string]string{
		"id":         "id",
		"name":       "LOWER(name)",     // Case-insensitive
		"username":   "LOWER(username)", // Case-insensitive
		"email":      "LOWER(email)",    // Case-insensitive
		"is_active":  "is_active",
		"created_at": "created_at",
	}

	order, ok := orderByMap[order]
	if !ok {
		order = "id"
	}

	query := `SELECT id, name, username, password, email, is_active FROM users WHERE ` + conditionStr
	query += fmt.Sprintf(` ORDER BY %s %s LIMIT %d OFFSET %d`, order, sort, limit, offset)

	rows, err := u.db.QueryContext(ctx, query, args...)
	if err != nil {
		u.log.Error(ctx, "error: querying users", slog.Any("error", err))
		return nil, count, err
	}
	defer rows.Close()

	var users []model.User = make([]model.User, 0)
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Username, &user.Password, &user.Email, &user.IsActive); err != nil {
			u.log.Error(ctx, "error: scanning user row", slog.Any("error", err))
			return nil, count, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		u.log.Error(ctx, "error: iterating user rows", slog.Any("error", err))
		return nil, count, err
	}

	return users, count, nil
}

func (u *userRepository) Create(ctx context.Context, tx *sql.Tx, user *model.User) error {
	query := `INSERT INTO users (id, name, username, password, email, is_active) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := tx.ExecContext(ctx, query, user.ID, user.Name, user.Username, user.Password, user.Email, user.IsActive)
	if err != nil {
		u.log.Error(ctx, "error: inserting user", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `
			SELECT u.id, u.name, u.username, u.password, u.email, u.is_active, 
					COALESCE(
						json_agg(
							json_build_object(
								'id', r.id,
								'name', r.name
							)
						) FILTER (WHERE r.id IS NOT NULL),
						'[]'::json
					)  AS roles
			FROM users u
			LEFT JOIN roles_users ru ON (u.id = ru.user_id)
			LEFT JOIN roles r ON (ru.role_id = r.id) 
			WHERE u.id = $1 AND u.deleted_at IS NULL 
			GROUP BY u.id, u.name, u.username, u.password, u.email, u.is_active`
	row := u.db.QueryRowContext(ctx, query, id)

	var user model.User
	var rolesJSON []byte
	if err := row.Scan(&user.ID, &user.Name, &user.Username, &user.Password, &user.Email, &user.IsActive, &rolesJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		u.log.Error(ctx, "error: scanning user row", slog.Any("error", err))
		return nil, err
	}

	if err := json.Unmarshal(rolesJSON, &user.Roles); err != nil {
		u.log.Error(ctx, "error: unmarshall roles", slog.Any("error", err))
		return nil, err
	}

	return &user, nil
}

func (u *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
			SELECT u.id, u.name, u.username, u.password, u.email, u.is_active, 
					COALESCE(
						json_agg(
							json_build_object(
								'id', r.id,
								'name', r.name
							)
						) FILTER (WHERE r.id IS NOT NULL),
						'[]'::json
					)  AS roles
			FROM users u
			LEFT JOIN roles_users ru ON (u.id = ru.user_id)
			LEFT JOIN roles r ON (ru.role_id = r.id) 
			WHERE u.email = $1 AND u.deleted_at IS NULL 
			GROUP BY u.id, u.name, u.username, u.password, u.email, u.is_active`
	row := u.db.QueryRowContext(ctx, query, email)

	var user model.User
	var rolesJSON []byte
	if err := row.Scan(&user.ID, &user.Name, &user.Username, &user.Password, &user.Email, &user.IsActive, &rolesJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		u.log.Error(ctx, "error: scanning user row", slog.Any("error", err))
		return nil, err
	}

	if err := json.Unmarshal(rolesJSON, &user.Roles); err != nil {
		u.log.Error(ctx, "error: unmarshall roles", slog.Any("error", err))
		return nil, err
	}

	return &user, nil
}

func (u *userRepository) Update(ctx context.Context, tx *sql.Tx, user *model.User) error {
	query := `UPDATE users SET name = $1, is_active = $2 WHERE id = $3 RETURNING username, email`
	err := tx.QueryRowContext(ctx, query, user.Name, user.IsActive, user.ID).Scan(&user.Username, &user.Email)
	if err != nil {
		u.log.Error(ctx, "error: updating user", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *userRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET deleted_at = timezone('utc', now()) WHERE id = $1`
	_, err := u.db.ExecContext(ctx, query, id)
	if err != nil {
		u.log.Error(ctx, "error: deleting user", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *userRepository) AssignRole(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
	query := `INSERT INTO roles_users (role_id, user_id) VALUES ($1, $2)`
	_, err := tx.ExecContext(ctx, query, roleID, userID)
	if err != nil {
		u.log.Error(ctx, "error: assign role", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *userRepository) RemoveRole(ctx context.Context, tx *sql.Tx, userID string, roleID int64) error {
	query := `DELETE FROM roles_users WHERE role_id = $1 AND user_id = $2`
	_, err := tx.ExecContext(ctx, query, roleID, userID)
	if err != nil {
		u.log.Error(ctx, "error: remove role", slog.Any("error", err))
		return err
	}

	return nil
}

func (u *userRepository) HasPermission(ctx context.Context, email, routePath, routeGroup string) bool {

	query := `
			SELECT EXISTS(
				SELECT true 
				FROM users u
				JOIN roles_users ru ON (u.id = ru.user_id)
				JOIN roles r ON (ru.role_id = r.id)
				JOIN access_roles ar ON (r.id = ar.role_id) 
				JOIN access a ON (ar.access_id = a.id)
				WHERE u.email = $1 AND (a.name = $2 OR a.name = $3 OR a.name = 'root') 
			)`

	var hasPermission bool = false
	err := u.db.QueryRowContext(ctx, query, email, routePath, routeGroup).Scan(&hasPermission)
	if err != nil {
		u.log.Error(ctx, "error: has permission", slog.Any("error", err))
		return false
	}
	return hasPermission
}
