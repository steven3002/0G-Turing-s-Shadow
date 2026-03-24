// internal/engine/router.go
package engine

import (
	"encoding/json"
	"fmt"

	// "log"

	"github.com/steven3002/0G-Turing-s-Shadow/backend/internal/network"
)

// RouteAction parses the raw socket bytes and dispatches them to the correct handler.
func (e *Engine) RouteAction(playerID string, rawMessage []byte) error {
	var req network.ClientRequest
	if err := json.Unmarshal(rawMessage, &req); err != nil {
		return fmt.Errorf("malformed JSON payload")
	}

	switch req.Action {
	case "MOVE":
		var p network.MovePayload
		if err := json.Unmarshal(req.Payload, &p); err != nil {
			return fmt.Errorf("invalid MOVE payload schema")
		}
		return e.HandleMove(playerID, p)

	case "START_TASK":
		var p network.TaskPayload
		if err := json.Unmarshal(req.Payload, &p); err != nil {
			return fmt.Errorf("invalid START_TASK payload schema")
		}
		return e.HandleTask(playerID, p)

	case "KILL":
		var p network.KillPayload
		if err := json.Unmarshal(req.Payload, &p); err != nil {
			return fmt.Errorf("invalid KILL payload schema")
		}
		return e.HandleKill(playerID, p)

	case "PANIC_BUTTON":
		return e.HandlePanicButton(playerID)

	case "REPORT_BODY":
		return e.HandleReportBody(playerID)

	default:
		return fmt.Errorf("unknown action type: %s", req.Action)
	}
}
