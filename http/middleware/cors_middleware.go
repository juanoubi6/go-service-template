package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func CreateCorsMiddleware(allowedOrigins []string) echo.MiddlewareFunc {
	if allowedOrigins == nil {
		allowedOrigins = []string{"*"}
	}

	if len(allowedOrigins) == 0 {
		allowedOrigins = append(allowedOrigins, "*")
	}

	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
}
