// @title Swagger go-service-template API
// @version 1.0
// @description Sample service that creates "locations"
// @tag.name go-service-template
// @tag.description API endpoints
package main

import (
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
	dalFactory := db.NewDALFactory(appCfg.DBConfig)
	googleMapsAPI := googleMapsRepo.NewGoogleMapsRepository(customHTTPClient)

	// Create services
	locationService := services.NewLocationService(dalFactory, googleMapsAPI)

	// Create controllers
	healthDBController := controllers.NewHealthController()
	swaggerController := controllers.NewSwaggerController()
	locationsController := controllers.NewLocationController(locationService, structValidator)

	customHTTP.CreateWebServer(
		appCfg.AppConfig,
		[]echo.MiddlewareFunc{ // Middlewares are run in the slice order
			echoMiddleware.Recover(),
			httpMiddleware.CreateCorsMiddleware(config.GetCorsOriginAddressByEnv(env)),
			httpMiddleware.CreateAppContextMiddleware(),
			echoMiddleware.Logger(),
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
}
