package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// SetupRoutes configures all HTTP routes and middleware
func SetupRoutes(handler *BatteryHandler, healthHandler *HealthHandler, log *zap.Logger) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoints (no /api/v1 prefix)
	router.HandleFunc("/health/live", healthHandler.LivenessCheck).Methods(http.MethodGet)
	router.HandleFunc("/health/ready", healthHandler.ReadinessCheck).Methods(http.MethodGet)

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Battery endpoints
	api.HandleFunc("/batteries", handler.CreateBattery).Methods(http.MethodPost)
	api.HandleFunc("/batteries/{id}", handler.GetBattery).Methods(http.MethodGet)
	api.HandleFunc("/batteries", handler.ListBatteries).Methods(http.MethodGet)

	// Add middleware
	router.Use(loggingMiddleware(log))
	router.Use(recoveryMiddleware(log))

	return router
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			log.Info("HTTP request started",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
			)

			next.ServeHTTP(w, r)

			log.Info("HTTP request completed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

// recoveryMiddleware recovers from panics
func recoveryMiddleware(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic recovered",
						zap.Any("error", err),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
