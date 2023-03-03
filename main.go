// @title Swagger go-service-template API
// @version 1.0
// @description Sample service that creates "locations"
// @tag.name go-service-template
// @tag.description API endpoints
package main

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go-service-template/config"
	_ "go-service-template/docs"
	customHTTP "go-service-template/http"
	"go-service-template/http/controllers"
	httpMiddleware "go-service-template/http/middleware"
	"go-service-template/monitor"
	"go-service-template/repositories/db"
	googleMapsRepo "go-service-template/repositories/googlemaps"
	"go-service-template/services"
	"go-service-template/utils"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	env, _ := config.GetEnvironment()
	appCfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Open Telemetry
	monitor.RegisterTraceProvider(appCfg.OpenTelemetryConfig, appCfg.AppConfig)

	// Create support structures
	customHTTPClient := customHTTP.CreateCustomHTTPClient(appCfg.HTTPClientConfig)
	structValidator := validator.New()

	// Create repositories
	dalFactory := db.NewDBFactory(appCfg.DBConfig)
	googleMapsAPI := googleMapsRepo.NewGoogleMapsRepository(customHTTPClient)

	// Create services
	locationService := services.NewLocationService(dalFactory, googleMapsAPI)

	// Create controllers
	healthDBController := controllers.NewHealthController()
	swaggerController := controllers.NewSwaggerController()
	locationsController := controllers.NewLocationController(locationService, structValidator)

	webServer := customHTTP.CreateWebServer(
		appCfg.AppConfig,
		[]customHTTP.Middleware{ // Middlewares are run in the slice order
			otelecho.Middleware(appCfg.AppConfig.Name, otelecho.WithSkipper(func(c echo.Context) bool {
				ignoredPaths := []string{"/health", "/metrics"}
				return utils.ListContains[string](ignoredPaths, c.Path())
			})),
			echoMiddleware.Logger(),
			echoMiddleware.Recover(),
			httpMiddleware.CreateCorsMiddleware(config.GetCorsOriginAddressByEnv(env)),
			httpMiddleware.CreateAppContextMiddleware(),
		},
		[]customHTTP.Endpoint{
			swaggerController.SwaggerEndpoint(),
			healthDBController.HealthEndpoint(),
			locationsController.CreateLocationEndpoint(),
			locationsController.UpdateLocationEndpoint(),
			locationsController.PaginatedLocationsEndpoint(),
			locationsController.LocationDetailsEndpoint(),
		},
	)

	serverCtx, serverCtxCancelFn := context.WithCancel(context.Background())

	go handleGracefulShutdown(serverCtx, serverCtxCancelFn, webServer)

	if err = webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
