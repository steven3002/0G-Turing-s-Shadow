// internal/engine/win.go
package engine

import (
	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/state"
)

// EvaluateWinCondition checks the mathematical balance of the lobby.
func (e *Engine) EvaluateWinCondition() {
	players := e.gameState.GetAllPlayersUnsafe()
	impostorsAlive := 0
	crewmatesAlive := 0

	for _, p := range players {
		p.Mu.RLock()
		if p.IsAlive {
			if p.Role == state.RoleImpostor {
				impostorsAlive++
			} else {
				crewmatesAlive++
			}
		}
		p.Mu.RUnlock()
	}

	// Win Condition 1: All Impostors Dead (Crewmate Win)
	if impostorsAlive == 0 {
		e.gameState.Mu.Lock()
		e.gameState.CurrentPhase = state.PhaseResolved
		e.gameState.Mu.Unlock()
		e.broadcastMeetingResult("CREWMATES WIN! The Impostor was eliminated.")
		return
	}

	// Win Condition 2: Impostors tie or outnumber Crewmates (Impostor Win)
	if impostorsAlive >= crewmatesAlive {
		e.gameState.Mu.Lock()
		e.gameState.CurrentPhase = state.PhaseResolved
		e.gameState.Mu.Unlock()
		e.broadcastMeetingResult("IMPOSTOR WINS! The crew has been decimated.")
		return
	}
}
