package service

import (
	"context"
	"fmt"
	"log/slog"
	"workshop/internal/model"
	"workshop/internal/repository"

	"github.com/jacky-htg/go-libs/logger"
	"github.com/jacky-htg/go-libs/uuid7"
	"golang.org/x/crypto/bcrypt"
)

type Users interface {
	List() ([]model.User, error)
	Create(*model.User) error
	FindById(id string) (*model.User, error)
	Update(*model.User) error
	Delete(id string) error
}

type users struct {
	log  logger.Logger
	repo repository.UserRepository
}

func NewUsers(repo repository.UserRepository, log logger.Logger) Users {
	return &users{repo: repo, log: log}
}

func (u *users) List() ([]model.User, error) {
	return u.repo.List()
}

func (u *users) Create(user *model.User) error {

	pass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		u.log.Error(context.Background(), "error generate password", slog.Any("error", err))
		return err
	}

	user.ID = uuid7.New()
	user.Password = string(pass)

	if err := u.repo.Create(user); err != nil {
		return err
	}

	return nil
}

func (u *users) FindById(id string) (*model.User, error) {
	return u.repo.FindById(id)
}

func (u *users) Update(user *model.User) error {
	existUser, err := u.repo.FindById(user.ID)
	if err != nil {
		return err
	}
	if existUser == nil {
		return fmt.Errorf("user not found")
	}
	return u.repo.Update(user)
}

func (u *users) Delete(id string) error {
	existUser, err := u.repo.FindById(id)
	if err != nil {
		return err
	}
	if existUser == nil {
		return fmt.Errorf("user not found")
	}
	return u.repo.Delete(id)
}
