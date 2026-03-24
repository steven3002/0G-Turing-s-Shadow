// cmd/test_kill_all/main.go
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
	log.Println("[TEST: KILL ALL] Starting...")

	var impostor *websocket.Conn
	var crewmates []string

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/lobby/join", RawQuery: "room=alpha_squad"}
	for i := 0; i < 9; i++ {
		c, _, _ := websocket.DefaultDialer.Dial(u.String(), nil)

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

				if strings.Contains(string(msg), "PROMPT_04") {
					impostor = conn
				} else if strings.Contains(string(msg), "PROMPT_03") {
					crewmates = append(crewmates, id)
				}
				if strings.Contains(string(msg), "IMPOSTOR WINS") {
					log.Println("\n[SUCCESS] Server broadcasted IMPOSTOR WIN condition!")
				}
			}
		}(c, agentID)
	}

	time.Sleep(3 * time.Second) // Wait for match start

	// Execute 7 Crewmates. (8 total, 7 kills leaves 1 Crewmate and 1 Impostor = Impostor Win)
	for i := 0; i < 7; i++ {
		target := crewmates[i]
		log.Printf("[ACTION] Killing %s...", target)
		impostor.WriteJSON(map[string]interface{}{
			"action":  "KILL",
			"payload": map[string]string{"target_id": target},
		})
		time.Sleep(5200 * time.Millisecond) // 5.2s wait to clear the 5.0s kill cooldown
	}

	time.Sleep(2 * time.Second)
	log.Println("[TEST COMPLETE] Check logs for Win Condition.")
}
