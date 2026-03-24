// internal/engine/matchmaker.go
package engine

import (
	"log"
	"math/rand"
	"time"

	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/state"
)

// CheckLobbyReadiness evaluates if the game should start.
func (e *Engine) CheckLobbyReadiness() {
	// 1. Safely check phase without a global Write Lock
	e.gameState.Mu.RLock()
	phase := e.gameState.CurrentPhase
	e.gameState.Mu.RUnlock()

	if phase != state.PhaseLobby {
		return
	}

	// 2. Safely get the players (GetAllPlayersUnsafe handles its own RLock)
	players := e.gameState.GetAllPlayersUnsafe()
	playerCount := len(players)

	if playerCount < 9 {
		log.Printf("[MATCHMAKER] Lobby at %d/9 players.", playerCount)
		return
	}

	log.Println("[MATCHMAKER] Lobby full. Initializing Match...")
	e.initializeMatch(players) // Pass the players slice in!
}

// initializeMatch assigns roles, teleports agents, and starts the spatial loop.
func (e *Engine) initializeMatch(players []*state.PlayerState) {
	// 1. Shuffle players for random role assignment
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(players), func(i, j int) { players[i], players[j] = players[j], players[i] })

	// 2. Assign Roles (1 Impostor, 8 Crewmates)
	impostorAssigned := false
	e.gameState.TotalCrewmatesAlive = 8
	e.gameState.TotalImpostorsAlive = 1

	for _, p := range players {
		p.Mu.Lock()
		if !impostorAssigned {
			p.Role = state.RoleImpostor
			impostorAssigned = true
		} else {
			p.Role = state.RoleCrewmate
		}

		// Reset state
		p.IsAlive = true
		p.CurrentNode = state.RoomCafeteria
		p.ActionStatus = state.StatusIdle

		// Reset all cooldowns
		now := time.Now().UnixMilli()
		p.MigrationUnlocksAt = now
		p.KillUnlocksAt = now
		p.TaskCompletesAt = now
		p.ChatUnlocksAt = now
		p.Mu.Unlock()
	}

	// 3. Move everyone into the Cafeteria Attendance Log safely
	cafeteria, _ := e.gameState.GetRoom(state.RoomCafeteria)
	for _, p := range players {
		cafeteria.AddToAttendance(p.ID)
	}

	// 4. Update Game Phase (NOW we write lock)
	e.gameState.Mu.Lock()
	e.gameState.CurrentPhase = state.PhaseInPlay
	e.gameState.Mu.Unlock()

	// 5. Broadcast Initialization Prompts
	go e.broadcastMatchStart(players)
}

func (e *Engine) broadcastMatchStart(players []*state.PlayerState) {
	for _, p := range players {
		if e.network != nil {
			if p.Role == state.RoleImpostor {
				e.network.SendToPlayer(p.ID, map[string]string{
					"prompt_id": "PROMPT_04_INIT_IMPOSTOR",
					"directive": "Eliminate the crew. You are the Shapeshifter.",
				})
			} else {
				e.network.SendToPlayer(p.ID, map[string]string{
					"prompt_id": "PROMPT_03_INIT_CREWMATE",
					"directive": "Survive and identify the anomaly.",
				})
			}
		}
		// Send them their first spatial view
		e.SendLocalMapUpdate(p.ID, state.RoomCafeteria)
	}
	// Finally, broadcast the sensory room state for the Cafeteria
	e.BroadcastRoomState(state.RoomCafeteria)
}
