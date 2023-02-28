package middleware

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type AppContextMiddlewareSuite struct {
	suite.Suite
	appContextMiddleware func(http.Handler) http.Handler
	testEndpoint         http.Handler
}

func (sut *AppContextMiddlewareSuite) SetupSuite() {
	sut.appContextMiddleware = CreateAppContextMiddleware()
	sut.testEndpoint = CreateTestEndpoint()
}

func TestAppContextMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(AppContextMiddlewareSuite))
}

func (sut *AppContextMiddlewareSuite) Test_AppContextMiddleware_DecoratesRequestAppContext() {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)

	wrappedTestHandler := sut.appContextMiddleware(sut.testEndpoint)

	wrappedTestHandler.ServeHTTP(res, req)

	assert.Equal(sut.T(), http.StatusOK, res.Code)
}

func (sut *AppContextMiddlewareSuite) Test_AppContextMiddleware_UsesCorrelationIDHeaderAsCorrelationID() {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Add(CorrelationIDHeader, "value")

	wrappedTestHandler := sut.appContextMiddleware(sut.testEndpoint)

	wrappedTestHandler.ServeHTTP(res, req)

	assert.Equal(sut.T(), http.StatusOK, res.Code)
	assert.Equal(sut.T(), "value", res.Body.String())
}

func CreateTestEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		appCtx := GetAppContext(r)
		if appCtx.GetCorrelationID() != "" {
			w.Write([]byte(appCtx.GetCorrelationID()))
			w.WriteHeader(200)
		} else {
			w.WriteHeader(400)
		}
	}
}
