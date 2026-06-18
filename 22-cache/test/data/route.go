package data_test

import (
	"fmt"
	"workshop/pkg/app"
)

func Api() {
	routes := []app.RouteDefinition{
		{Method: "GET", Path: "/accesses", Group: "accesses", Alias: "accesses::list", HandlerFunc: nil},

		{Method: "GET", Path: "/roles", Group: "roles", Alias: "roles::list", HandlerFunc: nil},
		{Method: "POST", Path: "/roles", Group: "roles", Alias: "roles::create", HandlerFunc: nil},
		{Method: "GET", Path: "/roles/{id}", Group: "roles", Alias: "roles::view", HandlerFunc: nil},
		{Method: "PUT", Path: "/roles/{id}", Group: "roles", Alias: "roles::update", HandlerFunc: nil},
		{Method: "DELETE", Path: "/roles/{id}", Group: "roles", Alias: "roles::delete", HandlerFunc: nil},
		{Method: "POST", Path: "/roles/{id}/access/{access_id}", Group: "roles", Alias: "roles::grant", HandlerFunc: nil},
		{Method: "DELETE", Path: "/roles/{id}/access/{access_id}", Group: "roles", Alias: "roles::revoke", HandlerFunc: nil},

		{Method: "GET", Path: "/users", Group: "users", Alias: "users::list", HandlerFunc: nil},
		{Method: "POST", Path: "/users", Group: "users", Alias: "users::create", HandlerFunc: nil},
		{Method: "GET", Path: "/users/{id}", Group: "users", Alias: "users::view", HandlerFunc: nil},
		{Method: "PUT", Path: "/users/{id}", Group: "users", Alias: "users::update", HandlerFunc: nil},
		{Method: "DELETE", Path: "/users/{id}", Group: "users", Alias: "users::delete", HandlerFunc: nil},
	}

	fmt.Println(routes)
}
