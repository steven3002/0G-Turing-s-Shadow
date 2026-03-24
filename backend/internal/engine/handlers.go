// internal/engine/handlers.go
package engine

import (
	"fmt"
	"time"

	"log"

	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/network"
	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/state"
)

// HandleMove validates adjacency and applies the 1.5s migration lock.
func (e *Engine) HandleMove(playerID string, payload network.MovePayload) error {
	player, err := e.gameState.GetPlayer(playerID)
	if err != nil {
		return err
	}

	// 1. Thread-safe validation block
	player.Mu.RLock()
	if !player.IsAlive || player.ActionStatus != state.StatusIdle {
		player.Mu.RUnlock()
		return fmt.Errorf("action rejected: dead or not idle")
	}
	if time.Now().UnixMilli() < player.MigrationUnlocksAt {
		player.Mu.RUnlock()
		return fmt.Errorf("action rejected: migration cooldown active")
	}
	currentRoomName := player.CurrentNode
	player.Mu.RUnlock()

	// 2. Map Topology Validation
	currentRoom, err := e.gameState.GetRoom(currentRoomName)
	if err != nil {
		return err
	}

	isValidMove := false
	currentRoom.Mu.RLock()
	for _, adj := range currentRoom.AdjacentNodes {
		if string(adj) == payload.Destination {
			isValidMove = true
			break
		}
	}
	currentRoom.Mu.RUnlock()

	if !isValidMove {
		return fmt.Errorf("invalid move: %s is not adjacent to %s", payload.Destination, currentRoomName)
	}

	targetRoom, err := e.gameState.GetRoom(state.RoomName(payload.Destination))
	if err != nil {
		return err
	}

	// 3. State Mutation Pipeline
	// Remove from old room
	currentRoom.RemoveFromAttendance(playerID)
	currentRoom.AddEvent(state.LogEvent{
		EventID:   fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Action:    "EXITED",
		ActorID:   playerID,                      // (Add shapeshift masking logic here later)
		ExpiresAt: time.Now().UnixMilli() + 2000, // 2.0s fast decay
	})

	// Add to new room
	targetRoom.AddToAttendance(playerID)
	targetRoom.AddEvent(state.LogEvent{
		EventID:   fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Action:    "ENTERED",
		ActorID:   playerID,
		ExpiresAt: time.Now().UnixMilli() + 2000,
	})

	// Update Player Location & Lock them for 1.5s
	player.UpdateLocation(state.RoomName(payload.Destination))
	player.ApplyCooldown("MIGRATION", time.Now().UnixMilli()+1500)

	// 1. Broadcast the exit event to everyone remaining in the OLD room
	e.BroadcastRoomState(currentRoomName)

	// 2. Broadcast the entry event to everyone in the NEW room
	e.BroadcastRoomState(state.RoomName(payload.Destination))

	// 3. Send the specific agent their new Local Map so they can plan their next move
	e.SendLocalMapUpdate(playerID, state.RoomName(payload.Destination))

	log.Printf("[ENGINE] %s moved to %s", playerID, payload.Destination)
	return nil
}

// HandleKill validates the 5s cooldown and room overlap.
func (e *Engine) HandleKill(playerID string, payload network.KillPayload) error {
	killer, err := e.gameState.GetPlayer(playerID)
	if err != nil {
		return err
	}

	killer.Mu.RLock()
	if killer.Role != state.RoleImpostor {
		killer.Mu.RUnlock()
		return fmt.Errorf("unauthorized: only impostor can kill")
	}
	if !killer.IsAlive || killer.ActionStatus != state.StatusIdle {
		killer.Mu.RUnlock()
		return fmt.Errorf("action rejected: dead or not idle")
	}
	if time.Now().UnixMilli() < killer.KillUnlocksAt {
		killer.Mu.RUnlock()
		return fmt.Errorf("action rejected: kill cooldown active")
	}
	killerRoom := killer.CurrentNode
	killer.Mu.RUnlock()

	target, err := e.gameState.GetPlayer(payload.TargetID)
	if err != nil {
		return err
	}

	target.Mu.RLock()
	if !target.IsAlive || target.CurrentNode != killerRoom {
		target.Mu.RUnlock()
		return fmt.Errorf("action rejected: target dead or not in room")
	}
	target.Mu.RUnlock()

	// Mutate State
	target.Kill()
	killer.ApplyCooldown("KILL", time.Now().UnixMilli()+5000)

	room, _ := e.gameState.GetRoom(killerRoom)
	room.RemoveFromAttendance(payload.TargetID)
	room.AddEvent(state.LogEvent{
		EventID:   fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		Action:    "ASSASSINATION",
		ActorID:   "UNKNOWN", // Kills don't explicitly name the killer in the log
		ExpiresAt: time.Now().UnixMilli() + 2000,
	})

	log.Printf("[ENGINE] %s assassinated %s in %s", playerID, payload.TargetID, killerRoom)
	return nil
}

// HandleTask locks the agent for 3s (Sensory Deprivation)
func (e *Engine) HandleTask(playerID string, payload network.TaskPayload) error {
	player, err := e.gameState.GetPlayer(playerID)
	if err != nil {
		return err
	}

	player.Mu.RLock()
	if player.Role == state.RoleImpostor {
		player.Mu.RUnlock()
		return fmt.Errorf("impostors cannot execute tasks")
	}
	if !player.IsAlive || player.ActionStatus != state.StatusIdle {
		player.Mu.RUnlock()
		return fmt.Errorf("action rejected: dead or busy")
	}
	player.Mu.RUnlock()

	// Lock the player (This triggers "Task Blindness" when pushing sensory logs)
	player.UpdateStatus(state.StatusDoingTask)
	player.ApplyCooldown("TASK", time.Now().UnixMilli()+3000)

	// Note: The global 100ms Engine Tick will check for expired TASK cooldowns
	// and transition them back to IDLE, pushing the 1.2s anonymous success pulse.

	log.Printf("[ENGINE] %s started task %s", playerID, payload.TaskID)
	return nil
}

// internal/engine/handlers.go (Add to the bottom)

// HandleSabotage allows an Impostor to trigger a room-wide alarm from anywhere on the map.
func (e *Engine) HandleSabotage(playerID string, payload network.SabotagePayload) error {
	impostor, err := e.gameState.GetPlayer(playerID)
	if err != nil {
		return err
	}

	impostor.Mu.RLock()
	if impostor.Role != state.RoleImpostor {
		impostor.Mu.RUnlock()
		return fmt.Errorf("unauthorized: only impostors can sabotage")
	}
	if !impostor.IsAlive || impostor.ActionStatus != state.StatusIdle {
		impostor.Mu.RUnlock()
		return fmt.Errorf("action rejected: dead or not idle")
	}
	// We elegantly repurpose the Task timer for the 15s Sabotage Cooldown
	if time.Now().UnixMilli() < impostor.TaskCompletesAt {
		impostor.Mu.RUnlock()
		return fmt.Errorf("action rejected: sabotage cooldown active")
	}
	impostor.Mu.RUnlock()

	targetRoomName := state.RoomName(payload.TargetRoom)
	targetRoom, err := e.gameState.GetRoom(targetRoomName)
	if err != nil {
		return err
	}

	// 1. Apply 15s Cooldown
	impostor.ApplyCooldown("TASK", time.Now().UnixMilli()+15000)

	// 2. Inject the Sabotage Alarm (Lasts 10 seconds)
	targetRoom.AddEvent(state.LogEvent{
		EventID:   fmt.Sprintf("sabotage_%d", time.Now().UnixNano()),
		Action:    "CRITICAL_SABOTAGE_ALARM",
		ActorID:   "SYSTEM",
		ExpiresAt: time.Now().UnixMilli() + 10000,
	})

	// 3. Instantly broadcast the chaos to everyone currently in the target room
	e.BroadcastRoomState(targetRoomName)

	log.Printf("[ENGINE] %s triggered a sabotage in %s", playerID, payload.TargetRoom)
	return nil
}
