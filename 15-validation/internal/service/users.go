package service

import (
	"context"
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
	FindById(ctx context.Context, id string) (*model.User, *errors.BusinessError)
	Update(ctx context.Context, user *model.User) *errors.BusinessError
	Delete(ctx context.Context, id string) *errors.BusinessError
}

type users struct {
	log  logger.Logger
	repo repository.UserRepository
}

func NewUsers(log logger.Logger, repo repository.UserRepository) Users {
	return &users{log: log, repo: repo}
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

	if err := u.repo.Create(ctx, user); err != nil {
		return errors.InternalServerErrorWrap(err, "error creating user")
	}

	return nil
}

func (u *users) FindById(ctx context.Context, id string) (*model.User, *errors.BusinessError) {
	user, err := u.repo.FindById(ctx, id)
	if err != nil {
		return nil, errors.InternalServerErrorWrap(err, "error finding user")
	}
	if user == nil {
		return nil, errors.NotFound("user not found")
	}
	return user, nil
}

func (u *users) Update(ctx context.Context, user *model.User) *errors.BusinessError {
	existUser, err := u.repo.FindById(ctx, user.ID)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error finding user")
	}
	if existUser == nil {
		return errors.NotFound("user not found")
	}
	err = u.repo.Update(ctx, user)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error updating user")
	}
	return nil
}

func (u *users) Delete(ctx context.Context, id string) *errors.BusinessError {
	existUser, err := u.repo.FindById(ctx, id)
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
