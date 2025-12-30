package http

import (
	"github.com/gorilla/mux"
)

// SetupRoutes configures all HTTP routes
func SetupRoutes(handler *TelemetryHandler) *mux.Router {
	router := mux.NewRouter()

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Telemetry endpoints
	api.HandleFunc("/telemetry/{batteryId}/current", handler.GetCurrentState).Methods("GET")
	api.HandleFunc("/telemetry/{batteryId}/history", handler.GetHistory).Methods("GET")

	return router
}
