package middleware

import (
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	customHTTP "go-service-template/http"
	"net/http"
)

const CorsMaxAge = 300

func CreateCorsMiddleware(allowedOrigins []string) customHTTP.Middleware {
	if allowedOrigins == nil {
		allowedOrigins = []string{"*"}
	}

	if len(allowedOrigins) == 0 {
		allowedOrigins = append(allowedOrigins, "*")
	}

	return echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodHead},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           CorsMaxAge, // Maximum value not ignored by any of major browsers
	})
}
