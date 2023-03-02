package http

import (
	"context"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"go-service-template/config"
	"go-service-template/monitor"
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
	shutdownLog := monitor.GetStdLogger("gracefulShutdown")
	appCtx := monitor.CreateAppContextFromContext(serverCtx, fnName, "")

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)

	<-c
	shutdownLog.Warn(appCtx, fnName, "Shutting down app")

	// Shutdown signal with grace period of 30 seconds
	shutdownCtx, shutdownCancelFn := context.WithTimeout(serverCtx, 30*time.Second)
	defer shutdownCancelFn()

	// Flush any buffered logs and traces
	monitor.FlushLogger()
	monitor.FlushTracerProvider(shutdownCtx)

	// Trigger graceful shutdown
	err := server.Shutdown(shutdownCtx)
	if err != nil {
		shutdownLog.Error(appCtx, fnName, "failed to shutdown server", err)
	}

	serverCancelFn()
}
