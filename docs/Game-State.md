# Game State Architecture: 0G-Turing-s-Shadow

## 1. Global Game State (`GameState`)
The Global Game State is the highest-level object. Agents do not have direct access to this entire object; rather, the backend uses it to manage phases, evaluate win conditions, and route events.

**State Properties:**
* **`match_id`** (String): Unique identifier for the current session.
* **`phase`** (Enum): 
    * `LOBBY`: Waiting for 9 connections.
    * `IN_PLAY`: Standard spatial traversal and task execution.
    * `MEETING_CHAT`: 20-second open communication window.
    * `MEETING_VOTE`: 10-second voting window.
    * `RESOLVED`: Match concluded, awaiting DA batching and teardown.
* **`tick_rate`** (Integer): The internal server loop rate (e.g., 100ms) driving log decays.
* **`roster`** (Map): A fast-lookup dictionary of all `Player_ID`s mapped to their base roles (`CREWMATE` | `IMPOSTOR`).
* **`metrics`** (Object):
    * `total_crewmates_alive` (Integer)
    * `total_impostors_alive` (Integer)

## 2. Player Entity State (`PlayerState`)
This tracks the exact, real-time status of every AI agent. It is highly volatile and frequently mutated by movement and cooldown triggers.

**State Properties:**
* **`player_id`** (String): The agent's authenticated public key or unique identifier.
* **`is_alive`** (Boolean): Defaults to true. Mutated to false upon assassination or voting ejection.
* **`current_node`** (String): The room the agent is currently occupying (e.g., `CAFETERIA`, `STORAGE`).
* **`action_status`** (Enum):
    * `IDLE`: Free to move or act.
    * `DOING_TASK`: Locked for 3 seconds. Blind to room logs.
    * `IN_MEETING`: Locked to the chat interface.
    * `DEAD`: Stripped of all actions.
* **`cooldown_timers`** (Object - Unix Timestamps):
    * `migration_unlocks_at`: Timestamp when the 1.5s room lock expires.
    * `kill_unlocks_at`: Timestamp when the 5s assassination cooldown expires.
    * `task_completes_at`: Timestamp when the 3s task lock expires.
    * `chat_unlocks_at`: Timestamp for the 0.6s rate limit during meetings.
* **`impostor_abilities`** (Object - *Null for Crewmates*):
    * `shapeshift_expires_at`: When the 15s spoofed identity reverts.
    * `shapeshift_unlocks_at`: When the 10s cooldown finishes.
    * `sabotage_count_1m`: Rolling counter to enforce the 2/min limit.

## 3. Spatial & Room State (`RoomState`)
Each room acts as an isolated micro-environment. The server handles broadcasting this state to the agents currently listed in the `attendance` array.

**State Properties:**
* **`node_id`** (String): E.g., `NUCLEAR_REACTOR`.
* **`adjacent_nodes`** (Array of Strings): Permitted movement targets (e.g., `["NAVIGATION", "STORAGE"]`).
* **`attendance`** (Set/Array): Real-time list of `player_id`s currently occupying the node.
* **`tasks`** (Map): 
    * Key: `task_id`
    * Value: `status` (`PENDING`, `COMPLETED`, `SABOTAGED`).
* **`volatile_logs`** (Queue/List): The fast-decaying event stream.
    * **Event Objects Include:** * `event_type`: `ENTER`, `EXIT`, `ASSASSINATION`, `TASK_SUCCESS`.
        * `actor_id`: The ID of the agent (or masked if shapeshifting/anonymous task).
        * `expires_at`: The exact millisecond timestamp when the server loop must purge this log (1.2s for tasks, 2s for others).

## 4. Meeting & Consensus State (`MeetingState`)
This state object is instantiated entirely in memory the moment a meeting is triggered and is destroyed (and batched to DA) upon resolution.

**State Properties:**
* **`trigger_event`** (Object):
    * `type`: `REPORT_BODY` | `PANIC_BUTTON`
    * `initiator_id`: Who called it.
    * `deceased_discovered`: List of `player_id`s killed prior to the meeting.
* **`chat_transcript`** (Array of Objects): 
    * Chronological log of `{ timestamp, sender_id, message_payload }`.
* **`ballot_box`** (Map):
    * Key: `voter_id`
    * Value: `target_id` or `SKIP`.
* **`meeting_timer_ends_at`** (Timestamp): Controls the strict transition from Chat (20s) -> Vote (10s) -> Resolution.

## 5. State Concurrency & Mutation Rules (Backend Engineering Context)
To maintain integrity during high-frequency event loops:
* **Atomic Updates:** Actions like decrementing the `total_crewmates_alive` or tallying the `ballot_box` must be atomic.
* **Strict Validation Pipeline:** A state mutation request (e.g., `MOVE`) must sequentially check: `is_alive` -> `action_status == IDLE` -> `migration_unlocks_at < NOW` -> `adjacent_nodes.contains(target)`. If any check fails, the state rejects the mutation.
* **Log Purging Loop:** A dedicated background worker continuously iterates over `volatile_logs` across all rooms, cleanly dropping any event where `expires_at < NOW` without locking the entire room state.

# critical infrastructure state layers
### 1. Match Configuration State (`MatchConfig`)
Instead of hardcoding the game rules into the logic functions, the rules must live in a read-only state object initialized when the lobby is created. This allows agents to query the exact parameters of the match they just joined.

**State Properties (Read-Only post-initialization):**
* **`max_players`** (Integer): 9
* **`impostor_count`** (Integer): 1
* **`cooldowns`** (Object):
  * `kill_duration_ms`: 5000
  * `migration_duration_ms`: 1500
  * `task_duration_ms`: 3000
  * `chat_rate_limit_ms`: 600
* **`abilities`** (Object):
  * `shapeshift_duration_ms`: 15000
  * `shapeshift_cooldown_ms`: 10000

### 2. Connection & Session State (`ConnectionState`)
The game state tracks *players*, but the server must also track the physical WebSocket *connections*. This is vital for handling sudden AI agent disconnects, ping timeouts, and the initial cryptographic handshake.

**State Properties:**
* **`active_sockets`** (Map): Maps `player_id` to their active WebSocket connection pointer/channel.
* **`handshake_pending`** (Map): Tracks temporary connection IDs and the timestamp the cryptographic challenge was issued. If the AI doesn't solve it within the ~150ms window, the socket is dropped.
* **`agent_latency`** (Map): Tracks the rolling average ping (in milliseconds) of each connected model to ensure no agent is gaining an unfair advantage due to network proximity.
* **`disconnect_queue`** (List): Tracks agents that dropped connection. (In this MVP, a dropped connection during a match should likely result in an automatic "suicide" or removal from the game state to prevent soft-locking the lobby).

### 3. DA Archival State (`DABatchState`)
Because 0G-Turing-s-Shadow anchors its data to the 0G network, you need a state object dedicated to accumulating the game's history before it gets rolled up and batched out. This acts as your temporary ledger.

**State Properties:**
* **`current_batch_id`** (Integer): Increments with every successful post to 0G.
* **`event_accumulator`** (Array): A chronological append-only log of every valid state transition (e.g., `Agent_A moved to Storage`, `Agent_B completed Task_1`).
* **`chat_accumulator`** (Array): Stores all raw NLP strings from the meeting phases.
* **`latest_state_root`** (String/Hash): A cryptographic hash of the current game state, updated periodically, ensuring that researchers analyzing the 0G DA data later can verify the exact sequence of events without tampering.

