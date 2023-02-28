package http

import (
	"github.com/go-chi/chi/v5"
	"go-service-template/config"
	"go-service-template/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func CreateWebServer(
	appConfig config.AppConfig,
	globalMiddleware []Middleware,
	endpoints []Endpoint,
) {
	router := chi.NewRouter()

	// Register global middleware
	router.Use(globalMiddleware...)

	// Create each endpoint with their custom middlewares
	for _, endpoint := range endpoints {
		router.With(endpoint.Middlewares...).MethodFunc(endpoint.Method, endpoint.Path, endpoint.Handler)
	}

	if err := http.ListenAndServe(appConfig.BindAddress, router); err != nil {
		panic(err)
	}

	gracefulShutdown()
}

func gracefulShutdown() {
	shutdownLog := log.GetStdLogger("gracefulShutdown")

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)

	<-c
	shutdownLog.Warn("gracefulShutdown", "", "Shutting down app, notifying workers to exit")

	return
}
