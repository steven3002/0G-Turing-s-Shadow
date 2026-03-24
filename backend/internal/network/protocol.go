// internal/network/protocol.go
package network

import "encoding/json"

// ClientRequest is the base wrapper for all incoming AI actions.
type ClientRequest struct {
	Action  string          `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

// Specific Action Payloads
type MovePayload struct {
	Destination string `json:"destination"`
}

type TaskPayload struct {
	TaskID string `json:"task_id"`
}

type SabotagePayload struct {
	TargetRoom string `json:"target_room"`
}

type KillPayload struct {
	TargetID string `json:"target_id"`
}

// ServerResponse is used to send validation errors back to the AI.
type ServerResponse struct {
	ResponseType string `json:"response_type"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

type ChatPayload struct {
	Message string `json:"message"`
}

type VotePayload struct {
	TargetID string `json:"target_id"` // Can be a player ID or "SKIP"
}
