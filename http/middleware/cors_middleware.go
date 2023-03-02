package middleware

import (
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	customHTTP "go-service-template/http"
)

func CreateCorsMiddleware(allowedOrigins []string) customHTTP.Middleware {
	if allowedOrigins == nil {
		allowedOrigins = []string{"*"}
	}

	if len(allowedOrigins) == 0 {
		allowedOrigins = append(allowedOrigins, "*")
	}

	return echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
}
