// internal/network/manager.go
package network

import (
	"encoding/json"
	"log"
	"net/http"
	"sync" // FIX: Added sync package

	"github.com/gorilla/websocket"
	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/state"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type GameEngine interface {
	RouteAction(playerID string, rawMessage []byte) error
	CheckLobbyReadiness()
}

type Manager struct {
	mu         sync.RWMutex // FIX: Protects the clients map
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	gameState  *state.GameStateManager
	engine     GameEngine
}

func NewManager(gs *state.GameStateManager) *Manager {
	return &Manager{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		gameState:  gs,
	}
}

func (m *Manager) SetEngine(e GameEngine) {
	m.engine = e
}

func (m *Manager) Start() {
	for {
		select {
		case client := <-m.register:
			m.mu.Lock() // FIX: Lock before writing
			m.clients[client] = true
			m.mu.Unlock() // FIX: Unlock after writing

			log.Printf("New connection registered. Total active: %d", len(m.clients))

			if m.engine != nil {
				m.engine.CheckLobbyReadiness()
			}

		case client := <-m.unregister:
			m.mu.Lock() // FIX: Lock before checking/writing
			if _, ok := m.clients[client]; ok {
				delete(m.clients, client)
				close(client.send)
				log.Printf("Connection unregistered. Total active: %d", len(m.clients))
			}
			m.mu.Unlock() // FIX: Unlock after
		}
	}
}

func (m *Manager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	playerID, err := PerformHandshake(conn)
	if err != nil {
		log.Printf("Handshake rejected for %s: %v", conn.RemoteAddr(), err)
		conn.WriteJSON(map[string]string{"error": "HANDSHAKE_FAILED", "reason": err.Error()})
		conn.Close()
		return
	}

	log.Printf("Handshake successful. Agent authenticated as: %s", playerID)

	newPlayer := &state.PlayerState{
		ID:           playerID,
		IsAlive:      true,
		ActionStatus: state.StatusIdle,
	}

	if err := m.gameState.AddPlayer(newPlayer); err != nil {
		log.Printf("Failed to add player to state: %v", err)
		conn.Close()
		return
	}

	client := &Client{
		manager:  m,
		conn:     conn,
		send:     make(chan []byte, 256),
		playerID: playerID,
	}

	m.register <- client

	go client.writePump()
	go client.readPump()
}

// SendToPlayer safely writes a JSON payload to a specific agent's socket.
func (m *Manager) SendToPlayer(playerID string, payload interface{}) {
	m.mu.RLock() // FIX: Read-lock before iterating over the map
	var targetClient *Client
	for client := range m.clients {
		if client.playerID == playerID {
			targetClient = client
			break
		}
	}
	m.mu.RUnlock() // FIX: Unlock after finding the target

	if targetClient == nil {
		return
	}

	jsonBytes, err := json.Marshal(payload)
	if err == nil {
		targetClient.send <- jsonBytes
	}
}
