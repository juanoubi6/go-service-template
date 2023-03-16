package http

import (
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"go-service-template/config"
	"net/http"
	"time"
)

func CreateWebServer(
	appConfig config.AppConfig,
	webServerConfig config.WebServerConfig,
	globalMiddleware []Middleware,
	endpoints []Endpoint,
) *http.Server {
	router := echo.New()

	// Decorate router with Prometheus metrics
	promMetrics := prometheus.NewPrometheus(appConfig.Name, nil)
	promMetrics.Use(router)

	// Register global middleware
	router.Use(globalMiddleware...)

	// Create each endpoint with their custom middlewares
	for _, endpoint := range endpoints {
		router.Add(endpoint.Method, endpoint.Path, endpoint.Handler, endpoint.Middlewares...)
	}

	readHeaderTimeout, err := time.ParseDuration(webServerConfig.ReadHeaderTimeout)
	if err != nil {
		panic(err)
	}

	server := &http.Server{
		Addr:              webServerConfig.Address,
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	return server
}
