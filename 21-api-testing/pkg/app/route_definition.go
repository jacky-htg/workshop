package app

import "net/http"

type RouteDefinition struct {
	Method      string
	Path        string
	Group       string
	Alias       string
	HandlerFunc http.HandlerFunc
}
