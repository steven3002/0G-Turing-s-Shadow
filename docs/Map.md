# 🗺️ 0G-Turing-s-Shadow: Map Topology & Local Map Protocol

**Document Type:** Sub-System Architecture Specification

**Context:** Spatial Logic and Agent Navigation

To simulate spatial awareness for text-based AI models, the environment abandons continuous 2D geometry (X, Y coordinates) in favor of a strictly defined node-based graph. Agents traverse "edges" (connections) between "nodes" (rooms).

---

<img width="1408" height="768" alt="image" src="https://github.com/user-attachments/assets/1e2a44d9-75a2-4743-b0eb-da070088c179" />


## 1. The Global Map Architecture (Node Graph)

The MVP global map consists of exactly 6 localized nodes. There are currently no transitional spaces or "corridors" (planned for V2+); moving from one node to a connected node is instantaneous, subject to cooldowns.

**The Definitive Routing Graph:**
The map operates on bidirectional edges. If Node A connects to Node B, Node B connects to Node A. The allowed traversal paths are strictly limited to the following matrix:

* **Cafeteria** ↔ Navigation
* **Cafeteria** ↔ Storage
* **Navigation** ↔ Electrical
* **Navigation** ↔ Nuclear Reactor
* **Storage** ↔ Nuclear Reactor
* **Storage** ↔ Medbay

*Note: There are no secret vents or teleportation mechanics in the MVP. The Impostor must abide by the exact same physical routing graph as the Crewmates.*

---

## 2. Node Specifics & Properties

Certain nodes possess unique state properties that affect gameplay mechanics:

* **Cafeteria (The Hub):** * **Spawn Point:** All 9 agents spawn here at the exact start of the match and immediately following any Meeting Resolution.
    * **The Panic Button:** The Cafeteria is the *only* node in the game where an agent can trigger an Emergency Meeting without requiring a dead body.
* **Nuclear Reactor (The Choke Point):**
    * Serves as the primary intersection between the left side of the map (Navigation) and the right side (Storage), making it a high-traffic area and a strategic location for the Impostor.
* **Terminal Nodes (Electrical & Medbay):**
    * These are dead-end nodes with only a single entry/exit vector. If an agent enters Medbay, the only possible exit route is back through Storage.

---

## 3. The Local Map Protocol (Agent Vision)

Agents do not possess a global view of the map state. They must explicitly query their environment to understand their navigation options. This is handled via the **Local Map** system.

**3.1 Concept:** The Local Map is a contextual data payload requested by an agent. It provides localized spatial awareness relative *only* to the agent's current node.

**3.2 Data Provided in the Local Map:**
When an agent requests the Local Map, the server responds with:
1.  **Current Location:** The exact node the agent is currently occupying.
2.  **Valid Destinations:** An array of directly adjacent nodes that the agent can legally traverse to in a single move.
3.  **State Constraints:** The current status of the agent's movement lock (e.g., whether the 1.5s migration cooldown is currently active, preventing immediate departure).

**3.3 Example Logic Flow:**
* *Scenario:* Agent `0x1A` is currently in **Storage**.
* *Action:* Agent queries the Local Map.
* *Server Response:* Informs the agent they are in `Storage`, and their valid next moves are restricted entirely to `Nuclear Reactor`, `Medbay`, or `Cafeteria`. The server omits `Electrical` and `Navigation` because they are not directly adjacent.

---

## 4. Traversal Constraints & Edge Rules

Navigating the map is subject to the system's high-performance event loop and cooldown restrictions.

**4.1 The One-Tile Rule:** An agent can only move one node at a time. An API request attempting to route an agent from the `Cafeteria` directly to `Electrical` will be forcefully rejected by the server because they do not share a direct edge. The agent must first move to `Navigation`.

**4.2 Migration Cooldown (1.5 Seconds):**
To prevent AI agents from rapidly cycling through all rooms at computational speed, a strict movement penalty is enforced:
* Upon successfully processing a move command, the agent arrives at the destination node.
* The agent is immediately locked in that room for **1.5 seconds**. 
* During this window, the agent can observe Activity Logs, execute Tasks, or Kill (if Impostor), but any attempt to query a new move or exit the node will fail until the timer expires.
