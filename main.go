// @title Swagger go-service-template API
// @version 1.0
// @description Sample service that creates "locations"
// @tag.name go-service-template
// @tag.description API endpoints
package main

import (
	"context"
	"errors"
	"go-service-template/config"
	_ "go-service-template/docs"
	"go-service-template/eventhandler"
	customHTTP "go-service-template/http"
	"go-service-template/http/controllers"
	httpMiddleware "go-service-template/http/middleware"
	"go-service-template/monitor"
	"go-service-template/pubsub"
	"go-service-template/repositories/db"
	googleMapsRepo "go-service-template/repositories/googlemaps"
	"go-service-template/services"
	"go-service-template/utils"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	watermillMiddleware "github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

const ShutdownTimeSec = 30

func main() {
	env, _ := config.GetEnvironment()
	appCfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// Open Telemetry tools
	monitor.RegisterMonitoringTools(appCfg.OpenTelemetryConfig, appCfg.AppConfig)

	// Create support structures
	customHTTPClient := customHTTP.CreateCustomHTTPClient(appCfg.HTTPClientConfig)
	structValidator := validator.New()
	subscriber, err := pubsub.CreateSubscriber(kafka.DefaultSaramaSubscriberConfig(), appCfg.KafkaConfig)
	if err != nil {
		panic(err)
	}
	publisher, err := pubsub.CreatePublisher(kafka.DefaultSaramaSyncPublisherConfig(), appCfg.KafkaConfig)
	if err != nil {
		panic(err)
	}

	// Create repositories
	dalFactory := db.NewFactory(appCfg.DBConfig)
	googleMapsAPI := googleMapsRepo.NewGoogleMapsRepository(customHTTPClient)

	// Create services
	locationService := services.NewLocationService(dalFactory, googleMapsAPI, publisher)

	// Create HTTP controllers
	healthDBController := controllers.NewHealthController()
	swaggerController := controllers.NewSwaggerController()
	locationsController := controllers.NewLocationController(locationService, structValidator)

	// Create event handlers
	newLocationHandler := eventhandler.CreateNewLocationHandler()
	updatedLocationHandler := eventhandler.CreateUpdatedLocationHandler()

	webServer := customHTTP.CreateWebServer(
		appCfg.AppConfig,
		appCfg.WebServerConfig,
		[]customHTTP.Middleware{ // Middlewares are run in the slice order
			otelecho.Middleware(appCfg.AppConfig.Name, otelecho.WithSkipper(func(c echo.Context) bool {
				ignoredPaths := []string{"/health", "/metrics"}
				return utils.ListContains(ignoredPaths, c.Path())
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
			locationsController.CreateLocationMockEndpoint(),
		},
	)

	eventRouter, err := pubsub.CreateRouter(
		[]message.HandlerMiddleware{watermillMiddleware.Recoverer},
		[]pubsub.EventHandler{newLocationHandler, updatedLocationHandler},
		subscriber,
	)
	if err != nil {
		panic(err)
	}

	serverCtx, serverCtxCancelFn := context.WithCancel(context.Background())

	// Prepare graceful shutdown handler
	go handleGracefulShutdown(serverCtx, serverCtxCancelFn, webServer, eventRouter)

	// Start event handler in new goroutine
	go func() {
		if routerErr := eventRouter.Run(serverCtx); routerErr != nil {
			panic(routerErr)
		}
	}()

	if err = webServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}

func handleGracefulShutdown(
	serverCtx context.Context,
	serverCancelFn context.CancelFunc,
	server *http.Server,
	router *message.Router,
) {
	fnName := "handleGracefulShutdown"
	shutdownLog := monitor.GetStdLogger("gracefulShutdown")

	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)

	<-c
	shutdownLog.Warn(fnName, "", "Shutting down application")

	// Shutdown signal with grace period of 30 seconds
	shutdownCtx, shutdownCancelFn := context.WithTimeout(serverCtx, ShutdownTimeSec*time.Second)
	defer shutdownCancelFn()

	// Flush any buffered logs, traces and metrics
	monitor.FlushLogger()
	monitor.FlushMonitorTools(shutdownCtx)

	// Close web server
	if err := server.Shutdown(shutdownCtx); err != nil {
		shutdownLog.Error(fnName, "", "failed to shutdown web server", err)
	}

	// Close event router
	if err := router.Close(); err != nil {
		shutdownLog.Error(fnName, "", "failed to shutdown event router", err)
	}

	serverCancelFn()
}
