// cmd/test_sabotage/main.go
package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	log.Println("[TEST: SABOTAGE] Starting...")

	var impostor *websocket.Conn
	var impostorID string

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/lobby/join", RawQuery: "room=alpha_squad"}
	for i := 0; i < 9; i++ {
		c, _, _ := websocket.DefaultDialer.Dial(u.String(), nil)

		// Handshake
		var ch map[string]interface{}
		c.ReadJSON(&ch)
		p := ch["payload"].(map[string]interface{})
		nonce, mult := p["nonce"].(string), int(p["multiplier"].(float64))
		runes := []rune(nonce)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}

		agentID := fmt.Sprintf("agent_%s", nonce[:2])
		c.WriteJSON(map[string]interface{}{"client_action": "HANDSHAKE_RESPONSE", "payload": map[string]string{"solution_string": fmt.Sprintf("%s%d", string(runes), mult)}})

		go func(conn *websocket.Conn, id string) {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}

				// Print sensory payloads so we can see the sabotage alarm appear
				if strings.Contains(string(msg), "CRITICAL_SABOTAGE_ALARM") {
					log.Printf("\n[SERVER ALARM DETECTED]: %s", string(msg))
				}
				if strings.Contains(string(msg), "PROMPT_04") {
					impostor = conn
					impostorID = id
					log.Printf("Found Impostor: %s", id)
				}
			}
		}(c, agentID)
	}

	time.Sleep(3 * time.Second) // Wait for matchmaker

	if impostor != nil {
		log.Printf("[ACTION] Impostor %s sabotaging ELECTRICAL...", impostorID)
		impostor.WriteJSON(map[string]interface{}{
			"action":  "SABOTAGE",
			"payload": map[string]string{"target_room": "ELECTRICAL"},
		})

		time.Sleep(2 * time.Second)
		log.Println("[TEST PASSED] Sabotage executed. Server logs should show the alarm.")
	} else {
		log.Println("[ERROR] Failed to find impostor socket.")
	}
}
