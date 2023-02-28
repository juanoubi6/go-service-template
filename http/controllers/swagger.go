package controllers

import (
	httpSwagger "github.com/swaggo/http-swagger"
	customHTTP "go-service-template/http"
	"net/http"
)

type SwaggerController struct{}

func NewSwaggerController() *SwaggerController {
	return &SwaggerController{}
}

func (c *SwaggerController) SwaggerEndpoint() customHTTP.Endpoint {
	return customHTTP.Endpoint{
		Method:  http.MethodGet,
		Path:    "/v1/swagger/*",
		Handler: httpSwagger.Handler(),
	}
}
