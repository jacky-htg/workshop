package router

import (
	"database/sql"
	"net/http"
	"workshop/config"
	"workshop/internal/handler"
	"workshop/internal/repository"
	"workshop/internal/service"
	mid "workshop/pkg/middleware"
	"workshop/pkg/response"

	"github.com/go-playground/validator/v10"
	"github.com/jacky-htg/go-libs/logger"
	"github.com/jacky-htg/go-libs/middleware"
)

func Api(
	cfg config.Config,
	db *sql.DB,
	log logger.Logger,
	validate *validator.Validate,
) http.Handler {
	mux := http.NewServeMux()

	base := middleware.Stack{
		mid.Recovery(log),
		mid.Timeout(log, cfg.Server.GatewayTimeout),
	}
	private := base.With(mid.Auth(db, log))

	userRepository := repository.NewUserRepository(db, log)
	userService := service.NewUsers(log, userRepository)
	userHandler := handler.NewUserHandler(log, validate, userService)

	mux.Handle("GET /health", base.Then(func(w http.ResponseWriter, r *http.Request) {
		response.SetOk(r.Context(), log, w, struct{}{})
	}))

	mux.Handle("GET /users", private.Then(userHandler.List))
	mux.Handle("POST /users", private.Then(userHandler.Create))
	mux.Handle("GET /users/{id}", private.Then(userHandler.FindById))
	mux.Handle("PUT /users/{id}", private.Then(userHandler.Update))
	mux.Handle("DELETE /users/{id}", private.Then(userHandler.Delete))

	return mux
}
