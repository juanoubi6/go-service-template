package http

import (
	"github.com/labstack/echo/v4"
)

type Middleware = echo.MiddlewareFunc
type Handler = echo.HandlerFunc

type Endpoint struct {
	Method      string
	Path        string
	Handler     Handler
	Middlewares []Middleware
}
