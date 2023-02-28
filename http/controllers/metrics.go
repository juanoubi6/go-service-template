package controllers

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	customHTTP "go-service-template/http"
	"go-service-template/monitor"
	"net/http"
)

type MetricsController struct {
	logger monitor.AppLogger
}

func NewMetricsController() *MetricsController {
	return &MetricsController{
		logger: monitor.GetStdLogger("Metrics Controller"),
	}
}

func (c *MetricsController) MetricsEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodGet,
		Path:    "/metrics",
		Handler: promhttp.Handler().ServeHTTP,
	}
}
