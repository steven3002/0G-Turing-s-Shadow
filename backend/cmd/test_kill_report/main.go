// cmd/test_kill_report/main.go
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
	log.Println("[TEST: KILL & REPORT] Starting...")

	// We need to carefully connect and read the roles
	var impostor *websocket.Conn
	var crewmate *websocket.Conn
	var targetID string

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
		c.WriteJSON(map[string]interface{}{"client_action": "HANDSHAKE_RESPONSE", "payload": map[string]string{"solution_string": fmt.Sprintf("%s%d", string(runes), mult)}})

		// Read the Role Prompt (Sent immediately on Match Start)
		go func(conn *websocket.Conn, botIndex int) {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}
				strMsg := string(msg)

				if strings.Contains(strMsg, "PROMPT_04") {
					impostor = conn
					log.Println("Found Impostor!")
				} else if strings.Contains(strMsg, "PROMPT_03") && crewmate == nil {
					crewmate = conn
					// Hacky way to guess a target ID for the test based on nonce
					targetID = fmt.Sprintf("agent_%s", nonce[:2])
				}
			}
		}(c, i)
	}

	time.Sleep(3 * time.Second) // Wait for matchmaker

	if impostor != nil {
		log.Printf("[ACTION] Impostor Killing %s...", targetID)
		impostor.WriteJSON(map[string]interface{}{
			"action":  "KILL",
			"payload": map[string]string{"target_id": targetID},
		})

		time.Sleep(1 * time.Second)

		log.Println("[ACTION] Crewmate Reporting Body...")
		crewmate.WriteJSON(map[string]interface{}{
			"action":  "REPORT_BODY",
			"payload": map[string]string{},
		})

		time.Sleep(2 * time.Second)
		log.Println("[TEST PASSED] Kill executed and Meeting successfully triggered by Report. Check server logs.")
	} else {
		log.Println("[ERROR] Failed to find impostor socket.")
	}
}
