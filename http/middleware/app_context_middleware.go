package middleware

import (
	"context"
	"github.com/labstack/echo/v4"
	"go-service-template/monitor"
	"net/http"
)

type ContextKey string

const AppContextKey ContextKey = "appContextKey"
const CorrelationIDHeader = "Correlation-Id"

func CreateAppContextMiddleware() echo.MiddlewareFunc {
	return echo.WrapMiddleware(
		func(next http.Handler) http.Handler {
			fn := func(w http.ResponseWriter, r *http.Request) {
				appCtx := monitor.CreateAppContextFromRequest(r, r.Header.Get(CorrelationIDHeader))
				r = r.WithContext(context.WithValue(r.Context(), AppContextKey, appCtx))
				next.ServeHTTP(w, r)
			}

			return http.HandlerFunc(fn)
		},
	)
}

func GetAppContext(c echo.Context) *monitor.AppContext {
	appCtx, ok := c.Request().Context().Value(AppContextKey).(*monitor.AppContext)
	if ok {
		return appCtx
	}

	return monitor.CreateAppContextFromRequest(c.Request(), "")
}
