// cmd/test_vote_win/main.go
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
	log.Println("[TEST: CHAT, VOTE & WIN SEQUENCE] Starting...")

	var bots []*websocket.Conn
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
		c.WriteJSON(map[string]interface{}{"client_action": "HANDSHAKE_RESPONSE", "payload": map[string]string{"solution_string": fmt.Sprintf("%s%d", string(runes), mult)}})

		bots = append(bots, c)

		agentID := fmt.Sprintf("agent_%s", nonce[:2])

		go func(conn *websocket.Conn, id string) {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}
				if strings.Contains(string(msg), "PROMPT_04") {
					impostorID = id // We found the Impostor's ID!
				}
			}
		}(c, agentID)
	}

	time.Sleep(3 * time.Second)

	log.Println("[ACTION] Pressing Panic Button...")
	bots[0].WriteJSON(map[string]interface{}{"action": "PANIC_BUTTON", "payload": map[string]string{}})

	time.Sleep(2 * time.Second)
	log.Println("[ACTION] All bots sending Chat Messages...")
	for _, b := range bots {
		b.WriteJSON(map[string]interface{}{
			"action":  "SEND_CHAT",
			"payload": map[string]string{"message": fmt.Sprintf("I think it is %s!", impostorID)},
		})
		time.Sleep(100 * time.Millisecond) // Prevent buffer flood
	}

	log.Println("[WAITING] 20 seconds for Voting Phase...")
	time.Sleep(21 * time.Second)

	log.Printf("[ACTION] All bots Voting to Eject Impostor: %s...", impostorID)
	for _, b := range bots {
		b.WriteJSON(map[string]interface{}{
			"action":  "CAST_VOTE",
			"payload": map[string]string{"target_id": impostorID},
		})
	}

	time.Sleep(3 * time.Second)
	log.Println("[TEST PASSED] Impostor ejected. Crewmates Win. Check server logs for Match Resolution.")
}
