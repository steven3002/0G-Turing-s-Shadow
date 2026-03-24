// internal/network/client.go
package network

import (
	// "bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 2048
)

// Client acts as the intermediary between the websocket connection and the game hub.
type Client struct {
	manager  *Manager
	conn     *websocket.Conn
	send     chan []byte
	playerID string // Assigned after successful handshake
}

// readPump pumps messages from the websocket connection to the server.
func (c *Client) readPump() {
	defer func() {
		c.manager.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		// Dispatch to the Engine Router
		err = c.manager.engine.RouteAction(c.playerID, message)

		if err != nil {
			// Provide strict negative feedback to the AI model
			response := ServerResponse{
				ResponseType: "ACTION_REJECTED",
				Status:       "ERROR",
				Message:      err.Error(),
			}
			jsonResponse, _ := json.Marshal(response)
			c.send <- jsonResponse
			log.Printf("Rejected action from %s: %v", c.playerID, err)
		}
	}
}

// writePump pumps messages from the server to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The manager closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// REMOVED: Batching of messages to prevent JSON parsing errors in the frontend.
			// Each send should be its own frame for cleaner state tracking.

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
