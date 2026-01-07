// Package logger provides application-wide logging facilities.
// It defines a singleton logger instance and HTTP middleware
// for logging incoming requests.
package logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Log is a global singleton logger instance.
// It is available to the entire application and must not be modified
// directly by any code except the Initialize function.
// By default, a no-op logger is used that produces no output.
var Log *zap.Logger = zap.NewNop()

// Initialize initializes the global logger singleton with the specified
// logging level. The level must be a valid zap logging level string
// (e.g. "DEBUG", "INFO", "WARN", "ERROR").
//
// This function must be called once during application startup.
func Initialize(level string) error {
	// Convert textual log level to zap atomic level.
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	// Create production logger configuration.
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	// Build logger from configuration.
	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	// Set global singleton logger.
	Log = zl
	return nil
}

// RequestLogger is an HTTP middleware that logs incoming requests.
// It records request method, URI, response status code, response size,
// and request processing duration using the global logger instance.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wrap response writer to capture status code and response size.
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		start := time.Now()
		next.ServeHTTP(ww, r)
		duration := time.Since(start)

		Log.Info("incoming request",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Int("status", ww.Status()),
			zap.Int("size", ww.BytesWritten()),
			zap.Duration("duration", duration),
		)
	})
}
