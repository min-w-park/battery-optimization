package events

import "time"

// ConflictResolved is published when a conflict resolution decision is made
// Publisher: Operator Service (manual) or Auto-Resolver Service (auto mode)
// Subscribers: Device Interface Service, Bidding Service
type ConflictResolved struct {
	ConflictID    string    `json:"conflict_id"`
	ChosenAction  string    `json:"chosen_action"`         // CONTINUE_CURRENT | SWITCH_TO_CHARGING
	DecisionMaker string    `json:"decision_maker"`        // OPERATOR | AUTO
	OperatorID    string    `json:"operator_id,omitempty"` // if manual
	Timestamp     time.Time `json:"timestamp"`
	EventVersion  string    `json:"event_version"` // "v1"
}
