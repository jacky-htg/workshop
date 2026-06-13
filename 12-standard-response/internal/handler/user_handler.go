package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"workshop/internal/dto"
	"workshop/internal/model"
	"workshop/internal/service"
	"workshop/pkg/response"

	"github.com/jacky-htg/go-libs/logger"
)

type UserHandler interface {
	List(w http.ResponseWriter, r *http.Request)
	Create(w http.ResponseWriter, r *http.Request)
	FindById(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type userHandler struct {
	log     logger.Logger
	service service.Users
}

func NewUserHandler(service service.Users, log logger.Logger) UserHandler {
	return &userHandler{service: service, log: log}
}

// List : http handler for returning list of users
func (u *userHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := u.service.List()
	if err != nil {
		u.log.Error(context.Background(), "error: listing users", slog.Any("error", err))
		response.SetError(u.log, w, http.StatusInternalServerError, response.AppBusinessStatusError, err, "Failed to list users")
		return
	}

	var resp []dto.UserResponse
	for _, user := range users {
		var ur dto.UserResponse
		ur.Transform(user)
		resp = append(resp, ur)
	}

	response.SetOk(u.log, w, resp)
}

// Create : http handler for creating a new user
func (u *userHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		u.log.Error(context.Background(), "error: decoding user request", slog.Any("error", err))
		response.SetError(u.log, w, http.StatusBadRequest, response.AppBusinessStatusError, err, "Invalid request payload")
		return
	}
	user := model.User{}
	req.Transform(&user)
	err := u.service.Create(&user)
	if err != nil {
		response.SetError(u.log, w, http.StatusInternalServerError, response.AppBusinessStatusError, err, "Failed to create user")
		return
	}

	var resp dto.UserResponse
	resp.Transform(user)
	response.SetCreated(u.log, w, resp)
}

// FindById : http handler for finding a user by ID
func (u *userHandler) FindById(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.SetError(u.log, w, http.StatusBadRequest, response.AppBusinessStatusError, nil, "Missing id parameter")
		return
	}

	user, err := u.service.FindById(id)
	if err != nil {
		response.SetError(u.log, w, http.StatusInternalServerError, response.AppBusinessStatusError, err, "Failed to find user")
		return
	}
	if user == nil {
		response.SetError(u.log, w, http.StatusNotFound, response.AppBusinessStatusError, nil, "User not found")
		return
	}

	var resp dto.UserResponse
	resp.Transform(*user)

	response.SetOk(u.log, w, resp)
}

func (u *userHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.SetError(u.log, w, http.StatusBadRequest, response.AppBusinessStatusError, nil, "Missing id parameter")
		return
	}

	var req dto.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		u.log.Error(context.Background(), "error: decoding user request", slog.Any("error", err))
		response.SetError(u.log, w, http.StatusBadRequest, response.AppBusinessStatusError, nil, "Invalid request payload")
		return
	}
	user := model.User{ID: id}
	req.Transform(&user)
	err := u.service.Update(&user)
	if err != nil {
		if err.Error() == "user not found" {
			response.SetError(u.log, w, http.StatusNotFound, response.AppBusinessStatusError, nil, "User not found")
		} else {
			response.SetError(u.log, w, http.StatusInternalServerError, response.AppBusinessStatusError, err, "Failed to update user")
		}
		return
	}

	var resp dto.UserResponse
	resp.Transform(user)

	response.SetOk(u.log, w, resp)
}

func (u *userHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		response.SetError(u.log, w, http.StatusBadRequest, response.AppBusinessStatusError, nil, "Missing id parameter")
		return
	}

	err := u.service.Delete(id)
	if err != nil {
		if err.Error() == "user not found" {
			response.SetError(u.log, w, http.StatusNotFound, response.AppBusinessStatusError, nil, "User not found")
		} else {
			response.SetError(u.log, w, http.StatusInternalServerError, response.AppBusinessStatusError, err, "Failed to delete user")
		}
		return
	}
	response.SetOk(u.log, w, struct{}{})
}
