package persist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/state"
)

type PlayerSnapshot struct {
	ID           string             `json:"id"`
	Role         state.Role         `json:"role"`
	IsAlive      bool               `json:"is_alive"`
	CurrentNode  state.RoomName     `json:"current_node"`
	ActionStatus state.ActionStatus `json:"action_status"`
}

type RoomSnapshot struct {
	Name         state.RoomName    `json:"name"`
	Logs         []state.LogEvent  `json:"logs"`
	Tasks        map[string]string `json:"tasks"` // Simplified task status
	Attendance   []string          `json:"attendance"`
}

type GameSnapshot struct {
	MatchID           string          `json:"match_id"`
	SnapshotAtUnixMs  int64           `json:"snapshot_at_unix_ms"`
	TriggerPhase      state.Phase     `json:"trigger_phase"`
	Players           []PlayerSnapshot `json:"players"`
	Rooms             []RoomSnapshot   `json:"rooms"`
	ActiveMeeting     *state.MeetingState `json:"active_meeting,omitempty"`
}

func SnapshotGameState(gs *state.GameStateManager, triggerPhase state.Phase, outputDir string) (string, error) {
	// 1. Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	snapshot := GameSnapshot{
		MatchID:          gs.MatchID,
		SnapshotAtUnixMs: time.Now().UnixMilli(),
		TriggerPhase:     triggerPhase,
		Players:          make([]PlayerSnapshot, 0),
		Rooms:            make([]RoomSnapshot, 0),
	}

	// 2. Snapshot Players
	players := gs.GetAllPlayersUnsafe()
	for _, p := range players {
		p.Mu.RLock()
		snapshot.Players = append(snapshot.Players, PlayerSnapshot{
			ID:           p.ID,
			Role:         p.Role,
			IsAlive:      p.IsAlive,
			CurrentNode:  p.CurrentNode,
			ActionStatus: p.ActionStatus,
		})
		p.Mu.RUnlock()
	}

	// 3. Snapshot Rooms
	rooms := gs.GetAllRooms()
	for _, r := range rooms {
		r.Mu.RLock()
		roomSnap := RoomSnapshot{
			Name:       r.Name,
			Logs:       r.GetActiveLogs(), // Already thread-safe copy
			Tasks:      make(map[string]string),
			Attendance: r.GetAttendanceSnapshot(), // Already thread-safe copy
		}
		for id, status := range r.Tasks {
			roomSnap.Tasks[id] = status.Status
		}
		snapshot.Rooms = append(snapshot.Rooms, roomSnap)
		r.Mu.RUnlock()
	}

	// 4. Snapshot Meeting
	gs.Mu.RLock()
	snapshot.ActiveMeeting = gs.ActiveMeeting
	gs.Mu.RUnlock()

	// 5. Serialize to JSON
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// 6. Write to File
	filename := fmt.Sprintf("match_%s_%s_%d.json", gs.MatchID, triggerPhase, snapshot.SnapshotAtUnixMs)
	filePath := filepath.Join(outputDir, filename)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return filePath, nil
}
