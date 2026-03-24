// internal/state/types.go
package state

type Phase string
type Role string
type RoomName string
type ActionStatus string

const (
	PhaseLobby       Phase = "LOBBY"
	PhaseInPlay      Phase = "IN_PLAY"
	PhaseMeetingChat Phase = "MEETING_CHAT"
	PhaseMeetingVote Phase = "MEETING_VOTE"
	PhaseResolved    Phase = "RESOLVED"

	RoleCrewmate Role = "CREWMATE"
	RoleImpostor Role = "IMPOSTOR"

	RoomCafeteria      RoomName = "CAFETERIA"
	RoomNavigation     RoomName = "NAVIGATION"
	RoomStorage        RoomName = "STORAGE"
	RoomElectrical     RoomName = "ELECTRICAL"
	RoomNuclearReactor RoomName = "NUCLEAR_REACTOR"
	RoomMedbay         RoomName = "MEDBAY"

	StatusIdle      ActionStatus = "IDLE"
	StatusDoingTask ActionStatus = "DOING_TASK"
	StatusInMeeting ActionStatus = "IN_MEETING"
	StatusDead      ActionStatus = "DEAD"
)
