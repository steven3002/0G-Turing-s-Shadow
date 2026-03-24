package main

import (
	"fmt"
	"log"
	"net/url"
	"time"
	"math/rand"

	"github.com/gorilla/websocket"
)

func main() {
	count := 8
	room := "alpha_squad"
	log.Printf("[GAME AUTOMATOR] Spawning %d bots for room: %s", count, room)
	
	bots := make([]*websocket.Conn, count)
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/lobby/join", RawQuery: "room=" + room}

	for i := 0; i < count; i++ {
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("Dial error:", err)
		}

		// 1. Handshake
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
		bots[i] = c
	}

	log.Println("[SUCCESS] 8 bots ready. Join as the 9th player to start!")

	// 2. Wait for match to start
	// We'll listen for the first sensory injection on bot 0
	var msg map[string]interface{}
	for {
		if err := bots[0].ReadJSON(&msg); err == nil {
			if msg["prompt_id"] == "PROMPT_03_INIT_CREWMATE" || msg["prompt_id"] == "PROMPT_04_INIT_IMPOSTOR" {
				log.Println("[MATCH STARTED] Bots entering strategic loop...")
				break
			}
		}
	}

	// 3. Automated Strategic Loop (Aggressive Movement)
	rooms := []string{"NAVIGATION", "STORAGE", "ELECTRICAL", "NUCLEAR_REACTOR", "MEDBAY", "CAFETERIA"}
	
	go func() {
		for {
			time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second) // Faster movement
			targetBot := rand.Intn(count)
			targetRoom := rooms[rand.Intn(len(rooms))]
			
			log.Printf("[BOT %d] Moving to %s", targetBot, targetRoom)
			bots[targetBot].WriteJSON(map[string]interface{}{
				"action": "MOVE",
				"payload": map[string]string{"destination": targetRoom},
			})

			// 3b. Random Kill Action (If bot 0 is impostor, or just simulation)
			if rand.Intn(10) == 0 {
				victim := rand.Intn(count)
				if victim != targetBot {
					log.Printf("[ACTION] Bot %d attempting KILL on Bot %d", targetBot, victim)
					bots[targetBot].WriteJSON(map[string]interface{}{
						"action": "KILL",
						"payload": map[string]string{"target_id": fmt.Sprintf("Bot_%d", victim)},
					})
				}
			}

			// 3c. Random Panic Button
			if rand.Intn(50) == 0 && targetRoom == "CAFETERIA" {
				log.Printf("[ACTION] Bot %d pressing PANIC BUTTON", targetBot)
				bots[targetBot].WriteJSON(map[string]interface{}{
					"action": "PANIC_BUTTON",
					"payload": map[string]string{},
				})
			}
		}
	}()

	// 4. Automated Task Loop
	go func() {
		for {
			time.Sleep(time.Duration(10+rand.Intn(10)) * time.Second)
			targetBot := rand.Intn(count)
			log.Printf("[BOT %d] Starting Task", targetBot)
			bots[targetBot].WriteJSON(map[string]interface{}{
				"action": "START_TASK",
				"payload": map[string]string{"task_id": "generic_task"},
			})
		}
	}()

	// 5. Automated Meeting Interaction (Listening for meeting start)
	for i := 0; i < count; i++ {
		go func(id int, conn *websocket.Conn) {
			for {
				var innerMsg map[string]interface{}
				if err := conn.ReadJSON(&innerMsg); err != nil {
					return
				}
				
				if innerMsg["prompt_id"] == "PROMPT_08_MEETING_START_CONTEXT" {
					log.Printf("[BOT %d] Meeting started. Preparing chat...", id)
					time.Sleep(2 * time.Second)
					conn.WriteJSON(map[string]interface{}{
						"action": "SEND_CHAT",
						"payload": map[string]string{"message": fmt.Sprintf("I am bot %d, I was in a room doing things.", id)},
					})
				}

				if innerMsg["prompt_id"] == "PROMPT_10_VOTING_DEMAND" {
					log.Printf("[BOT %d] Voting demand received. Skipping...", id)
					conn.WriteJSON(map[string]interface{}{
						"action": "VOTE",
						"payload": map[string]string{"target_id": "SKIP"},
					})
				}
			}
		}(i, bots[i])
	}

	select {} // Keep running
}
