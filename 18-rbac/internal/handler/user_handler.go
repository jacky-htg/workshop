package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"workshop/internal/dto"
	"workshop/internal/model"
	"workshop/internal/service"
	"workshop/pkg/errors"
	"workshop/pkg/response"
	"workshop/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/jacky-htg/go-libs/logger"
)

type UserHandler interface {
	List(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	FindByID(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type userHandler struct {
	log      logger.Logger
	service  service.Users
	validate *validator.Validate
}

func NewUserHandler(log logger.Logger, validate *validator.Validate, service service.Users) UserHandler {
	return &userHandler{log: log, validate: validate, service: service}
}

// List : http handler for returning list of users
func (u *userHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := u.service.List(ctx)
	if err != nil {
		u.log.Error(ctx, "error: listing users", slog.Any("error", err))
		response.SetError(ctx, u.log, w, err, nil)
		return
	}

	var resp []dto.UserResponse = make([]dto.UserResponse, 0)
	for _, user := range users {
		var ur dto.UserResponse
		ur.Transform(user)
		resp = append(resp, ur)
	}

	response.SetOk(ctx, u.log, w, resp)
}

// Create : http handler for creating a new user
func (u *userHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dto.UserRequest
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

	user := model.User{}
	req.Transform(&user)
	err := u.service.Create(ctx, &user)
	if err != nil {
		response.SetError(ctx, u.log, w, err, nil)
		return
	}

	var resp dto.UserResponse
	resp.Transform(user)
	response.SetCreated(ctx, u.log, w, resp)
}

// FindById : http handler for finding a user by ID
func (u *userHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Missing id parameter"), nil)
		return
	}

	user, err := u.service.FindByID(ctx, id)
	if err != nil {
		response.SetError(ctx, u.log, w, err, nil)
		return
	}

	var resp dto.UserResponse
	resp.Transform(*user)

	response.SetOk(ctx, u.log, w, resp)
}

func (u *userHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Missing id parameter"), nil)
		return
	}

	var req dto.UserUpdateRequest
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

	user := model.User{ID: id}
	req.Transform(&user)
	err := u.service.Update(ctx, &user)
	if err != nil {
		response.SetError(ctx, u.log, w, err, nil)
		return
	}

	var resp dto.UserResponse
	resp.Transform(user)

	response.SetOk(ctx, u.log, w, resp)
}

func (u *userHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		response.SetError(ctx, u.log, w, errors.InvalidInput("Missing id parameter"), nil)
		return
	}

	err := u.service.Delete(ctx, id)
	if err != nil {
		response.SetError(ctx, u.log, w, err, nil)
		return
	}
	response.SetOk(ctx, u.log, w, struct{}{})
}
