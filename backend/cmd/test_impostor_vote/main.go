// cmd/test_impostor_vote/main.go
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
	log.Println("[TEST: VOTE OUT INNOCENTS] Starting...")

	var bots []*websocket.Conn
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

		bots = append(bots, c)
		go func(conn *websocket.Conn, id string) {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}
				if strings.Contains(string(msg), "PROMPT_04") {
					impostor = conn
				}
				if strings.Contains(string(msg), "PROMPT_03") {
					crewmates = append(crewmates, id)
				}
				if strings.Contains(string(msg), "IMPOSTOR WINS") {
					log.Println("\n[SUCCESS] Innocent ejected. IMPOSTOR WIN triggered!")
				}
			}
		}(c, agentID)
	}

	time.Sleep(3 * time.Second)

	// 1. Rapidly thin the herd to save time (Kill 6 people)
	log.Println("[SETUP] Impostor assassinating 6 crewmates...")
	for i := 0; i < 6; i++ {
		impostor.WriteJSON(map[string]interface{}{"action": "KILL", "payload": map[string]string{"target_id": crewmates[i]}})
		time.Sleep(5200 * time.Millisecond)
	}

	// 2. Call a meeting with the remaining 3 players
	log.Println("[ACTION] Calling Meeting...")
	bots[0].WriteJSON(map[string]interface{}{"action": "PANIC_BUTTON", "payload": map[string]string{}})

	time.Sleep(21 * time.Second) // Wait out the chat phase

	// 3. Impostor and 1 confused Crewmate vote out the final innocent Crewmate
	targetToFrame := crewmates[7]
	log.Printf("[ACTION] Voting to eject innocent: %s", targetToFrame)
	for _, b := range bots {
		b.WriteJSON(map[string]interface{}{"action": "CAST_VOTE", "payload": map[string]string{"target_id": targetToFrame}})
	}

	time.Sleep(5 * time.Second)
	log.Println("[TEST COMPLETE]")
}
