package controllers_test

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	customHTTP "go-service-template/http"
	"go-service-template/http/controllers"
	"net/http"
	"net/http/httptest"
	"testing"
)

type HealthControllerSuite struct {
	suite.Suite
	healthEndpoint customHTTP.Endpoint
	echoRouter     *echo.Echo
}

func (s *HealthControllerSuite) SetupSuite() {
	s.echoRouter = echo.New()
}

func (s *HealthControllerSuite) SetupTest() {
	healthController := controllers.NewHealthController()

	s.healthEndpoint = healthController.HealthEndpoint()
}

func TestHealthControllerSuite(t *testing.T) {
	suite.Run(t, new(HealthControllerSuite))
}

func (s *HealthControllerSuite) Test_Health_Success() {
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodGet, "/health", http.NoBody)

	err := s.healthEndpoint.Handler(s.echoRouter.NewContext(req, w))

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, w.Code)
}
