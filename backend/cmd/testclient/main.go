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
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/lobby/join"}
	log.Printf("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer c.Close()

	// 1. Wait for the Challenge (Must reply < 150ms)
	var challenge map[string]interface{}
	err = c.ReadJSON(&challenge)
	if err != nil {
		log.Fatal("Failed to read challenge:", err)
	}

	payload := challenge["payload"].(map[string]interface{})
	nonce := payload["nonce"].(string)
	multiplier := int(payload["multiplier"].(float64))

	// 2. Solve it instantly
	solution := fmt.Sprintf("%s%d", reverseString(nonce), multiplier)

	response := map[string]interface{}{
		"client_action": "HANDSHAKE_RESPONSE",
		"payload": map[string]string{
			"solution_string": solution,
		},
	}

	err = c.WriteJSON(response)
	if err != nil {
		log.Fatal("Failed to send solution:", err)
	}

	log.Println("Handshake solved and sent. Awaiting game start...")

	// 3. Listen to the Game Event Stream
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("Connection closed:", err)
			return
		}
		log.Printf("\n[SERVER DATA]: %s", message)
		time.Sleep(10 * time.Millisecond) // Prevent console flooding
	}
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
