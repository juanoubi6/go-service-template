package http

import "net/http"

type Middleware = func(http.Handler) http.Handler

type Endpoint struct {
	Method      string
	Path        string
	Handler     http.HandlerFunc
	Middlewares []Middleware
}
