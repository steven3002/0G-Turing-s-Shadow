// cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/engine"
	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/network"
	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/state"
)

// MatchCoordinator manages multiple concurrent game instances
type MatchCoordinator struct {
	mu      sync.RWMutex
	matches map[string]*network.Manager
}

func NewMatchCoordinator() *MatchCoordinator {
	return &MatchCoordinator{
		matches: make(map[string]*network.Manager),
	}
}

// GetOrCreateMatch checks if a room exists. If not, it spins up a new isolated game engine.
func (c *MatchCoordinator) GetOrCreateMatch(matchID string) *network.Manager {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If the match is already running, return its specific network manager
	if manager, exists := c.matches[matchID]; exists {
		return manager
	}

	log.Printf("[COORDINATOR] Booting new isolated match sandbox: %s", matchID)

	// 1. Initialize Thread-Safe State for this specific match
	gameState := state.NewGameStateManager(matchID)

	// 2. Initialize the Game Engine for this specific match
	gameEngine := engine.NewEngine(gameState)

	// 3. Initialize the WebSocket Network Manager for this specific match
	wsManager := network.NewManager(gameState)

	// 4. Wire the interfaces together
	wsManager.SetEngine(gameEngine)
	gameEngine.SetNetwork(wsManager)

	// 5. Start Background Workers independently
	go wsManager.Start()
	gameEngine.Start() // Non-blocking, starts the 100ms ticker

	// Save to active matches
	c.matches[matchID] = wsManager
	return wsManager
}

// HandleJoin intercepts the HTTP request, parses the room code, and routes the socket
func (c *MatchCoordinator) HandleJoin(w http.ResponseWriter, r *http.Request) {
	matchID := r.URL.Query().Get("room")
	if matchID == "" {
		http.Error(w, "Missing 'room' query parameter", http.StatusBadRequest)
		return
	}

	// Fetch the specific room's network manager and pass the socket upgrade to it
	manager := c.GetOrCreateMatch(matchID)
	manager.HandleConnection(w, r)
}

func main() {
	log.Println("Booting 0G-Turing-s-Shadow Multi-State Backend...")

	// Initialize the global lobby coordinator
	coordinator := NewMatchCoordinator()

	mux := http.NewServeMux()
	mux.HandleFunc("/lobby/join", coordinator.HandleJoin)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server successfully listening on :8080. Awaiting room connections...")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}
