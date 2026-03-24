// internal/engine/broadcaster.go
package engine

import (
	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/prompting"
	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/state"
)

// SendLocalMapUpdate pushes PROMPT_06 to a specific player.
func (e *Engine) SendLocalMapUpdate(playerID string, roomName state.RoomName) {
	room, err := e.gameState.GetRoom(roomName)
	if err != nil {
		return
	}

	// Convert typed RoomNames to strings for the JSON schema
	adjacents := make([]string, 0, len(room.AdjacentNodes))
	for _, adj := range room.AdjacentNodes {
		adjacents = append(adjacents, string(adj))
	}

	payload := prompting.Prompt06LocalMapUpdate{
		PromptID:      "PROMPT_06_LOCAL_MAP_UPDATE",
		Type:          "sensory_injection",
		CurrentNode:   string(roomName),
		AdjacentNodes: adjacents,
	}

	if e.network != nil {
		e.network.SendToPlayer(playerID, payload)
	}
}

// BroadcastRoomState pushes PROMPT_05 to all eligible agents in a specific room.
func (e *Engine) BroadcastRoomState(roomName state.RoomName) {
	room, err := e.gameState.GetRoom(roomName)
	if err != nil {
		return
	}

	// Fetch thread-safe snapshots of the highly volatile data
	attendance := room.GetAttendanceSnapshot()
	activeLogs := room.GetActiveLogs()

	payload := prompting.Prompt05StateTickUpdate{
		PromptID:    "PROMPT_05_STATE_TICK_UPDATE",
		Type:        "sensory_injection",
		Room:        string(roomName),
		Attendance:  attendance,
		ActivityLog: activeLogs,
	}

	// Iterate through everyone in the room and enforce Task Blindness
	for _, targetID := range attendance {
		player, err := e.gameState.GetPlayer(targetID)
		if err != nil {
			continue
		}

		player.Mu.RLock()
		isBlind := player.ActionStatus == state.StatusDoingTask
		player.Mu.RUnlock()

		// THE FOG OF WAR: Do not send the payload if they are executing a task!
		if !isBlind && e.network != nil {
			e.network.SendToPlayer(targetID, payload)
		}
	}
}
