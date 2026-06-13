package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"workshop/internal/dto"
	"workshop/internal/model"
	"workshop/internal/service"

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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var response []dto.UserResponse
	for _, user := range users {
		var ur dto.UserResponse
		ur.Transform(user)
		response = append(response, ur)
	}

	data, err := json.Marshal(response)
	if err != nil {
		u.log.Error(context.Background(), "error: marshaling users to JSON", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(data); err != nil {
		u.log.Error(context.Background(), "error: writing response", slog.Any("error", err))
	}
}

// Create : http handler for creating a new user
func (u *userHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		u.log.Error(context.Background(), "error: decoding user request", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	user := model.User{}
	req.Transform(&user)
	err := u.service.Create(&user)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var response dto.UserResponse
	response.Transform(user)

	data, err := json.Marshal(response)
	if err != nil {
		u.log.Error(context.Background(), "error: marshaling user to JSON", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(data); err != nil {
		u.log.Error(context.Background(), "error: writing response", slog.Any("error", err))
	}
}

// FindById : http handler for finding a user by ID
func (u *userHandler) FindById(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Bad Request: missing id parameter", http.StatusBadRequest)
		return
	}

	user, err := u.service.FindById(id)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	var response dto.UserResponse
	response.Transform(*user)

	data, err := json.Marshal(response)
	if err != nil {
		u.log.Error(context.Background(), "error: marshaling user to JSON", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(data); err != nil {
		u.log.Error(context.Background(), "error: writing response", slog.Any("error", err))
	}
}

func (u *userHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Bad Request: missing id parameter", http.StatusBadRequest)
		return
	}

	var req dto.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		u.log.Error(context.Background(), "error: decoding user request", slog.Any("error", err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	user := model.User{ID: id}
	req.Transform(&user)
	err := u.service.Update(&user)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "Not Found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	var response dto.UserResponse
	response.Transform(user)

	data, err := json.Marshal(response)
	if err != nil {
		u.log.Error(context.Background(), "error: marshaling user to JSON", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(data); err != nil {
		u.log.Error(context.Background(), "error: writing response", slog.Any("error", err))
	}
}

func (u *userHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Bad Request: missing id parameter", http.StatusBadRequest)
		return
	}

	err := u.service.Delete(id)
	if err != nil {
		if err.Error() == "user not found" {
			http.Error(w, "Not Found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
