// cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/engine"
	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/network"
	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/state"
)

func main() {
	log.Println("Booting 0G-Turing-s-Shadow Backend...")

	// 1. Initialize Thread-Safe State
	gameState := state.NewGameStateManager("match_local_001")

	// 2. Initialize the Game Engine
	gameEngine := engine.NewEngine(gameState)

	// 3. Initialize the WebSocket Network Manager
	wsManager := network.NewManager(gameState) // Only takes gameState n

	// 4. Wire the interfaces together safely!
	wsManager.SetEngine(gameEngine)
	gameEngine.SetNetwork(wsManager)

	// 5. Start Background Workers
	go wsManager.Start()
	gameEngine.Start() // Starts the 100ms tick loop

	// 6. Configure HTTP Router
	mux := http.NewServeMux()
	mux.HandleFunc("/lobby/join", wsManager.HandleConnection)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server successfully listening on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}
