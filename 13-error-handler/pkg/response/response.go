package response

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"workshop/pkg/errors"

	"github.com/jacky-htg/go-libs/logger"
)

const AppBusinessStatusSuccess = "B1"

type StandardResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func SetResponse(log logger.Logger, w http.ResponseWriter, httpStatus int, appBusinessLogicStatus string, message string, data any) {
	standardResponse := StandardResponse{
		Status:  appBusinessLogicStatus,
		Message: message,
		Data:    data,
	}

	resp, err := json.Marshal(standardResponse)
	if err != nil {
		log.Error(context.Background(), "error: marshaling users to JSON", slog.Any("error", err))
		httpStatus = http.StatusInternalServerError
		appBusinessLogicStatus = errors.InternalServerErrorCode
		message = "Internal Server Error"
	}

	w.Header().Set("Content-Type", "application/json")
	if httpStatus != http.StatusOK {
		w.WriteHeader(httpStatus)
	}

	if _, err = w.Write(resp); err != nil {
		log.Error(context.Background(), "error: writing response", slog.Any("error", err))
	}
}

func SetError(log logger.Logger, w http.ResponseWriter, err *errors.BusinessError, message ...string) {
	finalMessage := ""
	if len(message) > 0 && len(message[0]) > 0 {
		finalMessage = message[0]
	}

	if finalMessage == "" && err != nil {
		finalMessage = err.Message
	}
	SetResponse(log, w, err.HTTPStatus, err.Code, finalMessage, struct{}{})
}

func SetOk(log logger.Logger, w http.ResponseWriter, data any) {
	SetResponse(log, w, http.StatusOK, AppBusinessStatusSuccess, "Success", data)
}

func SetCreated(log logger.Logger, w http.ResponseWriter, data any) {
	SetResponse(log, w, http.StatusCreated, AppBusinessStatusSuccess, "Created", data)
}
