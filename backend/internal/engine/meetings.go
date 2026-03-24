// internal/engine/meetings.go
package engine

import (
	"fmt"
	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/network"
	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/state"
	"time"
)

// TriggerMeeting pauses the game and boots up the NLP chat phase.
func (e *Engine) TriggerMeeting(callerID string, meetingType string) error {
	// 1. Gather the dead bodies BEFORE locking the main state to prevent deadlocks
	players := e.gameState.GetAllPlayersUnsafe()
	var deadBodies []string

	for _, p := range players {
		p.Mu.RLock()
		if !p.IsAlive {
			deadBodies = append(deadBodies, p.ID)
		}
		p.Mu.RUnlock()
	}

	// 2. Lock and Transition State
	e.gameState.Mu.Lock()
	if e.gameState.CurrentPhase != state.PhaseInPlay {
		e.gameState.Mu.Unlock()
		return fmt.Errorf("meeting rejected: game not in play state")
	}

	e.gameState.CurrentPhase = state.PhaseMeetingChat
	e.gameState.ActiveMeeting = &state.MeetingState{
		CallerID:        callerID,
		MeetingType:     meetingType,
		DeadDiscovered:  deadBodies, // TODO SOLVED: Dynamically populated
		ChatTranscript:  make([]string, 0),
		BallotBox:       make(map[string]string),
		PhaseEndsAtUnix: time.Now().UnixMilli() + 20000,
	}
	e.gameState.Mu.Unlock()

	// 3. Broadcast PROMPT_08_MEETING_START_CONTEXT to all alive players
	payload := map[string]interface{}{
		"prompt_id":       "PROMPT_08_MEETING_START_CONTEXT",
		"type":            "meeting_start",
		"caller_id":       callerID,
		"meeting_type":    meetingType,
		"dead_discovered": deadBodies,
		"message":         "A meeting has been called. You have 20 seconds to discuss before voting begins.",
	}

	for _, p := range players {
		p.Mu.RLock()
		isAlive := p.IsAlive
		p.Mu.RUnlock()

		if isAlive && e.network != nil {
			e.network.SendToPlayer(p.ID, payload)
		}
	}

	return nil
}

// HandlePanicButton enforces the Cafeteria-only rule.
func (e *Engine) HandlePanicButton(playerID string) error {
	player, _ := e.gameState.GetPlayer(playerID)

	player.Mu.RLock()
	inCafeteria := player.CurrentNode == state.RoomCafeteria
	player.Mu.RUnlock()

	if !inCafeteria {
		return fmt.Errorf("panic button can only be pressed in the CAFETERIA")
	}

	return e.TriggerMeeting(playerID, "PANIC")
}

// internal/engine/meetings.go (continued)

func (e *Engine) HandleChat(playerID string, payload network.ChatPayload) error {
	e.gameState.Mu.RLock()
	if e.gameState.CurrentPhase != state.PhaseMeetingChat && e.gameState.CurrentPhase != state.PhaseMeetingVote {
		e.gameState.Mu.RUnlock()
		return fmt.Errorf("chat rejected: no active meeting")
	}
	meeting := e.gameState.ActiveMeeting
	e.gameState.Mu.RUnlock()

	player, _ := e.gameState.GetPlayer(playerID)

	// 1. Enforce the 0.6s Anti-Spam Rate Limit
	player.Mu.RLock()
	if time.Now().UnixMilli() < player.ChatUnlocksAt {
		player.Mu.RUnlock()
		return fmt.Errorf("chat rejected: rate limit active (0.6s)")
	}
	if !player.IsAlive {
		player.Mu.RUnlock()
		return fmt.Errorf("dead agents cannot speak")
	}
	player.Mu.RUnlock()

	// 2. Process Chat
	player.ApplyCooldown("CHAT", time.Now().UnixMilli()+600)

	formattedMessage := fmt.Sprintf("[%s]: %s", playerID, payload.Message)

	e.gameState.Mu.Lock()
	meeting.ChatTranscript = append(meeting.ChatTranscript, formattedMessage)
	e.gameState.Mu.Unlock()

	// 3. Broadcast PROMPT_09_CHAT_EVALUATION_TICK to all alive players
	e.BroadcastChatUpdate(formattedMessage)

	return nil
}

func (e *Engine) HandleVote(playerID string, payload network.VotePayload) error {
	e.gameState.Mu.Lock()
	defer e.gameState.Mu.Unlock()

	if e.gameState.CurrentPhase != state.PhaseMeetingVote {
		return fmt.Errorf("vote rejected: not in voting phase")
	}

	// Lock in the vote (overwriting previous votes if they change their mind)
	e.gameState.ActiveMeeting.BallotBox[playerID] = payload.TargetID
	return nil
}

// internal/engine/meetings.go (continued)

func (e *Engine) ResolveMeeting() {
	e.gameState.Mu.Lock()
	meeting := e.gameState.ActiveMeeting
	e.gameState.Mu.Unlock() // Unlock early so we can safely interact with players

	voteCounts := make(map[string]int)
	for _, target := range meeting.BallotBox {
		voteCounts[target]++
	}

	// Find the highest voted target
	maxVotes := 0
	var ejectedID string
	isTie := false

	for target, count := range voteCounts {
		if count > maxVotes {
			maxVotes = count
			ejectedID = target
			isTie = false
		} else if count == maxVotes {
			isTie = true
		}
	}

	// Execution Logic
	if !isTie && ejectedID != "SKIP" && ejectedID != "" {
		player, err := e.gameState.GetPlayer(ejectedID)
		if err == nil {
			player.Kill()

			// Check if we caught the Impostor
			player.Mu.RLock()
			isImpostor := player.Role == state.RoleImpostor
			player.Mu.RUnlock()

			// Formulate and broadcast the reveal message
			revealMsg := fmt.Sprintf("%s was ejected. They were NOT the Impostor.", ejectedID)
			if isImpostor {
				revealMsg = fmt.Sprintf("%s was ejected. They WERE the Impostor.", ejectedID)
			}
			e.broadcastMeetingResult(revealMsg)

			// Evaluate Win Conditions
			if isImpostor {
				e.gameState.Mu.Lock()
				e.gameState.CurrentPhase = state.PhaseResolved
				e.gameState.Mu.Unlock()
				// CREWMATE WIN
				return
			}
		}
	} else {
		// Tie or Skip
		e.broadcastMeetingResult("No one was ejected (Tie or Skip).")
	}

	// If the game didn't end, reset everyone to Cafeteria and resume
	e.gameState.Mu.Lock()
	e.gameState.CurrentPhase = state.PhaseInPlay
	e.gameState.ActiveMeeting = nil
	e.gameState.Mu.Unlock()

	e.ResetToCafeteria()
}

// BroadcastChatUpdate sends the new chat message to all alive players.
func (e *Engine) BroadcastChatUpdate(message string) {
	// We need a helper method to get all players to iterate through them
	// Assuming you have GetAllPlayersUnsafe() or similar in gamestate.go
	players := e.gameState.GetAllPlayersUnsafe()

	payload := map[string]interface{}{
		"prompt_id": "PROMPT_09_CHAT_EVALUATION_TICK",
		"type":      "meeting_chat_update",
		"message":   message,
	}

	for _, p := range players {
		p.Mu.RLock()
		isAlive := p.IsAlive
		p.Mu.RUnlock()

		if isAlive && e.network != nil {
			e.network.SendToPlayer(p.ID, payload)
		}
	}
}

// ResetToCafeteria teleports everyone back to the start and clears cooldowns after a meeting.
func (e *Engine) ResetToCafeteria() {
	players := e.gameState.GetAllPlayersUnsafe()
	cafeteria, _ := e.gameState.GetRoom(state.RoomCafeteria)

	now := time.Now().UnixMilli()

	for _, p := range players {
		p.Mu.Lock()

		// Remove from old room attendance safely
		if oldRoom, err := e.gameState.GetRoom(p.CurrentNode); err == nil && p.CurrentNode != state.RoomCafeteria {
			oldRoom.RemoveFromAttendance(p.ID)
		}

		// Reset state
		p.CurrentNode = state.RoomCafeteria
		p.ActionStatus = state.StatusIdle
		p.MigrationUnlocksAt = now
		p.KillUnlocksAt = now
		p.TaskCompletesAt = now
		p.ChatUnlocksAt = now

		p.Mu.Unlock() // Unlock player before touching the room to prevent deadlocks

		//  Use the thread-safe State Access method instead of the raw private map!
		cafeteria.AddToAttendance(p.ID)
	}

	// Broadcast the new local map to everyone
	for _, p := range players {
		e.SendLocalMapUpdate(p.ID, state.RoomCafeteria)
	}
}

// broadcastMeetingResult sends the voting outcome to everyone.
func (e *Engine) broadcastMeetingResult(message string) {
	players := e.gameState.GetAllPlayersUnsafe()

	payload := map[string]interface{}{
		"prompt_id": "PROMPT_11_MATCH_RESOLUTION",
		"type":      "meeting_result",
		"message":   message,
	}

	for _, p := range players {
		if e.network != nil {
			e.network.SendToPlayer(p.ID, payload)
		}
	}
}

// HandleReportBody allows an agent to call a meeting if they are in the same room as a dead body.
func (e *Engine) HandleReportBody(playerID string) error {
	player, err := e.gameState.GetPlayer(playerID)
	if err != nil {
		return err
	}

	player.Mu.RLock()
	if !player.IsAlive {
		player.Mu.RUnlock()
		return fmt.Errorf("dead agents cannot report bodies")
	}
	currentRoom := player.CurrentNode
	player.Mu.RUnlock()

	// Check if there is actually a dead body in this room
	players := e.gameState.GetAllPlayersUnsafe()
	bodyFound := false

	for _, p := range players {
		p.Mu.RLock()
		isDead := !p.IsAlive
		isInSameRoom := p.CurrentNode == currentRoom
		p.Mu.RUnlock()

		if isDead && isInSameRoom {
			bodyFound = true
			break
		}
	}

	if !bodyFound {
		return fmt.Errorf("action rejected: no dead body found in %s", currentRoom)
	}

	// Body verified! Trigger the meeting phase.
	return e.TriggerMeeting(playerID, "REPORT")
}
