package controllers

import (
	customHTTP "go-service-template/http"
	"go-service-template/http/middleware"
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

func (hc *HealthController) health(writer http.ResponseWriter, r *http.Request) {
	appCtx := middleware.GetAppContext(r)

	err := sendSuccessResponse(writer, "Service is healthy", http.StatusOK)
	if err != nil {
		hc.logger.Error("Health", appCtx.GetCorrelationID(), "unable to send response", err)
		return
	}
}
