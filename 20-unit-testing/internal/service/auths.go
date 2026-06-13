package service

import (
	"context"
	"workshop/config"
	"workshop/internal/model"
	"workshop/internal/repository"
	"workshop/pkg/errors"

	"github.com/jacky-htg/go-libs/logger"
	"github.com/jacky-htg/go-libs/token"
	"golang.org/x/crypto/bcrypt"
)

type Auths interface {
	Login(ctx context.Context, email, password string) (string, *model.User, []string, *errors.BusinessError)
}

type auths struct {
	log      logger.Logger
	repo     repository.UserRepository
	roleRepo repository.RoleRepository
	cfgToken config.TokenConfig
}

func NewAuths(log logger.Logger, cfgToken config.TokenConfig, repo repository.UserRepository, roleRepo repository.RoleRepository) Auths {
	return &auths{log: log, repo: repo, roleRepo: roleRepo}
}

func (u *auths) Login(ctx context.Context, email, password string) (string, *model.User, []string, *errors.BusinessError) {
	list := make([]string, 0)

	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", nil, list, errors.InternalServerErrorWrap(err, "error finding user")
	}
	if user == nil {
		return "", nil, list, errors.InvalidInput("Invalid username/password")
	}

	var roleIDs []int = make([]int, 0)
	for _, val := range user.Roles {
		roleIDs = append(roleIDs, val.ID)
	}

	accesses, err := u.roleRepo.GetAccessesByRoles(ctx, roleIDs)
	if err != nil {
		return "", nil, list, errors.InternalServerErrorWrap(err, "error finding user")
	}

	for _, val := range accesses {
		list = append(list, val.Alias)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		u.log.Error(ctx, "Invalid username/password")
		return "", nil, list, errors.InvalidInput("Invalid username/password")
	}

	if !user.IsActive {
		return "", nil, list, errors.Forbidden("user inavtive")
	}

	myToken, err := token.ClaimToken(map[string]any{
		"email": user.Email,
		"id":    user.ID,
	}, u.cfgToken.TokenExp)

	if err != nil {
		u.log.Error(ctx, "claim token")
		return "", nil, list, errors.InternalServerErrorWrap(err)
	}

	return myToken, user, list, nil
}
