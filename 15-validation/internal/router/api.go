package router

import (
	"database/sql"
	"net/http"
	"workshop/config"
	"workshop/internal/handler"
	"workshop/internal/repository"
	"workshop/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/jacky-htg/go-libs/logger"
)

func Api(
	cfg config.Config,
	db *sql.DB,
	log logger.Logger,
	validate *validator.Validate,
) http.Handler {
	mux := http.NewServeMux()

	userRepository := repository.NewUserRepository(db, log)
	userService := service.NewUsers(log, userRepository)
	userHandler := handler.NewUserHandler(log, validate, userService)
	mux.HandleFunc("GET /users", userHandler.List)
	mux.HandleFunc("POST /users", userHandler.Create)
	mux.HandleFunc("GET /users/{id}", userHandler.FindById)
	mux.HandleFunc("PUT /users/{id}", userHandler.Update)
	mux.HandleFunc("DELETE /users/{id}", userHandler.Delete)

	return mux
}
