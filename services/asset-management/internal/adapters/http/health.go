package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db       *sql.DB
	natsConn *nats.Conn
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(db *sql.DB, natsConn *nats.Conn) *HealthHandler {
	return &HealthHandler{
		db:       db,
		natsConn: natsConn,
	}
}

// LivenessCheck checks if the service is running
// Returns 200 if the service is alive (can handle requests)
// This endpoint should never fail unless the service is completely down
func (h *HealthHandler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"service": "asset-management",
	})
}

// ReadinessCheck checks if the service is ready to handle traffic
// Returns 200 if all dependencies are healthy
// Returns 503 if any dependency is unavailable
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Check database connection
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "unavailable",
			"service": "asset-management",
			"checks": map[string]string{
				"database": "unhealthy: " + err.Error(),
				"nats":     "not checked",
			},
		})
		return
	}

	// Check NATS connection (optional - service can work without NATS)
	natsStatus := "healthy"
	if h.natsConn != nil && !h.natsConn.IsConnected() {
		natsStatus = "unhealthy (optional)"
	} else if h.natsConn == nil {
		natsStatus = "not configured"
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ready",
		"service": "asset-management",
		"checks": map[string]string{
			"database": "healthy",
			"nats":     natsStatus,
		},
	})
}
