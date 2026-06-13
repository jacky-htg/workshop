package response

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jacky-htg/go-libs/logger"
)

const (
	AppBusinessStatusSuccess = "B1"
	AppBusinessStatusError   = "B0"
)

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
		appBusinessLogicStatus = AppBusinessStatusError
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

func SetError(log logger.Logger, w http.ResponseWriter, httpStatus int, appBusinessLogicStatus string, err error, message string) {
	finalMessage := message
	if finalMessage == "" && err != nil {
		finalMessage = err.Error()
	}
	SetResponse(log, w, httpStatus, appBusinessLogicStatus, finalMessage, struct{}{})
}

func SetOk(log logger.Logger, w http.ResponseWriter, data any) {
	SetResponse(log, w, http.StatusOK, AppBusinessStatusSuccess, "Success", data)
}

func SetCreated(log logger.Logger, w http.ResponseWriter, data any) {
	SetResponse(log, w, http.StatusCreated, AppBusinessStatusSuccess, "Created", data)
}
