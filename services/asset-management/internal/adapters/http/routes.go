package http

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all HTTP routes and middleware
func SetupRoutes(handler *BatteryHandler, healthHandler *HealthHandler) *mux.Router {
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
	router.Use(loggingMiddleware)
	router.Use(recoveryMiddleware)

	return router
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// recoveryMiddleware recovers from panics
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
