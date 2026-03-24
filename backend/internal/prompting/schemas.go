// internal/prompting/schemas.go
package prompting

import "github.com/steven3002/0G-Turing-s-Shadow/backend/internal/state"

// Prompt05StateTickUpdate is pushed when room state changes.
type Prompt05StateTickUpdate struct {
	PromptID    string           `json:"prompt_id"`
	Type        string           `json:"type"`
	Room        string           `json:"room"`
	Attendance  []string         `json:"attendance"`
	ActivityLog []state.LogEvent `json:"activity_log"`
}

// Prompt06LocalMapUpdate is pushed immediately after a successful MOVE.
type Prompt06LocalMapUpdate struct {
	PromptID      string   `json:"prompt_id"`
	Type          string   `json:"type"`
	CurrentNode   string   `json:"current_node"`
	AdjacentNodes []string `json:"adjacent_nodes"`
}

// Prompt01Welcome is pushed immediately after handshake.
type Prompt01Welcome struct {
	PromptID string   `json:"prompt_id"`
	Type     string   `json:"type"`
	PlayerID string   `json:"player_id"`
	Room     string   `json:"room"`
	Phase    string   `json:"phase"`
	Players  []string `json:"players"` // All player IDs in match
}
