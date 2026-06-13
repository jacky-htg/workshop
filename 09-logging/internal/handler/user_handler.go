package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"workshop/internal/dto"
	"workshop/internal/service"

	"github.com/jacky-htg/go-libs/logger"
)

type UserHandler interface {
	List(w http.ResponseWriter, r *http.Request)
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
