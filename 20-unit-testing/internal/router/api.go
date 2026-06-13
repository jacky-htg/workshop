package router

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"workshop/config"
	"workshop/internal/handler"
	"workshop/internal/repository"
	"workshop/internal/service"
	"workshop/pkg/app"
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

	accessRepository := repository.NewAccessRepository(db, log)
	roleRepository := repository.NewRoleRepository(db, log)
	userRepository := repository.NewUserRepository(db, log)

	accessService := service.NewAccesses(db, log, accessRepository)
	authService := service.NewAuths(log, cfg.Token, userRepository, roleRepository)
	roleService := service.NewRoles(log, roleRepository)
	userService := service.NewUsers(db, log, userRepository)

	accessHandler := handler.NewAccessHandler(log, validate, accessService)
	authHandler := handler.NewAuthHandler(log, validate, authService)
	roleHandler := handler.NewRoleHandler(log, validate, roleService)
	userHandler := handler.NewUserHandler(log, validate, userService)

	mux.Handle("GET /health", base.Then(func(w http.ResponseWriter, r *http.Request) {
		response.SetOk(r.Context(), log, w, struct{}{})
	}))

	mux.Handle("POST /login", base.Then(authHandler.Login))

	privateRoutes := []app.RouteDefinition{
		{Method: "GET", Path: "/accesses", Group: "accesses", Alias: "accesses::list", HandlerFunc: accessHandler.List},

		{Method: "GET", Path: "/roles", Group: "roles", Alias: "roles::list", HandlerFunc: roleHandler.List},
		{Method: "POST", Path: "/roles", Group: "roles", Alias: "roles::create", HandlerFunc: roleHandler.Create},
		{Method: "GET", Path: "/roles/{id}", Group: "roles", Alias: "roles::view", HandlerFunc: roleHandler.FindByID},
		{Method: "PUT", Path: "/roles/{id}", Group: "roles", Alias: "roles::update", HandlerFunc: roleHandler.Update},
		{Method: "DELETE", Path: "/roles/{id}", Group: "roles", Alias: "roles::delete", HandlerFunc: roleHandler.Delete},
		{Method: "POST", Path: "/roles/{id}/access/{access_id}", Group: "roles", Alias: "roles::grant", HandlerFunc: roleHandler.Grant},
		{Method: "DELETE", Path: "/roles/{id}/access/{access_id}", Group: "roles", Alias: "roles::revoke", HandlerFunc: roleHandler.Revoke},

		{Method: "GET", Path: "/users", Group: "users", Alias: "users::list", HandlerFunc: userHandler.List},
		{Method: "POST", Path: "/users", Group: "users", Alias: "users::create", HandlerFunc: userHandler.Create},
		{Method: "GET", Path: "/users/{id}", Group: "users", Alias: "users::view", HandlerFunc: userHandler.FindByID},
		{Method: "PUT", Path: "/users/{id}", Group: "users", Alias: "users::update", HandlerFunc: userHandler.Update},
		{Method: "DELETE", Path: "/users/{id}", Group: "users", Alias: "users::delete", HandlerFunc: userHandler.Delete},
	}

	for _, route := range privateRoutes {
		pattern := fmt.Sprintf("%s %s", route.Method, route.Path)
		wrappedHandler := wrapWithRoutePattern(pattern, route.Group, private.Then(route.HandlerFunc))
		mux.Handle(pattern, wrappedHandler)
	}

	return mux
}

func wrapWithRoutePattern(pattern, group string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), app.MyCtx("route-path"), pattern)
		ctx = context.WithValue(ctx, app.MyCtx("route-group"), group)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
