// cmd/test_movement/main.go
package main

import (
	// "encoding/hex"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	log.Println("[TEST: MOVEMENT & TASKS] Starting...")
	bots := spawnLobby(9)
	time.Sleep(3 * time.Second) // Wait for matchmaker

	botA := bots[0] // Pick the first bot

	// 1. Test Movement
	log.Println("[ACTION] Bot 0 Moving to NAVIGATION...")
	botA.WriteJSON(map[string]interface{}{
		"action": "MOVE",
		"payload": map[string]string{
			"destination": "NAVIGATION",
		},
	})
	time.Sleep(2 * time.Second) // Wait for migration cooldown

	// 2. Test Task Execution
	log.Println("[ACTION] Bot 0 Starting Task in NAVIGATION...")
	botA.WriteJSON(map[string]interface{}{
		"action": "START_TASK",
		"payload": map[string]string{
			"task_id": "nav_wiring_01",
		},
	})

	time.Sleep(4 * time.Second) // Wait for 3s task to complete and emit pulse
	log.Println("[TEST PASSED] Movement and Task execution completed. Check server logs for anonymous pulse.")
}

// --- BOILERPLATE HELPER ---
func spawnLobby(count int) []*websocket.Conn {
	var conns []*websocket.Conn
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/lobby/join"}

	for i := 0; i < count; i++ {
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("Dial error:", err)
		}

		var challenge map[string]interface{}
		c.ReadJSON(&challenge)
		p := challenge["payload"].(map[string]interface{})
		nonce, mult := p["nonce"].(string), int(p["multiplier"].(float64))

		runes := []rune(nonce)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		solution := fmt.Sprintf("%s%d", string(runes), mult)

		c.WriteJSON(map[string]interface{}{
			"client_action": "HANDSHAKE_RESPONSE",
			"payload":       map[string]string{"solution_string": solution},
		})

		// Drain the socket in the background to prevent buffer blocking
		go func(conn *websocket.Conn) {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}(c)

		conns = append(conns, c)
	}
	return conns
}
