package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
)

// REST API Helpers

// CreateBattery registers a new battery via Asset Management API
func CreateBattery(baseURL string, capacity, maxPower, rampRate, efficiency float64) (string, error) {
	payload := map[string]interface{}{
		"capacity":     capacity,
		"max_power":    maxPower,
		"ramp_rate":    rampRate,
		"efficiency":   efficiency,
		"location":     "SA",
		"manufacturer": "Tesla",
		"constraints": map[string]interface{}{
			"warranty_eol":          0.8,
			"max_cycles":            10000,
			"operating_temp_min":    -20.0,
			"operating_temp_max":    60.0,
			"grid_compliance_level": "FCAS",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(baseURL+"/api/v1/batteries", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	batteryID, ok := result["id"].(string)
	if !ok {
		return "", fmt.Errorf("battery ID not found in response")
	}

	return batteryID, nil
}

// CreatePrice creates a new market price via Market Data API
func CreatePrice(baseURL string, price float64, interval int) error {
	intervalStart := time.Now().Add(24 * time.Hour).Truncate(5 * time.Minute)
	publishedAt := time.Now().Add(-5 * time.Minute)

	payload := map[string]interface{}{
		"region":         "NSW",
		"price":          price,
		"demand":         8200.0,
		"interval_type":  "5MIN_PREDISPATCH",
		"interval_start": intervalStart.Format(time.RFC3339),
		"published_at":   publishedAt.Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(baseURL+"/api/v1/prices", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// GetBattery retrieves a battery via Asset Management API
func GetBattery(baseURL string, batteryID string) (map[string]interface{}, error) {
	resp, err := http.Get(baseURL + "/api/v1/batteries/" + batteryID)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// NATS Helpers

// PublishBatteryState publishes a battery state event to NATS
func PublishBatteryState(nc *nats.Conn, batteryID string, soc, power float64, state string) error {
	event := map[string]interface{}{
		"battery_id":      batteryID,
		"soc":             soc,
		"power":           power,
		"operation_state": state,
		"temperature":     25.0,
		"voltage":         400.0,
		"current":         power / 400.0,
		"timestamp":       time.Now().Format(time.RFC3339),
		"event_version":   "v1",
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := nc.Publish("battery.state.changed.v1", data); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

// SubscribeToEvents subscribes to NATS events with a wildcard pattern
func SubscribeToEvents(nc *nats.Conn, pattern string) (*nats.Subscription, error) {
	ch := make(chan *nats.Msg, 100)
	sub, err := nc.ChanSubscribe(pattern, ch)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}
	return sub, nil
}

// WaitForEvent waits for a specific event subject with timeout
func WaitForEvent(sub *nats.Subscription, subject string, timeout time.Duration) (*nats.Msg, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		msg, err := sub.NextMsg(100 * time.Millisecond)
		if err != nil {
			if err == nats.ErrTimeout {
				continue
			}
			return nil, fmt.Errorf("failed to receive message: %w", err)
		}

		if msg.Subject == subject {
			return msg, nil
		}
	}

	return nil, fmt.Errorf("timeout waiting for event: %s", subject)
}

// WaitForAnyEvent waits for any event matching the subscription pattern
func WaitForAnyEvent(sub *nats.Subscription, timeout time.Duration) (*nats.Msg, error) {
	msg, err := sub.NextMsg(timeout)
	if err != nil {
		if err == nats.ErrTimeout {
			return nil, fmt.Errorf("timeout waiting for any event")
		}
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}
	return msg, nil
}

// AssertNoEvent verifies that no event is received within timeout
func AssertNoEvent(sub *nats.Subscription, timeout time.Duration) error {
	msg, err := sub.NextMsg(timeout)
	if err == nats.ErrTimeout {
		return nil // Expected - no event received
	}
	if err != nil {
		return fmt.Errorf("unexpected error: %w", err)
	}
	return fmt.Errorf("unexpected event received: %s", msg.Subject)
}

// DrainEvents drains all pending events from a subscription
func DrainEvents(sub *nats.Subscription) {
	for {
		_, err := sub.NextMsg(10 * time.Millisecond)
		if err == nats.ErrTimeout {
			break
		}
	}
}

// ConnectToNATS connects to NATS server with retry logic
func ConnectToNATS(url string, retries int, retryDelay time.Duration) (*nats.Conn, error) {
	var nc *nats.Conn
	var err error

	for i := 0; i < retries; i++ {
		nc, err = nats.Connect(url)
		if err == nil {
			return nc, nil
		}

		if i < retries-1 {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("failed to connect to NATS after %d retries: %w", retries, err)
}

// WaitForHTTPService waits for an HTTP service to be ready
func WaitForHTTPService(baseURL string, retries int, retryDelay time.Duration) error {
	for i := 0; i < retries; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		if i < retries-1 {
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("service not ready after %d retries", retries)
}
