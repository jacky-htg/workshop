package service

import (
	"context"
	"workshop/internal/model"
	"workshop/internal/repository"
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
	log  logger.Logger
	repo repository.RoleRepository
}

func NewRoles(log logger.Logger, repo repository.RoleRepository) Roles {
	return &roles{log: log, repo: repo}
}

func (u *roles) List(ctx context.Context) ([]model.Role, *errors.BusinessError) {
	list, err := u.repo.List(ctx)
	if err != nil {
		return nil, errors.InternalServerErrorWrap(err, "error listing roles")
	}
	return list, nil
}

func (u *roles) FindByID(ctx context.Context, id int) (*model.Role, *errors.BusinessError) {
	obj, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.InternalServerErrorWrap(err, "error finding role")
	}
	if obj == nil {
		return nil, errors.NotFound("role not found")
	}
	return obj, nil
}

func (u *roles) Create(ctx context.Context, role *model.Role) *errors.BusinessError {
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
		err = u.repo.RevokeAccess(ctx, roleID, accessID)
		if err != nil {
			return errors.InternalServerErrorWrap(err, "error revoke access")
		}
	}
	return nil
}
