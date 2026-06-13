package router

import (
	"database/sql"
	"net/http"
	"workshop/config"
	"workshop/internal/handler"
	"workshop/internal/repository"
	"workshop/internal/service"

	"github.com/jacky-htg/go-libs/logger"
)

func Api(
	cfg config.Config,
	db *sql.DB,
	log logger.Logger,
) http.Handler {
	mux := http.NewServeMux()

	userRepository := repository.NewUserRepository(db, log)
	userService := service.NewUsers(userRepository, log)
	userHandler := handler.NewUserHandler(userService, log)
	mux.HandleFunc("GET /users", userHandler.List)
	mux.HandleFunc("POST /users", userHandler.Create)
	mux.HandleFunc("GET /users/{id}", userHandler.FindById)
	mux.HandleFunc("PUT /users/{id}", userHandler.Update)
	mux.HandleFunc("DELETE /users/{id}", userHandler.Delete)

	return mux
}
