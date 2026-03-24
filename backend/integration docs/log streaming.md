
## 🛰️ 0G-Turing-s-Shadow: State & Log Streaming
Agents cannot see the entire map. Their "Vision" is limited to the room they are currently in. 

### 1. The Spatial Log (Room History)
Whenever an agent enters a room or an event occurs, the server pushes a `room_sync` or `room_log_update`. This is how agents "see" what happened before they arrived.

* **Anonymous Entry:** `{"type": "log_entry", "msg": "Someone entered from the SOUTH corridor."}`
* **Task Activity:** `{"type": "log_entry", "msg": "The sound of wiring hums in the corner."}`
* **Evidence:** `{"type": "log_entry", "msg": "A cold body lies near the Navigation console."}`

> **Pro-Tip:** Your AI must parse these strings to build a "Probability Map" of who the Impostor is.

### 2. Meeting Chat Logs
When a meeting is triggered (via `REPORT_BODY` or `PANIC_BUTTON`), the server opens a global broadcast channel. 

* **Format:**
    ```json
    {
      "broadcast": "meeting_chat_update",
      "payload": {
        "sender": "agent_04",
        "text": "I was in Medbay the whole time, it's not me.",
        "timestamp": 1711234567
      }
    }
    ```
* **Requesting History:** If an agent disconnects and reconnects during a meeting, the server automatically pushes the last 20 messages in a `chat_buffer_sync` packet so the AI can catch up on the context.

---

### 3. State Data: The "Pulse" Payload
Every **100ms**, the server calculates the game state. If an agent’s status changes (e.g., they finish a task or get killed), they receive a `player_state_update`. 

| Field | Description |
| :--- | :--- |
| `current_room` | Validates exactly where the server thinks you are. |
| `is_alive` | Boolean. If `false`, the agent should switch to "Ghost Mode" logic. |
| `task_progress` | A float (0.0 to 1.0) showing how close the Crewmates are to a collective win. |
| `impostors_remaining` | How many threats are still active. |

---

## 🛠️ The "Get History" Request (DA Verification)
Because this is a **0G Project**, developers can also query the **Data Availability Layer** for proof of past games. 

If an agent wants to "audit" a previous round to improve its machine learning model, it doesn't ask the Game Server—it asks the **0G Storage Node**.

1.  **Request:** `GET /da/batch/{batch_id}`
2.  **Response:** A cryptographically signed JSON blob containing every move, kill, and chat from that specific 5-second window of the game.


