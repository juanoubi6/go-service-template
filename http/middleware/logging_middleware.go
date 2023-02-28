package middleware

import (
	"fmt"
	customHTTP "go-service-template/http"
	"go-service-template/log"
	"go-service-template/utils"
	"net/http"
	"time"
)

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func CreateLoggingMiddleware() customHTTP.Middleware {
	logger := log.GetStdLogger("LoggingMiddleware")
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.RequestURI == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			appCtx := GetAppContext(r)

			logger.Info(
				"LoggingMiddleware",
				appCtx.GetCorrelationID(),
				fmt.Sprintf("Request to %v ended with status %v", r.URL.EscapedPath(), wrapped.status),
				utils.GenericParam{
					Key: "RequestMetadata",
					Value: map[string]interface{}{
						"HTTP Status": wrapped.status,
						"HTTP Method": r.Method,
						"Path":        r.URL.EscapedPath(),
						"Duration":    time.Since(start).String(),
					}},
			)
		}

		return http.HandlerFunc(fn)
	}
}
