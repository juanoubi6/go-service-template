package middleware

import (
	"fmt"
	customHTTP "go-service-template/http"
	"go-service-template/monitor"
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
	logger := monitor.GetStdLogger("LoggingMiddleware")
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
				monitor.LoggingParam{
					Name: "request_metadata",
					Value: map[string]interface{}{
						"http_status": wrapped.status,
						"http_method": r.Method,
						"path":        r.URL.EscapedPath(),
						"duration":    time.Since(start).String(),
					}},
			)
		}

		return http.HandlerFunc(fn)
	}
}
