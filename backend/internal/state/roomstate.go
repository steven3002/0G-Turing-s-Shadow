// internal/state/roomstate.go
package state

import "sync"

type LogEvent struct {
	EventID   string
	Action    string // ENTERED, EXITED, ASSASSINATED, TASK_COMPLETED
	ActorID   string // Can be masked for shapeshifters/tasks
	ExpiresAt int64  // Exact millisecond this must be purged
}

type TaskStatus struct {
	TaskID string
	Status string // PENDING, COMPLETED, SABOTAGED
}

type RoomState struct {
	Mu sync.RWMutex

	Name          RoomName
	AdjacentNodes []RoomName
	attendance    map[string]bool // Fast O(1) presence checks
	activityLogs  []LogEvent
	Tasks         map[string]TaskStatus
}

func NewRoomState(name RoomName, adjacent []RoomName) *RoomState {
	return &RoomState{
		Name:          name,
		AdjacentNodes: adjacent,
		attendance:    make(map[string]bool),
		activityLogs:  make([]LogEvent, 0),
		Tasks:         make(map[string]TaskStatus),
	}
}

// Thread-Safe Log Purging (Called by the 100ms Engine Tick)
func (r *RoomState) PurgeExpiredLogs(nowUnixMilli int64) {
	r.Mu.Lock()
	defer r.Mu.Unlock()

	// Filter out expired logs in-place to minimize allocations
	validLogs := r.activityLogs[:0]
	for _, log := range r.activityLogs {
		if log.ExpiresAt > nowUnixMilli {
			validLogs = append(validLogs, log)
		}
	}
	r.activityLogs = validLogs
}

// AddEvent pushes a new decaying log to the room stream.
func (r *RoomState) AddEvent(event LogEvent) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.activityLogs = append(r.activityLogs, event)
}

// GetActiveLogs returns a safe copy of the current logs for JSON serialization.
func (r *RoomState) GetActiveLogs() []LogEvent {
	r.Mu.RLock()
	defer r.Mu.RUnlock()

	// Return a copy to prevent race conditions during JSON marshalling
	snapshot := make([]LogEvent, len(r.activityLogs))
	copy(snapshot, r.activityLogs)
	return snapshot
}

// AddToAttendance marks an agent as present.
func (r *RoomState) AddToAttendance(playerID string) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	r.attendance[playerID] = true
}

// RemoveFromAttendance removes an agent.
func (r *RoomState) RemoveFromAttendance(playerID string) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	delete(r.attendance, playerID)
}

// GetAttendanceSnapshot returns a thread-safe list of active agent IDs.
func (r *RoomState) GetAttendanceSnapshot() []string {
	r.Mu.RLock()
	defer r.Mu.RUnlock()

	snapshot := make([]string, 0, len(r.attendance))
	for id := range r.attendance {
		snapshot = append(snapshot, id)
	}
	return snapshot
}
