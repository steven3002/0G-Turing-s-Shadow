// internal/network/handshake.go
package network

import (
	"crypto/rand"
	"encoding/hex"
	// "encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/gorilla/websocket"
)

const HandshakeTimeout = 150 * time.Millisecond

// ChallengePayload matches the PRD schema for the server challenge.
type ChallengePayload struct {
	SystemEvent string `json:"system_event"`
	Payload     struct {
		Nonce       string `json:"nonce"`
		Multiplier  int    `json:"multiplier"`
		Instruction string `json:"instruction"`
	} `json:"payload"`
	DeadlineMs int `json:"deadline_ms"`
}

// ResponsePayload matches the expected AI response schema.
type ResponsePayload struct {
	ClientAction string `json:"client_action"`
	Payload      struct {
		SolutionString string `json:"solution_string"`
	} `json:"payload"`
}

// PerformHandshake locks the connection into a 150ms window.
// If successful, it returns a generated PlayerID.
func PerformHandshake(conn *websocket.Conn) (string, error) {
	// 1. Generate the dynamic challenge
	nonceBytes := make([]byte, 4) // 8 hex characters
	rand.Read(nonceBytes)
	nonce := hex.EncodeToString(nonceBytes)

	multiplierBig, _ := rand.Int(rand.Reader, big.NewInt(90))
	multiplier := int(multiplierBig.Int64()) + 10 // Range 10-99

	// Calculate the expected answer: reverse(nonce) + multiplier
	expectedSolution := fmt.Sprintf("%s%d", reverseString(nonce), multiplier)

	challenge := ChallengePayload{
		SystemEvent: "HANDSHAKE_CHALLENGE",
		DeadlineMs:  150,
	}
	challenge.Payload.Nonce = nonce
	challenge.Payload.Multiplier = multiplier
	challenge.Payload.Instruction = "Reverse the nonce, append the multiplier, and return as 'solution_string'."

	// 2. Dispatch the challenge
	if err := conn.WriteJSON(challenge); err != nil {
		return "", fmt.Errorf("failed to write challenge: %w", err)
	}

	// 3. Enforce the strict 150ms physical socket deadline
	conn.SetReadDeadline(time.Now().Add(HandshakeTimeout))

	// 4. Await the response
	var response ResponsePayload
	if err := conn.ReadJSON(&response); err != nil {
		// If the error is a timeout, this drops human players.
		return "", fmt.Errorf("handshake failed/timed out: %w", err)
	}

	// 5. Clear the deadline so the socket can operate normally post-handshake
	conn.SetReadDeadline(time.Time{})

	// 6. Validate strict accuracy
	if response.ClientAction != "HANDSHAKE_RESPONSE" {
		return "", errors.New("invalid client_action field")
	}
	if response.Payload.SolutionString != expectedSolution {
		return "", errors.New("cryptographic solution mismatch")
	}

	// Handshake passed. Generate a Player ID for this session.
	// In production, this might map to a 0G wallet address or public key.
	playerID := fmt.Sprintf("agent_%s", hex.EncodeToString(nonceBytes[:2]))
	return playerID, nil
}

// Utility function to reverse the hex string
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
