package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aegisbox/aegisbox/internal/executor"
	"github.com/aegisbox/aegisbox/internal/logging"
)

// NewRouter constructs and configures the HTTP multiplexer with middleware.
func NewRouter(mgr *executor.ExecutionManager, resolver executor.RuntimeResolver, logger *logging.Logger) http.Handler {
	handler := NewHandler(mgr, resolver, logger)
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.Health)
	mux.HandleFunc("/execute", handler.Execute)

	// Wrap in logging and recovery middleware
	return recoveryMiddleware(loggingMiddleware(mux, logger), logger)
}

func loggingMiddleware(next http.Handler, logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("http.request", map[string]interface{}{
			"method":      r.Method,
			"path":        r.URL.Path,
			"remote_addr": r.RemoteAddr,
			"duration_ms": time.Since(start).Milliseconds(),
		})
	})
}

func recoveryMiddleware(next http.Handler, logger *logging.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				errStr := fmt.Sprintf("%v", rec)
				logger.Error("panic recovered in HTTP handler", map[string]interface{}{
					"panic": errStr,
					"path":  r.URL.Path,
				})
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{
					Error:   "Internal Server Error",
					Details: errStr,
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
