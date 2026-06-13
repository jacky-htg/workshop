package service

import (
	"context"
	"database/sql"
	"log/slog"
	"workshop/internal/model"
	"workshop/internal/repository"
	"workshop/pkg/errors"

	"github.com/jacky-htg/go-libs/logger"
	"github.com/jacky-htg/go-libs/uuid7"
	"golang.org/x/crypto/bcrypt"
)

type Users interface {
	List(ctx context.Context) ([]model.User, *errors.BusinessError)
	Create(ctx context.Context, user *model.User) *errors.BusinessError
	FindByID(ctx context.Context, id string) (*model.User, *errors.BusinessError)
	Update(ctx context.Context, user *model.User) *errors.BusinessError
	Delete(ctx context.Context, id string) *errors.BusinessError
}

type users struct {
	db   *sql.DB
	log  logger.Logger
	repo repository.UserRepository
}

func NewUsers(db *sql.DB, log logger.Logger, repo repository.UserRepository) Users {
	return &users{db: db, log: log, repo: repo}
}

func (u *users) List(ctx context.Context) ([]model.User, *errors.BusinessError) {
	users, err := u.repo.List(ctx)
	if err != nil {
		return nil, errors.InternalServerErrorWrap(err, "error listing users")
	}
	return users, nil
}

func (u *users) Create(ctx context.Context, user *model.User) *errors.BusinessError {
	pass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		u.log.Error(ctx, "error generate password", slog.Any("error", err))
		return errors.InternalServerErrorWrap(err, "error generating password")
	}

	user.ID = uuid7.New()
	user.Password = string(pass)

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		u.log.Error(ctx, "error begin tx", slog.Any("error", err))
		return errors.InternalServerErrorWrap(err)
	}
	defer tx.Rollback()

	if err := u.repo.Create(ctx, tx, user); err != nil {
		return errors.InternalServerErrorWrap(err, "error creating user")
	}

	for _, v := range user.Roles {
		if err := u.repo.AssignRole(ctx, tx, user.ID, int64(v.ID)); err != nil {
			return errors.InternalServerErrorWrap(err, "error assign role")
		}
	}

	if err = tx.Commit(); err != nil {
		return errors.InternalServerErrorWrap(err)
	}

	return nil
}

func (u *users) FindByID(ctx context.Context, id string) (*model.User, *errors.BusinessError) {
	user, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.InternalServerErrorWrap(err, "error finding user")
	}
	if user == nil {
		return nil, errors.NotFound("user not found")
	}
	return user, nil
}

func (u *users) Update(ctx context.Context, user *model.User) *errors.BusinessError {
	existUser, err := u.repo.FindByID(ctx, user.ID)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error finding user")
	}
	if existUser == nil {
		return errors.NotFound("user not found")
	}

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		u.log.Error(ctx, "error begin tx", slog.Any("error", err))
		return errors.InternalServerErrorWrap(err)
	}
	defer tx.Rollback()

	err = u.repo.Update(ctx, tx, user)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error updating user")
	}

	mapExistingRoles := make(map[int]model.Role)
	mapNewRoles := make(map[int]model.Role)

	for _, v := range existUser.Roles {
		mapExistingRoles[v.ID] = v
	}

	for _, w := range user.Roles {
		if _, ok := mapExistingRoles[w.ID]; ok {
			delete(mapExistingRoles, w.ID)
		} else {
			mapNewRoles[w.ID] = w
		}
	}

	for _, val := range mapNewRoles {
		if err := u.repo.AssignRole(ctx, tx, user.ID, int64(val.ID)); err != nil {
			return errors.InternalServerErrorWrap(err, "error update assign role")
		}
	}

	for _, val := range mapExistingRoles {
		if err := u.repo.RemoveRole(ctx, tx, user.ID, int64(val.ID)); err != nil {
			return errors.InternalServerErrorWrap(err, "error update assign role")
		}
	}

	if err = tx.Commit(); err != nil {
		return errors.InternalServerErrorWrap(err)
	}

	return nil
}

func (u *users) Delete(ctx context.Context, id string) *errors.BusinessError {
	existUser, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error finding user")
	}
	if existUser == nil {
		return errors.NotFound("user not found")
	}
	err = u.repo.Delete(ctx, id)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error deleting user")
	}
	return nil
}
