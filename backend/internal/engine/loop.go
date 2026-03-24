// internal/engine/loop.go
package engine

import (
	"fmt"
	"log"
	"time"

	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/state"
)

type NetworkDispatcher interface {
	SendToPlayer(playerID string, payload interface{})
}

type Engine struct {
	gameState *state.GameStateManager
	ticker    *time.Ticker
	quit      chan struct{}
	network   NetworkDispatcher
}

func NewEngine(gs *state.GameStateManager) *Engine {
	return &Engine{
		gameState: gs,
		ticker:    time.NewTicker(100 * time.Millisecond),
		quit:      make(chan struct{}),
	}
}

func (e *Engine) SetNetwork(n NetworkDispatcher) {
	e.network = n
}

func (e *Engine) Start() {
	log.Println("Starting Global Event Loop (Tick Rate: 100ms)...")
	go func() {
		for {
			select {
			case <-e.ticker.C:
				e.tick()
			case <-e.quit:
				e.ticker.Stop()
				log.Println("Global Event Loop stopped.")
				return
			}
		}
	}()
}

func (e *Engine) Stop() {
	close(e.quit)
}

// tick is the core logic executed 10 times per second.
func (e *Engine) tick() {
	now := time.Now().UnixMilli()

	e.gameState.Mu.RLock()
	phase := e.gameState.CurrentPhase
	meeting := e.gameState.ActiveMeeting
	e.gameState.Mu.RUnlock()

	if phase == state.PhaseInPlay {
		// 1. Purge decaying spatial logs
		rooms := e.gameState.GetAllRooms()
		for _, room := range rooms {
			room.PurgeExpiredLogs(now)
		}

		// 2. Free agents who finished tasks and emit the anonymous pulse
		e.processTaskCompletions(now)

	} else if phase == state.PhaseMeetingChat {
		// Check if the 20s Chat phase is over
		if meeting != nil && now >= meeting.PhaseEndsAtUnix {
			e.gameState.Mu.Lock()
			e.gameState.CurrentPhase = state.PhaseMeetingVote
			e.gameState.ActiveMeeting.PhaseEndsAtUnix = now + 10000 // Add 10s for voting
			e.gameState.Mu.Unlock()

			// Broadcast PROMPT_10_VOTING_DEMAND
			e.broadcastVotingDemand()
		}
	} else if phase == state.PhaseMeetingVote {
		// Check if the 10s Voting phase is over
		if meeting != nil && now >= meeting.PhaseEndsAtUnix {
			e.ResolveMeeting()
		}
	}
}

// processTaskCompletions checks for agents whose task timers have expired.
func (e *Engine) processTaskCompletions(now int64) {
	players := e.gameState.GetAllPlayersUnsafe()

	for _, p := range players {
		p.Mu.Lock()
		if p.ActionStatus == state.StatusDoingTask && now >= p.TaskCompletesAt {
			// Task is complete! Free the agent.
			p.ActionStatus = state.StatusIdle
			roomName := p.CurrentNode
			p.Mu.Unlock()

			// Inject the 1.2s Anonymous Pulse into the room
			room, err := e.gameState.GetRoom(roomName)
			if err == nil {
				room.AddEvent(state.LogEvent{
					EventID:   fmt.Sprintf("evt_%d", time.Now().UnixNano()),
					Action:    "TASK_COMPLETED",
					ActorID:   "ANONYMOUS",                   // Strict anonymity
					ExpiresAt: time.Now().UnixMilli() + 1200, // Hyper-extreme 1.2s decay
				})
				// Broadcast the update so everyone in the room (who isn't blind) sees the pulse
				e.BroadcastRoomState(roomName)
			}
		} else {
			p.Mu.Unlock()
		}
	}
}

// broadcastVotingDemand pushes PROMPT_10 to all alive agents to force a vote.
func (e *Engine) broadcastVotingDemand() {
	players := e.gameState.GetAllPlayersUnsafe()

	payload := map[string]interface{}{
		"prompt_id": "PROMPT_10_VOTING_DEMAND",
		"type":      "action_request",
		"message":   "The chat phase has ended. You have 10 seconds to cast your vote or SKIP.",
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
