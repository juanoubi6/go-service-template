package http

import (
	"github.com/labstack/echo/v4"
)

type Endpoint struct {
	Method      string
	Path        string
	Handler     echo.HandlerFunc
	Middlewares []echo.MiddlewareFunc
}
