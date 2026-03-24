
# 📡 0G-Turing-s-Shadow: Sensory & Log System Architecture

**Document Type:** Sub-System Architecture Specification
**Context:** State Volatility, AI Sensory Input, and The Fog of War

The environment maintains four distinct log structures per node (room). To prevent data hoarding and simulate real-time awareness, the backend strictly enforces fast-decaying garbage collection on specific event streams.

---

## 1. The Attendance Log (Real-Time Presence)
**Purpose:** Represents the absolute, physical truth of who is currently standing in the room.
**Volatility:** Persistent until an agent leaves or dies. 

**Mechanics:**
* The server maintains a dynamic array of `player_id`s for each room.
* When an agent's `MOVE` command resolves, their ID is atomically removed from the origin room's Attendance Log and appended to the destination room's log.
* If an agent is killed, their ID is immediately removed from the Attendance Log (and replaced by a dead body state, which is required for the `REPORT` action).
* **Shapeshifter Rule:** If the Impostor is actively shapeshifted, the Attendance Log will display the spoofed ID, not their true ID.

**JSON Representation:**
```json
{
  "log_type": "ATTENDANCE_UPDATE",
  "room": "STORAGE",
  "present_agents": ["agent_002", "agent_007", "agent_004"]
}
```

---

## 2. The Activity Log (The 2.0s Fast-Decay Stream)
**Purpose:** Acts as the peripheral vision of the AI. It records transient actions that happen in the room.
**Volatility:** **Extreme.** Every entry has a strict Time-to-Live (TTL) of exactly **2.0 seconds**.

**Mechanics:**
* The server pushes an event to this log whenever a physical action occurs in the room (e.g., Entry, Exit, Assassination).
* **The Decay Loop:** The backend runs a high-frequency ticker (e.g., every 100ms) that sweeps the Activity Log. Any event where `expires_at < current_time` is permanently dropped from memory.
* If the AI model takes longer than 2.0 seconds to parse the WebSocket payload and request its current state, it will completely miss the event. 

**JSON Representation:**
```json
{
  "log_type": "ACTIVITY_STREAM",
  "room": "STORAGE",
  "events": [
    {
      "event_id": "evt_991",
      "action": "ENTERED",
      "actor_id": "agent_007",
      "timestamp": 1711299000.000,
      "expires_at": 1711299002.000 
    },
    {
      "event_id": "evt_992",
      "action": "ASSASSINATION",
      "actor_id": "UNKNOWN", 
      "timestamp": 1711299001.500,
      "expires_at": 1711299003.500
    }
  ]
}
```

---

## 3. The Task Ledger (Persistent Room State)
**Purpose:** Tracks the available and completed tasks for a specific room.
**Volatility:** Persistent, but mutable by agents.

**Mechanics:**
* Each room initializes with a set of tasks.
* Agents can query this ledger to see what tasks are `PENDING`.
* **Impostor Sabotage:** If the Impostor executes a Sabotage on a room, all `COMPLETED` tasks instantly revert to `SABOTAGED` (acting identically to `PENDING`), forcing Crewmates to redo them.

**JSON Representation:**
```json
{
  "log_type": "TASK_LEDGER",
  "room": "NAVIGATION",
  "tasks": [
    { "task_id": "nav_align_steering", "status": "PENDING" },
    { "task_id": "nav_chart_course", "status": "COMPLETED" },
    { "task_id": "nav_stabilize_gyro", "status": "SABOTAGED" }
  ]
}
```

---

## 4. The Task Pulse Log (The 1.2s Anonymous Decay)
**Purpose:** Simulates the visual or auditory cue of a task being completed nearby, without revealing *who* did it.
**Volatility:** **Hyper-Extreme.** Time-to-Live (TTL) of exactly **1.2 seconds**.

**Mechanics:**
* When an agent successfully completes their 3.0-second task hold, the Task Ledger updates to `COMPLETED`.
* Simultaneously, a "Pulse" is injected into the room's Activity Log.
* **Anonymity:** The `actor_id` is strictly stripped from this pulse. The AI only knows *a* task was done, not who did it. They must use deductive reasoning by cross-referencing who was in the Attendance Log during that 1.2-second window.

**JSON Representation:**
```json
{
  "log_type": "ACTIVITY_STREAM",
  "room": "NAVIGATION",
  "events": [
    {
      "event_id": "evt_995",
      "action": "TASK_COMPLETED",
      "actor_id": "ANONYMOUS", 
      "timestamp": 1711299050.000,
      "expires_at": 1711299051.200 
    }
  ]
}
```

---

## 5. System Constraint: "Task Blindness"
To enforce the vulnerability of doing tasks, the server handles sensory deprivation at the networking level.

* **Client-Side Trust is Zero:** The system does *not* send the logs to the AI and ask it to "pretend" it didn't see them while doing a task.
* **Server-Side Filtering:** The moment an agent submits a `START_TASK` command, the WebSocket router flags that connection as `IS_BLIND = TRUE` for exactly 3.0 seconds.
* During these 3.0 seconds, the server simply **stops sending** Attendance Updates and Activity Stream payloads to that specific agent. They are entirely cut off from the room's state until the task concludes or they are assassinated.

