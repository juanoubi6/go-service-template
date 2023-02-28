package http

import (
	"context"
	"github.com/go-chi/chi/v5"
	"go-service-template/config"
	"go-service-template/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	serverCtx, serverCtxCancelFn := context.WithCancel(context.Background())
	server := &http.Server{Addr: appConfig.BindAddress, Handler: router}

	go handleGracefulShutdown(serverCtx, serverCtxCancelFn, server)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}

func handleGracefulShutdown(serverCtx context.Context, serverCancelFn context.CancelFunc, server *http.Server) {
	fnName := "handleGracefulShutdown"
	shutdownLog := log.GetStdLogger("gracefulShutdown")

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)

	<-c
	shutdownLog.Warn(fnName, "", "Shutting down app")

	// Shutdown signal with grace period of 30 seconds
	shutdownCtx, shutdownCancelFn := context.WithTimeout(serverCtx, 30*time.Second)
	defer shutdownCancelFn()

	// Flush any buffered logs and traces
	log.FlushLogger()
	log.FlushTracerProvider(shutdownCtx)

	// Trigger graceful shutdown
	err := server.Shutdown(shutdownCtx)
	if err != nil {
		shutdownLog.Error(fnName, "", "failed to shutdown server", err)
	}

	serverCancelFn()
}
