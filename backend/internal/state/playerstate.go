// internal/state/playerstate.go
package state

import (
	"sync"
	"time"
)

type PlayerState struct {
	Mu sync.RWMutex

	ID           string
	Role         Role
	IsAlive      bool
	CurrentNode  RoomName
	ActionStatus ActionStatus

	// Cooldowns (Unix Milliseconds)
	MigrationUnlocksAt int64
	KillUnlocksAt      int64
	TaskCompletesAt    int64
	ChatUnlocksAt      int64

	// Impostor Specifics
	ShapeshiftExpiresAt int64
	ShapeshiftUnlocksAt int64
	SabotageCount1m     int
}

func NewPlayerState(id string, role Role) *PlayerState {
	return &PlayerState{
		ID:           id,
		Role:         role,
		IsAlive:      true,
		CurrentNode:  RoomCafeteria, // Universal spawn
		ActionStatus: StatusIdle,
	}
}

// Example Thread-Safe Accessor
func (p *PlayerState) CanMove() bool {
	p.Mu.RLock()
	defer p.Mu.RUnlock()

	if !p.IsAlive || p.ActionStatus != StatusIdle {
		return false
	}
	return time.Now().UnixMilli() >= p.MigrationUnlocksAt
}

// UpdateLocation securely teleports the agent.
// Note: This does NOT update the Room attendance logs; the Engine layer handles that orchestration.
func (p *PlayerState) UpdateLocation(newRoom RoomName) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.CurrentNode = newRoom
}

// UpdateStatus securely changes the agent's action state (e.g., IDLE -> DOING_TASK).
func (p *PlayerState) UpdateStatus(status ActionStatus) {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.ActionStatus = status
}

// ApplyCooldown sets a specific timestamp lock securely.
func (p *PlayerState) ApplyCooldown(cooldownType string, unlockTimeUnixMilli int64) {
	p.Mu.Lock()
	defer p.Mu.Unlock()

	switch cooldownType {
	case "MIGRATION":
		p.MigrationUnlocksAt = unlockTimeUnixMilli
	case "KILL":
		p.KillUnlocksAt = unlockTimeUnixMilli
	case "TASK":
		p.TaskCompletesAt = unlockTimeUnixMilli
	case "CHAT":
		p.ChatUnlocksAt = unlockTimeUnixMilli
	}
}

// Kill securely flags the agent as dead and resets their status.
func (p *PlayerState) Kill() {
	p.Mu.Lock()
	defer p.Mu.Unlock()
	p.IsAlive = false
	p.ActionStatus = StatusDead
}
