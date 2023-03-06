package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	customHTTP "go-service-template/http"
	"net/http"
	"net/http/httptest"
	"testing"
)

type AppContextMiddlewareSuite struct {
	suite.Suite
	appContextMiddleware customHTTP.Middleware
	testEndpoint         customHTTP.Handler
	echoRouter           *echo.Echo
	recorder             *httptest.ResponseRecorder
}

func (s *AppContextMiddlewareSuite) SetupSuite() {
	s.appContextMiddleware = CreateAppContextMiddleware()
	s.testEndpoint = CreateTestEndpoint()
	s.echoRouter = echo.New()
}

func (s *AppContextMiddlewareSuite) SetupTest() {
	s.recorder = httptest.NewRecorder()
}

func TestAppContextMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(AppContextMiddlewareSuite))
}

func (s *AppContextMiddlewareSuite) Test_AppContextMiddleware_DecoratesRequestAppContext() {
	req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)

	wrappedTestHandler := s.appContextMiddleware(s.testEndpoint)

	err := wrappedTestHandler(s.echoRouter.NewContext(req, s.recorder))

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, s.recorder.Code)
}

func (s *AppContextMiddlewareSuite) Test_AppContextMiddleware_UsesCorrelationIDHeaderAsCorrelationID() {
	req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Add(CorrelationIDHeader, "value")

	wrappedTestHandler := s.appContextMiddleware(s.testEndpoint)

	err := wrappedTestHandler(s.echoRouter.NewContext(req, s.recorder))

	assert.Nil(s.T(), err)
	assert.Equal(s.T(), http.StatusOK, s.recorder.Code)
	assert.Equal(s.T(), "value", s.recorder.Body.String())
}

func CreateTestEndpoint() customHTTP.Handler {
	return func(c echo.Context) error {
		appCtx := GetAppContext(c)
		if appCtx.GetCorrelationID() != "" {
			return c.String(http.StatusOK, appCtx.GetCorrelationID())
		}

		return c.String(http.StatusBadRequest, "error")
	}
}
