package middleware

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strings"
	"workshop/pkg/errors"
	"workshop/pkg/response"

	"github.com/jacky-htg/go-libs/logger"
	lib "github.com/jacky-htg/go-libs/middleware"
)

func Auth(db *sql.DB, log logger.Logger) lib.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				err := errors.Unauthorized()
				log.Error(ctx, "Unauthorized", slog.Any("error", err))
				response.SetError(ctx, log, w, err, nil)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				err := errors.Unauthorized("Invalid authorization header")
				log.Error(ctx, "Unauthorized", slog.Any("error", err))
				response.SetError(ctx, log, w, err, nil)
				return
			}

			if len(token) <= 10 {
				err := errors.Unauthorized("Invalid token")
				log.Error(ctx, "Unauthorized", slog.Any("error", err))
				response.SetError(ctx, log, w, err, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
