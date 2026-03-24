// cmd/test_chats/main.go
package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"strings"
	"time"
)

func main() {
	log.Println("[TEST: CHAT SPAM & RATE LIMITS] Starting...")

	var bots []*websocket.Conn
	msgCount := 0

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/lobby/join"}
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
		c.WriteJSON(map[string]interface{}{"client_action": "HANDSHAKE_RESPONSE", "payload": map[string]string{"solution_string": fmt.Sprintf("%s%d", string(runes), mult)}})

		bots = append(bots, c)
		go func(conn *websocket.Conn) {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}
				if strings.Contains(string(msg), "meeting_chat_update") {
					msgCount++
				}
			}
		}(c)
	}

	time.Sleep(3 * time.Second)

	log.Println("[ACTION] Triggering Meeting...")
	bots[0].WriteJSON(map[string]interface{}{"action": "PANIC_BUTTON", "payload": map[string]string{}})
	time.Sleep(1 * time.Second)

	log.Println("[ACTION] All bots spamming 3 messages each (0.7s spacing)...")
	for loop := 0; loop < 3; loop++ {
		for i, b := range bots {
			b.WriteJSON(map[string]interface{}{
				"action":  "SEND_CHAT",
				"payload": map[string]string{"message": fmt.Sprintf("Spam message %d from bot %d", loop, i)},
			})
		}
		time.Sleep(700 * time.Millisecond) // Barely clear the 0.6s rate limit
	}

	time.Sleep(2 * time.Second)
	log.Printf("[TEST COMPLETE] Server successfully broadcasted %d/27 chat payloads.", msgCount/9)
	// (Divide by 9 because every bot receives a copy of the broadcast)
}
