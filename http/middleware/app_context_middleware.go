package middleware

import (
	"context"
	customHTTP "go-service-template/http"
	"go-service-template/monitor"
	"net/http"
)

type ContextKey string

const AppContextKey ContextKey = "appContextKey"
const CorrelationIDHeader = "Correlation-Id"

func CreateAppContextMiddleware() customHTTP.Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			appCtx := monitor.CreateAppContextFromRequest(r, r.Header.Get(CorrelationIDHeader))
			r = r.WithContext(context.WithValue(r.Context(), AppContextKey, appCtx))
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func GetAppContext(r *http.Request) *monitor.AppContext {
	appCtx, ok := r.Context().Value(AppContextKey).(*monitor.AppContext)
	if ok {
		return appCtx
	}

	return monitor.CreateAppContextFromRequest(r, "")
}
