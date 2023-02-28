package middleware

import (
	"context"
	"go-service-template/domain"
	customHTTP "go-service-template/http"
	"net/http"
)

type ContextKey string

const AppContextKey ContextKey = "appContextKey"
const CorrelationIDHeader = "Correlation-Id"

func CreateAppContextMiddleware() customHTTP.Middleware {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			appCtx := domain.CreateAppContext(r.Context(), r.Header.Get(CorrelationIDHeader))
			r = r.WithContext(context.WithValue(r.Context(), AppContextKey, appCtx))
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func GetAppContext(r *http.Request) *domain.AppContext {
	appCtx, ok := r.Context().Value(AppContextKey).(*domain.AppContext)
	if ok {
		return appCtx
	}

	return domain.CreateAppContext(r.Context(), "")
}
