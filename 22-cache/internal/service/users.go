package service

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"workshop/config"
	"workshop/internal/model"
	"workshop/internal/repository"
	"workshop/pkg/cache"
	"workshop/pkg/errors"
	"workshop/pkg/listcache"

	"github.com/jacky-htg/go-libs/logger"
	"github.com/jacky-htg/go-libs/uuid7"
	"golang.org/x/crypto/bcrypt"
)

const (
	usersListPrefix = "users::list::"
	usersIndexKey   = "users::list::index"
	maxIndexSize    = 1000 // Batasi jumlah key di index
)

type Users interface {
	List(ctx context.Context, search, order, sort string, limit, page int) ([]model.User, model.Pagination, *errors.BusinessError)
	Create(ctx context.Context, user *model.User) *errors.BusinessError
	FindByID(ctx context.Context, id string) (*model.User, *errors.BusinessError)
	Update(ctx context.Context, user *model.User) *errors.BusinessError
	Delete(ctx context.Context, id string) *errors.BusinessError
}

type users struct {
	db    *sql.DB
	cache cache.CacheClient
	log   logger.Logger
	ttl   config.TTLConfig
	repo  repository.UserRepository
}

func NewUsers(db *sql.DB, cache cache.CacheClient, log logger.Logger, ttl config.TTLConfig, repo repository.UserRepository) Users {
	return &users{db: db, cache: cache, log: log, ttl: ttl, repo: repo}
}

func (u *users) List(ctx context.Context, search, order, sort string, limit, page int) ([]model.User, model.Pagination, *errors.BusinessError) {
	cacheKey := listcache.GenerateListCacheKey(usersListPrefix, order, sort, search, limit, page)

	var cachedResult struct {
		Users      []model.User     `json:"users"`
		Pagination model.Pagination `json:"pagination"`
	}

	err := u.cache.GetJSON(ctx, cacheKey, &cachedResult)
	if err == nil {
		return cachedResult.Users, cachedResult.Pagination, nil
	}

	pagination := model.Pagination{Page: page, Limit: limit}
	offset := (pagination.Page - 1) * pagination.Limit

	users, count, err := u.repo.List(ctx, search, order, sort, pagination.Limit, offset)
	if err != nil {
		return nil, pagination, errors.InternalServerErrorWrap(err, "error listing users")
	}
	pagination.Count = count

	if err := u.saveListCache(ctx, cacheKey, users, pagination); err != nil {
		u.log.Warn(ctx, "failed to save cache", slog.Any("error", err))
	}

	return users, pagination, nil
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

	if err := listcache.InvalidateListCache(ctx, u.log, u.cache, usersIndexKey); err != nil {
		return errors.InternalServerErrorWrap(err, "error invalidate list cache")
	}

	if err = tx.Commit(); err != nil {
		return errors.InternalServerErrorWrap(err)
	}

	return nil
}

func (u *users) FindByID(ctx context.Context, id string) (*model.User, *errors.BusinessError) {
	cacheKey := "users::" + id
	var user *model.User
	if err := u.cache.GetJSON(ctx, cacheKey, &user); err == nil {
		return user, nil
	}

	user, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.InternalServerErrorWrap(err, "error finding user")
	}
	if user == nil {
		return nil, errors.NotFound("user not found")
	}

	if err := u.cache.SetJSONWithExpiry(ctx, cacheKey, user, u.ttl.TTLDefault); err != nil {
		u.log.Warn(ctx, "set cache failed ", slog.Any("error", err))
	}

	return user, nil
}

func (u *users) Update(ctx context.Context, user *model.User) *errors.BusinessError {
	cacheKey := "users::" + user.ID

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

	if _, err := u.cache.Del(ctx, []string{cacheKey}); err != nil {
		u.log.Error(ctx, "del cache failed", slog.Any("error", err))
		return errors.InternalServerErrorWrap(err)
	}

	if err = tx.Commit(); err != nil {
		return errors.InternalServerErrorWrap(err)
	}

	return nil
}

func (u *users) Delete(ctx context.Context, id string) *errors.BusinessError {
	cacheKey := "users::" + id
	existUser, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error finding user")
	}
	if existUser == nil {
		return errors.NotFound("user not found")
	}

	if _, err := u.cache.Del(ctx, []string{cacheKey}); err != nil {
		u.log.Error(ctx, "del cache failed", slog.Any("error", err))
		return errors.InternalServerErrorWrap(err)
	}

	if err := listcache.InvalidateListCache(ctx, u.log, u.cache, usersIndexKey); err != nil {
		u.log.Error(ctx, "del list cache failed", slog.Any("error", err))
		return errors.InternalServerErrorWrap(err)
	}

	err = u.repo.Delete(ctx, id)
	if err != nil {
		return errors.InternalServerErrorWrap(err, "error deleting user")
	}
	return nil
}

func (u *users) saveListCache(ctx context.Context, cacheKey string, users []model.User, pagination model.Pagination) error {
	cacheData := struct {
		Users      []model.User     `json:"users"`
		Pagination model.Pagination `json:"pagination"`
	}{
		Users:      users,
		Pagination: pagination,
	}

	if err := listcache.AddKeyToIndex(ctx, u.log, u.cache, cacheKey, usersIndexKey, maxIndexSize); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	if err := u.cache.SetJSONWithExpiry(ctx, cacheKey, cacheData, u.ttl.TTLDefault); err != nil {
		u.log.Warn(ctx, "data cache failed, index will be cleaned up",
			slog.String("key", cacheKey),
			slog.Any("error", err))

		u.cache.SRem(ctx, usersIndexKey, cacheKey)
		return fmt.Errorf("failed to set data: %w", err)
	}

	return nil
}
