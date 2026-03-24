// cmd/testclient/main.go
package main

import (
	// "encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/lobby/join", RawQuery: "room=alpha_squad"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer c.Close()

	// 1. Defeat the Gatekeeper Handshake
	var challenge map[string]interface{}
	c.ReadJSON(&challenge)
	payload := challenge["payload"].(map[string]interface{})
	nonce := payload["nonce"].(string)
	multiplier := int(payload["multiplier"].(float64))

	solution := fmt.Sprintf("%s%d", reverseString(nonce), multiplier)
	c.WriteJSON(map[string]interface{}{
		"client_action": "HANDSHAKE_RESPONSE",
		"payload": map[string]string{
			"solution_string": solution,
		},
	})

	// 2. Start a background reader to print server events
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				return
			}
			// Only print the meeting and chat updates so the console isn't flooded
			msgStr := string(message)
			if contains(msgStr, "PROMPT_08") || contains(msgStr, "PROMPT_09") || contains(msgStr, "PROMPT_10") || contains(msgStr, "PROMPT_11") {
				log.Printf("\n[SERVER]: %s", msgStr)
			}
		}
	}()

	// ==========================================
	// THE CHOREOGRAPHED E2E TEST SEQUENCE
	// ==========================================

	// Wait for the Matchmaker to gather 9 players and start the game
	time.Sleep(3 * time.Second)
	log.Println("[BOT] Match started. Attempting to press PANIC BUTTON...")

	// Action 1: Trigger a Meeting (Everyone is in Cafeteria, so this is valid)
	c.WriteJSON(map[string]interface{}{
		"action":  "PANIC_BUTTON",
		"payload": map[string]string{},
	})

	// Action 2: Chatting
	time.Sleep(2 * time.Second) // Let the phase transition to MeetingChat
	log.Println("[BOT] Chat Phase Active. Sending NLP message...")
	c.WriteJSON(map[string]interface{}{
		"action": "SEND_CHAT",
		"payload": map[string]string{
			"message": "Who pressed the button? I was just standing here!",
		},
	})

	// Action 3: Voting
	// The MeetingChat phase is strictly 20 seconds. We must wait for it to end.
	log.Println("[BOT] Waiting 20 seconds for Chat Phase to end...")
	time.Sleep(21 * time.Second)

	log.Println("[BOT] Voting Phase Active. Casting vote to SKIP...")
	c.WriteJSON(map[string]interface{}{
		"action": "CAST_VOTE",
		"payload": map[string]string{
			"target_id": "SKIP",
		},
	})

	// Wait to receive the Match Resolution broadcast
	time.Sleep(3 * time.Second)
	log.Println("[BOT] Test Sequence Complete. Disconnecting.")
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > 0 && s[0] == substr[0] // Simple hacky contains for logs
}
