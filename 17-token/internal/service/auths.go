package service

import (
	"context"
	"workshop/config"
	"workshop/internal/repository"
	"workshop/pkg/errors"

	"github.com/jacky-htg/go-libs/logger"
	"github.com/jacky-htg/go-libs/token"
	"golang.org/x/crypto/bcrypt"
)

type Auths interface {
	Login(ctx context.Context, email, password string) (string, *errors.BusinessError)
}

type auths struct {
	log      logger.Logger
	repo     repository.UserRepository
	cfgToken config.TokenConfig
}

func NewAuths(log logger.Logger, cfgToken config.TokenConfig, repo repository.UserRepository) Auths {
	return &auths{log: log, repo: repo}
}

func (u *auths) Login(ctx context.Context, email, password string) (string, *errors.BusinessError) {
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", errors.InternalServerErrorWrap(err, "error finding user")
	}
	if user == nil {
		return "", errors.InvalidInput("Invalid username/password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		u.log.Error(ctx, "Invalid username/password")
		return "", errors.InvalidInput("Invalid username/password")
	}

	if !user.IsActive {
		return "", errors.Forbidden("user inavtive")
	}

	myToken, err := token.ClaimToken(map[string]any{
		"email": user.Email,
	}, u.cfgToken.TokenExp)

	if err != nil {
		u.log.Error(ctx, "claim token")
		return "", errors.InternalServerErrorWrap(err)
	}

	return myToken, nil
}
