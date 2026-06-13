package middleware

import (
	"context"
	"net/http"
	"time"
	"workshop/pkg/errors"
	"workshop/pkg/response"

	"github.com/jacky-htg/go-libs/logger"
	lib "github.com/jacky-htg/go-libs/middleware"
)

func Timeout(log logger.Logger, timeout time.Duration) lib.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			originalCtx := r.Context()
			ctx, cancel := context.WithTimeout(originalCtx, timeout)
			defer cancel()

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r.WithContext(ctx))
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				response.SetError(originalCtx, log, w, errors.GatewayTimeout(), nil)
			}
		})
	}
}
