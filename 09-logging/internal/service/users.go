package service

import (
	"workshop/internal/model"
	"workshop/internal/repository"

	"github.com/jacky-htg/go-libs/logger"
)

type Users interface {
	List() ([]model.User, error)
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
