package service

import (
	"context"
	"fmt"
	"log/slog"
	"workshop/config"
	"workshop/internal/model"
	"workshop/internal/repository"
	"workshop/pkg/cache"
	"workshop/pkg/errors"

	"github.com/jacky-htg/go-libs/logger"
)

type Roles interface {
	List(ctx context.Context) ([]model.Role, *errors.BusinessError)
	FindByID(ctx context.Context, id int) (*model.Role, *errors.BusinessError)
	Create(ctx context.Context, role *model.Role) *errors.BusinessError
	Update(ctx context.Context, role *model.Role) *errors.BusinessError
	Delete(ctx context.Context, id int) *errors.BusinessError
	Grant(ctx context.Context, roleID, accessID int) *errors.BusinessError
	Revoke(ctx context.Context, roleID, accessID int) *errors.BusinessError
}

type roles struct {
	cache cache.CacheClient
	log   logger.Logger
	ttl   config.TTLConfig
	repo  repository.RoleRepository
}

func NewRoles(cache cache.CacheClient, log logger.Logger, ttl config.TTLConfig, repo repository.RoleRepository) Roles {
	return &roles{cache: cache, log: log, ttl: ttl, repo: repo}
}

func (u *roles) List(ctx context.Context) ([]model.Role, *errors.BusinessError) {
	cacheKey := "roles::list"
	var list []model.Role
	if err := u.cache.GetJSON(ctx, cacheKey, &list); err == nil {
		return list, nil
	}

	list, err := u.repo.List(ctx)
	if err != nil {
		return nil, errors.InternalServerErrorWrap(err, "error listing roles")
	}

	if err := u.cache.SetJSONWithExpiry(ctx, cacheKey, list, u.ttl.TTLDefault); err != nil {
		u.log.Warn(ctx, "set cache failed", slog.Any("error", err), slog.String("key", cacheKey))
	}
	return list, nil
}

func (u *roles) FindByID(ctx context.Context, id int) (*model.Role, *errors.BusinessError) {
	cacheKey := fmt.Sprintf("roles::%d", id)
	var obj *model.Role
	if err := u.cache.GetJSON(ctx, cacheKey, &obj); err == nil {
		return obj, nil
	}

	obj, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.InternalServerErrorWrap(err, "error finding role")
	}
	if obj == nil {
		return nil, errors.NotFound("role not found")
	}

	if err := u.cache.SetJSONWithExpiry(ctx, cacheKey, obj, u.ttl.TTLDefault); err != nil {
		u.log.Warn(ctx, "set cache failed", slog.Any("error", err), slog.String("key", cacheKey))
	}
	return obj, nil
}

func (u *roles) Create(ctx context.Context, role *model.Role) *errors.BusinessError {
	if _, err := u.cache.Del(ctx, []string{"roles:list"}); err != nil {
		u.log.Error(ctx, "delete cache failed", slog.Any("error", err))
		return errors.InternalServerErrorWrap(err, "error creating role")
	}

	if err := u.repo.Create(ctx, role); err != nil {
		return errors.InternalServerErrorWrap(err, "error creating role")
	}

	return nil
}

func (u *roles) Update(ctx context.Context, role *model.Role) *errors.BusinessError {
	existObj, err := u.repo.FindByID(ctx, role.ID)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error finding role")
	}
	if existObj == nil {
		return errors.NotFound("role not found")
	}

	if _, err := u.cache.Del(ctx, []string{"roles::list", fmt.Sprintf("roles::%d", role.ID)}); err != nil {
		u.log.Error(ctx, "delete cache failed", slog.Any("error", err))
		return errors.InternalServerErrorWrap(err, "error updating role")
	}

	err = u.repo.Update(ctx, role)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error updating role")
	}
	return nil
}

func (u *roles) Delete(ctx context.Context, id int) *errors.BusinessError {
	existObj, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error finding role")
	}
	if existObj == nil {
		return errors.NotFound("role not found")
	}

	if _, err := u.cache.Del(ctx, []string{"roles::list", fmt.Sprintf("roles::%d", id)}); err != nil {
		u.log.Error(ctx, "delete cache failed", slog.Any("error", err))
		return errors.InternalServerErrorWrap(err, "error delete role")
	}

	err = u.repo.Delete(ctx, id)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error delete role")
	}
	return nil
}

func (u *roles) Grant(ctx context.Context, roleID, accessID int) *errors.BusinessError {
	hasAccess, err := u.repo.HasAccess(ctx, roleID, accessID)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error grant access")
	}

	if !hasAccess {
		if _, err := u.cache.Del(ctx, []string{fmt.Sprintf("roles::%d", roleID)}); err != nil {
			u.log.Error(ctx, "delete cache failed", slog.Any("error", err))
			return errors.InternalServerErrorWrap(err, "error grant access")
		}

		err = u.repo.GrantAccess(ctx, roleID, accessID)
		if err != nil {
			return errors.InternalServerErrorWrap(err, "error grant access")
		}
	}
	return nil
}

func (u *roles) Revoke(ctx context.Context, roleID, accessID int) *errors.BusinessError {
	hasAccess, err := u.repo.HasAccess(ctx, roleID, accessID)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error revoke access")
	}

	if hasAccess {
		if _, err := u.cache.Del(ctx, []string{fmt.Sprintf("roles::%d", roleID)}); err != nil {
			u.log.Error(ctx, "delete cache failed", slog.Any("error", err))
			return errors.InternalServerErrorWrap(err, "error revoke access")
		}

		err = u.repo.RevokeAccess(ctx, roleID, accessID)
		if err != nil {
			return errors.InternalServerErrorWrap(err, "error revoke access")
		}
	}
	return nil
}
