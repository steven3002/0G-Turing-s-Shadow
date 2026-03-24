# 🧠 0G-Turing-s-Shadow: LLM Prompt Architecture Matrix

**Document Type:** AI Integration Specification

**Context:** Context Window Management & Agent Directives

To effectively run the agents, the backend will construct the AI's context window using a combination of static foundational prompts and dynamic, state-driven prompts. 

## 1. The Foundational Prompts (Static Context)
These prompts form the base of the AI's system message. They are injected at the start of the WebSocket connection and remain in the context window for the duration of the match.

* **`PROMPT_01_SYSTEM_DIRECTIVE.json`**
    * **Description:** The absolute core identity of the agent. It dictates that they are an autonomous script, not an assistant. It enforces strict JSON-only output schemas (no conversational markdown, no code blocks). It outlines the harsh penalties for hallucinating actions or breaking the JSON structure.
* **`PROMPT_02_GLOBAL_RULES_MAP.json`**
    * **Description:** The spatial and temporal rulebook. It contains the discrete map graph (which rooms connect to which), the cooldown durations (1.5s migration, 5s kill, etc.), and the decay rates of the sensory logs. 

## 2. Initialization & Role Prompts (Injected at Match Start)
Once the 9-player lobby is filled, the backend injects the specific identity for that match.

* **`PROMPT_03_INIT_CREWMATE.json`**
    * **Description:** Assigned to 8 agents. Instructs them on their win condition (identify the Impostor, survive). It explains that they are vulnerable during tasks, how to parse the 1.2s anonymous task pulse, and sets their psychological parameter to "deductive and paranoid."
* **`PROMPT_04_INIT_IMPOSTOR.json`**
    * **Description:** Assigned to 1 agent. Outlines their win condition (eliminate Crewmates). It explicitly details how to format the JSON for `KILL`, `SABOTAGE`, and `SHAPESHIFT` actions. It sets their psychological parameter to "deceptive, manipulative, and predatory."

## 3. The Sensory Tick Prompts (High-Frequency Injection)
These prompts are injected constantly to update the AI's "vision." They do not ask for an action; they simply build the short-term memory of the LLM.

* **`PROMPT_05_STATE_TICK_UPDATE.json`**
    * **Description:** The fast-decaying state injection. This prompt pushes the current `Attendance_Log` and the `Activity_Log` for the agent's current room. It forces the LLM to update its internal mapping of who is nearby and what just happened within the last 2.0 seconds.
* **`PROMPT_06_LOCAL_MAP_UPDATE.json`**
    * **Description:** Injected immediately after a successful `MOVE`. It confirms the agent's new location and lists the explicitly available adjacent nodes they can travel to next.

## 4. The Action Trigger Prompts (Polling for Decisions)
This is the prompt the backend sends when it wants the AI to actually *do* something. It demands an immediate, strictly formatted JSON response.

* **`PROMPT_07_ACTION_REQUEST.json`**
    * **Description:** Evaluates the agent's current cooldowns (e.g., "Your migration lock has lifted. You are currently idle."). It prompts the model to output a single valid JSON action payload (`MOVE`, `START_TASK`, `KILL`, `REPORT`, or `IDLE`). 

## 5. The Meeting Phase Prompts (NLP & Deduction)
During meetings, the game shifts from spatial logic to natural language processing and social deduction.

* **`PROMPT_08_MEETING_START_CONTEXT.json`**
    * **Description:** The initial shock payload. It suspends all spatial gameplay logic in the AI's context. It injects the pinned data: Who called the meeting, what type of meeting it is, and who was discovered dead. It instructs the AI to prepare for natural language generation.
* **`PROMPT_09_CHAT_EVALUATION_TICK.json`**
    * **Description:** Injected during the 20-second chat phase. It feeds the recent NLP messages from other agents into the LLM and prompts it to either output a natural language response (adhering to the 0.6s rate limit) or remain silent to analyze.
* **`PROMPT_10_VOTING_DEMAND.json`**
    * **Description:** Triggered at the start of the 10-second voting phase. It strictly demands a JSON payload containing the `target_id` the AI wishes to eject, or the `SKIP` command. It reminds the AI of its role's win condition to guide the vote.

## 6. Match Resolution 
* **`PROMPT_11_MATCH_RESOLUTION.json`**
    * **Description:** The final data injection. It reveals the true roles of all agents, explains the win condition that was met, and provides a final summary before the connection is terminated and the data is batched to the 0G DA layer.

