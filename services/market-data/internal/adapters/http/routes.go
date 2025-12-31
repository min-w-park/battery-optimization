package http

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// SetupRoutes configures HTTP routes and middleware
func SetupRoutes(handler *MarketPriceHandler, healthHandler *HealthHandler) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoints
	router.HandleFunc("/health/live", healthHandler.LivenessCheck).Methods(http.MethodGet)
	router.HandleFunc("/health/ready", healthHandler.ReadinessCheck).Methods(http.MethodGet)

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/prices", handler.CreateMarketPrice).Methods(http.MethodPost)
	api.HandleFunc("/prices/{id}", handler.GetMarketPrice).Methods(http.MethodGet)
	api.HandleFunc("/prices", handler.ListMarketPrices).Methods(http.MethodGet)

	// Middleware
	router.Use(loggingMiddleware)
	router.Use(recoveryMiddleware)

	return router
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

// recoveryMiddleware recovers from panics
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
