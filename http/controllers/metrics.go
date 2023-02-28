package controllers

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	customHTTP "go-service-template/http"
	"go-service-template/log"
	"net/http"
)

type MetricsController struct {
	logger log.AppLogger
}

func NewMetricsController() *MetricsController {
	return &MetricsController{
		logger: log.GetStdLogger("Metrics Controller"),
	}
}

func (c *MetricsController) MetricsEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodGet,
		Path:    "/metrics",
		Handler: promhttp.Handler().ServeHTTP,
	}
}
