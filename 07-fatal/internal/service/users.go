package service

import (
	"workshop/internal/model"
	"workshop/internal/repository"
)

type Users interface {
	List() ([]model.User, error)
}

type users struct {
	repo repository.UserRepository
}

func NewUsers(repo repository.UserRepository) Users {
	return &users{repo: repo}
}

func (u *users) List() ([]model.User, error) {
	return u.repo.List()
}
