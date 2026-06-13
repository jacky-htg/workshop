package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"workshop/internal/dto"
	"workshop/internal/service"
	"workshop/pkg/errors"
	"workshop/pkg/response"
	"workshop/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/jacky-htg/go-libs/logger"
)

type AuthHandler interface {
	Login(w http.ResponseWriter, r *http.Request)
}

type authHandler struct {
	log      logger.Logger
	service  service.Auths
	validate *validator.Validate
}

func NewAuthHandler(log logger.Logger, validate *validator.Validate, service service.Auths) AuthHandler {
	return &authHandler{log: log, validate: validate, service: service}
}

func (u *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		u.log.Error(ctx, "error: decoding login request", slog.Any("error", err))
		response.SetError(ctx, u.log, w, errors.InvalidInputWrap(err), nil)
		return
	}

	if err := u.validate.Struct(req); err != nil {
		u.log.Error(ctx, "error: decoding login request", slog.Any("error", err))
		response.SetError(ctx, u.log, w, errors.InvalidInputWrap(err), validation.FormatValidationErrors(err))
		return
	}

	token, user, accesses, err := u.service.Login(ctx, req.Username, req.Password)
	if err != nil {
		response.SetError(ctx, u.log, w, err, nil)
		return
	}

	resp := dto.LoginResponse{Token: token}
	resp.Transform(token, *user, accesses)
	response.SetOk(ctx, u.log, w, resp)
}
