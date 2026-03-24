// internal/state/gamestate.go
package state

import (
	"errors"
	"fmt"
	"sync"
)

type MeetingState struct {
	CallerID        string
	MeetingType     string // "PANIC" or "REPORT"
	DeadDiscovered  []string
	ChatTranscript  []string
	BallotBox       map[string]string // map[voterID]targetID
	PhaseEndsAtUnix int64
}

type GameStateManager struct {
	Mu sync.RWMutex

	MatchID             string
	CurrentPhase        Phase
	TotalCrewmatesAlive int
	TotalImpostorsAlive int

	// Pointers to the granular mutex objects
	players map[string]*PlayerState
	rooms   map[RoomName]*RoomState

	// Add this to GameStateManager struct:
	ActiveMeeting *MeetingState
}

func NewGameStateManager(matchID string) *GameStateManager {
	manager := &GameStateManager{
		MatchID:      matchID,
		CurrentPhase: PhaseLobby,
		players:      make(map[string]*PlayerState),
		rooms:        make(map[RoomName]*RoomState),
	}

	// Bootstrap the Map Topology
	manager.rooms[RoomCafeteria] = NewRoomState(RoomCafeteria, []RoomName{RoomNavigation, RoomStorage})
	manager.rooms[RoomNavigation] = NewRoomState(RoomNavigation, []RoomName{RoomElectrical, RoomNuclearReactor, RoomCafeteria})
	manager.rooms[RoomStorage] = NewRoomState(RoomStorage, []RoomName{RoomNuclearReactor, RoomMedbay, RoomCafeteria})
	manager.rooms[RoomElectrical] = NewRoomState(RoomElectrical, []RoomName{RoomNavigation})
	manager.rooms[RoomNuclearReactor] = NewRoomState(RoomNuclearReactor, []RoomName{RoomNavigation, RoomStorage})
	manager.rooms[RoomMedbay] = NewRoomState(RoomMedbay, []RoomName{RoomStorage})

	return manager
}

// AddPlayer registers a new AI connection securely.
func (m *GameStateManager) AddPlayer(player *PlayerState) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if _, exists := m.players[player.ID]; exists {
		return fmt.Errorf("player %s already exists in the lobby", player.ID)
	}
	m.players[player.ID] = player
	return nil
}

// GetPlayer securely fetches a pointer to the agent's state.
// The caller MUST use the PlayerState's internal mutex to mutate it.
func (m *GameStateManager) GetPlayer(id string) (*PlayerState, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	player, exists := m.players[id]
	if !exists {
		return nil, errors.New("player not found")
	}
	return player, nil
}

// GetRoom securely fetches a room for localized operations.
func (m *GameStateManager) GetRoom(name RoomName) (*RoomState, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	room, exists := m.rooms[name]
	if !exists {
		return nil, fmt.Errorf("invalid room topology: %s", name)
	}
	return room, nil
}

// GetAllRooms returns a slice of pointers to all room states.
// This allows the Engine loop to iterate over the rooms without holding
// the global GameStateManager mutex for the entire tick.
func (m *GameStateManager) GetAllRooms() []*RoomState {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	rooms := make([]*RoomState, 0, len(m.rooms))
	for _, room := range m.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// GetAllPlayersUnsafe returns a slice of pointers to all currently registered players.
// It is suffixed with "Unsafe" as a reminder to the caller: while the slice itself
// is safely generated without race conditions, the individual PlayerState objects
// inside the slice MUST be individually locked (e.g., p.Mu.RLock()) before reading
// or writing their properties.
func (m *GameStateManager) GetAllPlayersUnsafe() []*PlayerState {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	// Pre-allocate the slice capacity to prevent unnecessary memory allocations during the loop
	players := make([]*PlayerState, 0, len(m.players))

	for _, player := range m.players {
		players = append(players, player)
	}

	return players
}
