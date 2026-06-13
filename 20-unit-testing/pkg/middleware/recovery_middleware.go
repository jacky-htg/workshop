package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"workshop/pkg/errors"
	"workshop/pkg/response"

	"github.com/jacky-htg/go-libs/logger"
	lib "github.com/jacky-htg/go-libs/middleware"
)

func Recovery(log logger.Logger) lib.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					ctx := r.Context()
					log.Error(ctx, "panic recovered",
						slog.Any("error", err),
						slog.String("stack", string(debug.Stack())))
					response.SetError(ctx, log, w, errors.InternalServerError(), nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
