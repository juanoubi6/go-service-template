// @title Swagger go-service-template API
// @version 1.0
// @description Sample service that creates "locations"
// @tag.name go-service-template
// @tag.description API endpoints
package main

import (
	"go-service-template/config"
	_ "go-service-template/docs"
	customHTTP "go-service-template/http"
	"go-service-template/http/controllers"
	httpMiddleware "go-service-template/http/middleware"
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

	// Create support structures
	customHTTPClient := customHTTP.CreateCustomHTTPClient(appCfg.HTTPClientConfig)

	// Create repositories
	dalFactory := db.NewDALFactory(appCfg.DBConfig)
	googleMapsAPI := googleMapsRepo.NewGoogleMapsRepository(customHTTPClient)

	// Create services
	locationService := services.NewLocationService(dalFactory, googleMapsAPI)

	// Create controllers
	healthDBController := controllers.NewHealthController()
	metricsController := controllers.NewMetricsController()
	swaggerController := controllers.NewSwaggerController()
	locationsController := controllers.NewLocationController(locationService)

	customHTTP.CreateWebServer(
		appCfg.AppConfig,
		[]customHTTP.Middleware{
			httpMiddleware.CreateCorsMiddleware(config.GetCorsOriginAddressByEnv(env)).Handler,
			httpMiddleware.CreateAppContextMiddleware(),
			httpMiddleware.CreateLoggingMiddleware(),
		},
		[]customHTTP.Endpoint{
			swaggerController.SwaggerEndpoint(),
			healthDBController.HealthEndpoint(),
			metricsController.MetricsEndpoint(),
			locationsController.CreateLocationEndpoint(),
			locationsController.UpdateLocationEndpoint(),
			locationsController.PaginatedLocationsEndpoint(),
			locationsController.LocationDetailsEndpoint(),
		},
	)
}
