package middleware

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"strings"
	"workshop/internal/repository"
	"workshop/pkg/app"
	"workshop/pkg/errors"
	"workshop/pkg/response"

	"github.com/jacky-htg/go-libs/logger"
	lib "github.com/jacky-htg/go-libs/middleware"
	"github.com/jacky-htg/go-libs/token"
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

			mytoken := strings.TrimPrefix(authHeader, "Bearer ")
			if mytoken == authHeader {
				err := errors.Unauthorized("Invalid authorization header")
				log.Error(ctx, "Unauthorized", slog.Any("error", err))
				response.SetError(ctx, log, w, err, nil)
				return
			}

			isValid, claim := token.ValidateToken(mytoken)
			if !isValid {
				err := errors.Unauthorized("Invalid token")
				log.Error(ctx, "Unauthorized", slog.Any("error", err))
				response.SetError(ctx, log, w, err, nil)
				return
			}

			email := token.GetString(claim, "email")
			repo := repository.NewUserRepository(db, log)
			hasPermission := repo.HasPermission(
				ctx,
				email,
				ctx.Value(app.MyCtx("route-path")).(string),
				ctx.Value(app.MyCtx("route-group")).(string),
			)
			if !hasPermission {
				response.SetError(ctx, log, w, errors.Forbidden(), nil)
				return
			}

			ctx = context.WithValue(ctx, app.MyCtx("email"), email)
			ctx = context.WithValue(ctx, app.MyCtx("userID"), token.GetString(claim, "id"))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
