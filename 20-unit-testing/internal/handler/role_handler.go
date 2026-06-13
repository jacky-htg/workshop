package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"workshop/internal/dto"
	"workshop/internal/model"
	"workshop/internal/service"
	"workshop/pkg/errors"
	"workshop/pkg/response"
	"workshop/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/jacky-htg/go-libs/logger"
)

type RoleHandler interface {
	List(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	FindByID(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Grant(w http.ResponseWriter, r *http.Request)
	Revoke(w http.ResponseWriter, r *http.Request)
}

type roleHandler struct {
	log      logger.Logger
	service  service.Roles
	validate *validator.Validate
}

func NewRoleHandler(log logger.Logger, validate *validator.Validate, service service.Roles) RoleHandler {
	return &roleHandler{log: log, validate: validate, service: service}
}

func (u *roleHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := u.service.List(ctx)
	if err != nil {
		u.log.Error(ctx, "error: listing roles", slog.Any("error", err))
		response.SetError(ctx, u.log, w, err, nil)
		return
	}

	var resp []dto.RoleResponse = make([]dto.RoleResponse, 0)
	for _, val := range list {
		var obj dto.RoleResponse
		obj.Transform(val)
		resp = append(resp, obj)
	}

	response.SetOk(ctx, u.log, w, resp)
}

func (u *roleHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dto.RoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		u.log.Error(ctx, "error: decoding role request", slog.Any("error", err))
		response.SetError(ctx, u.log, w, errors.InvalidInputWrap(err), nil)
		return
	}

	if err := u.validate.Struct(req); err != nil {
		u.log.Error(ctx, "error: decoding role request", slog.Any("error", err))
		response.SetError(ctx, u.log, w, errors.InvalidInputWrap(err), validation.FormatValidationErrors(err))
		return
	}

	obj := model.Role{}
	req.Transform(&obj)
	err := u.service.Create(ctx, &obj)
	if err != nil {
		response.SetError(ctx, u.log, w, err, nil)
		return
	}

	var resp dto.RoleResponse
	resp.Transform(obj)
	response.SetCreated(ctx, u.log, w, resp)
}

func (u *roleHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Missing id parameter"), nil)
		return
	}

	roleID, err := strconv.Atoi(id)
	if err != nil {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Invalid id"), nil)
		return
	}

	role, bizErr := u.service.FindByID(ctx, roleID)
	if bizErr != nil {
		response.SetError(ctx, u.log, w, bizErr, nil)
		return
	}

	var resp dto.RoleResponse
	resp.Transform(*role)

	response.SetOk(ctx, u.log, w, resp)
}

func (u *roleHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Missing id parameter"), nil)
		return
	}

	roleID, err := strconv.Atoi(id)
	if err != nil {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Invalid id"), nil)
		return
	}

	var req dto.RoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		u.log.Error(ctx, "error: decoding user request", slog.Any("error", err))
		response.SetError(ctx, u.log, w, errors.InvalidInputWrap(err), nil)
		return
	}

	if err := u.validate.Struct(req); err != nil {
		u.log.Error(ctx, "error: decoding user request", slog.Any("error", err))
		response.SetError(ctx, u.log, w, errors.InvalidInputWrap(err), validation.FormatValidationErrors(err))
		return
	}

	obj := model.Role{ID: roleID}
	req.Transform(&obj)
	if err := u.service.Update(ctx, &obj); err != nil {
		response.SetError(ctx, u.log, w, err, nil)
		return
	}

	var resp dto.RoleResponse
	resp.Transform(obj)

	response.SetOk(ctx, u.log, w, resp)
}

func (u *roleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Missing id parameter"), nil)
		return
	}

	roleID, err := strconv.Atoi(id)
	if err != nil {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Invalid id"), nil)
		return
	}

	if err := u.service.Delete(ctx, roleID); err != nil {
		response.SetError(ctx, u.log, w, err, nil)
		return
	}
	response.SetOk(ctx, u.log, w, struct{}{})
}

func (u *roleHandler) Grant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Missing id parameter"), nil)
		return
	}

	idAccess := r.PathValue("access_id")
	if idAccess == "" {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Missing access_id parameter"), nil)
		return
	}

	roleID, err := strconv.Atoi(id)
	if err != nil {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Invalid id"), nil)
		return
	}

	accessID, err := strconv.Atoi(idAccess)
	if err != nil {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Invalid access_id"), nil)
		return
	}

	if err := u.service.Grant(ctx, roleID, accessID); err != nil {
		response.SetError(ctx, u.log, w, err, nil)
		return
	}
	response.SetOk(ctx, u.log, w, struct{}{})
}

func (u *roleHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Missing id parameter"), nil)
		return
	}

	idAccess := r.PathValue("access_id")
	if idAccess == "" {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Missing access_id parameter"), nil)
		return
	}

	roleID, err := strconv.Atoi(id)
	if err != nil {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Invalid id"), nil)
		return
	}

	accessID, err := strconv.Atoi(idAccess)
	if err != nil {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Invalid access_id"), nil)
		return
	}

	if err := u.service.Revoke(ctx, roleID, accessID); err != nil {
		response.SetError(ctx, u.log, w, err, nil)
		return
	}
	response.SetOk(ctx, u.log, w, struct{}{})
}
