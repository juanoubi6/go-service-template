package controllers

import (
	"github.com/labstack/echo/v4"
	customHTTP "go-service-template/http"
	"go-service-template/monitor"
	"net/http"
)

type HealthController struct {
	logger monitor.AppLogger
}

func NewHealthController() *HealthController {
	return &HealthController{
		logger: monitor.GetStdLogger("Health Controller"),
	}
}

// Nada godoc
// @Summary Check health
// @Description Simple healthcheck endpoint
// @Produce plain
// @Success 200
// @Router /health [get]
func (hc *HealthController) HealthEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodGet,
		Path:    "/health",
		Handler: hc.health,
	}
}

func (hc *HealthController) health(c echo.Context) error {
	return c.String(http.StatusOK, "Service is healthy")
}
