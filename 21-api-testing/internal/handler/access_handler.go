package handler

import (
	"log/slog"
	"net/http"

	"workshop/internal/dto"
	"workshop/internal/service"
	"workshop/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/jacky-htg/go-libs/logger"
)

type AccessHandler interface {
	List(w http.ResponseWriter, r *http.Request)
}

type accessHandler struct {
	log      logger.Logger
	service  service.Accesses
	validate *validator.Validate
}

func NewAccessHandler(log logger.Logger, validate *validator.Validate, service service.Accesses) AccessHandler {
	return &accessHandler{log: log, validate: validate, service: service}
}

func (u *accessHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := u.service.List(ctx)
	if err != nil {
		u.log.Error(ctx, "error: listing access", slog.Any("error", err))
		response.SetError(ctx, u.log, w, err, nil)
		return
	}

	var resp []dto.AccessTreeResponse = make([]dto.AccessTreeResponse, 0)
	for _, val := range list {
		var obj dto.AccessTreeResponse
		obj.Transform(*val)
		resp = append(resp, obj)
	}

	response.SetOk(ctx, u.log, w, resp)
}
