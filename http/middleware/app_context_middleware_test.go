package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type AppContextMiddlewareSuite struct {
	suite.Suite
	appContextMiddleware echo.MiddlewareFunc
	testEndpoint         echo.HandlerFunc
	echoRouter           *echo.Echo
}

func (sut *AppContextMiddlewareSuite) SetupSuite() {
	sut.appContextMiddleware = CreateAppContextMiddleware()
	sut.testEndpoint = CreateTestEndpoint()
	sut.echoRouter = echo.New()
}

func TestAppContextMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(AppContextMiddlewareSuite))
}

func (sut *AppContextMiddlewareSuite) Test_AppContextMiddleware_DecoratesRequestAppContext() {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)

	wrappedTestHandler := sut.appContextMiddleware(sut.testEndpoint)

	err := wrappedTestHandler(sut.echoRouter.NewContext(req, res))

	assert.Nil(sut.T(), err)
	assert.Equal(sut.T(), http.StatusOK, res.Code)
}

func (sut *AppContextMiddlewareSuite) Test_AppContextMiddleware_UsesCorrelationIDHeaderAsCorrelationID() {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Add(CorrelationIDHeader, "value")

	wrappedTestHandler := sut.appContextMiddleware(sut.testEndpoint)

	err := wrappedTestHandler(sut.echoRouter.NewContext(req, res))

	assert.Nil(sut.T(), err)
	assert.Equal(sut.T(), http.StatusOK, res.Code)
	assert.Equal(sut.T(), "value", res.Body.String())
}

func CreateTestEndpoint() echo.HandlerFunc {
	return func(c echo.Context) error {
		appCtx := GetAppContext(c)
		if appCtx.GetCorrelationID() != "" {
			return c.String(http.StatusOK, appCtx.GetCorrelationID())
		}

		return c.String(http.StatusBadRequest, "error")
	}
}
