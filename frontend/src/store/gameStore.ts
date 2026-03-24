import { create } from 'zustand';

export type GamePhase = 'LOBBY' | 'IN_PLAY' | 'MEETING_CHAT' | 'MEETING_VOTE' | 'RESOLVED';

export interface PlayerState {
  player_id: string;
  is_alive: boolean;
  current_node?: string;
  role?: 'CREWMATE' | 'IMPOSTOR'; // Only revealed to the specific player or at end
  action_status: 'IDLE' | 'DOING_TASK' | 'IN_MEETING' | 'DEAD';
}

export interface ActivityLogEvent {
  EventID: string;
  Action: 'ENTERED' | 'EXITED' | 'ASSASSINATED' | 'TASK_COMPLETED' | 'CRITICAL_SABOTAGE_ALARM';
  ActorID: string;
  ExpiresAt: number;
}

interface GameStore {
  phase: GamePhase;
  currentNode: string;
  adjacentNodes: string[];
  localLogs: ActivityLogEvent[];
  chatTranscript: any[];
  players: Record<string, any>;
  attendance: string[]; // New: list of players in current room
  
  setPhase: (p: GamePhase) => void;
  updateLocalMap: (current: string, adjacents: string[]) => void;
  updatePlayers: (players: string[]) => void;
  setAttendance: (players: string[]) => void;
  addLog: (log: ActivityLogEvent) => void;
  addChat: (chat: any) => void;
}

export const useGameStore = create<GameStore>((set) => ({
  phase: 'LOBBY',
  currentNode: 'CAFETERIA',
  adjacentNodes: ['NAVIGATION', 'STORAGE'],
  localLogs: [],
  chatTranscript: [],
  players: {},
  attendance: [],

  setPhase: (p) => set({ phase: p }),
  updateLocalMap: (current, adjacents) => set({ currentNode: current, adjacentNodes: adjacents }),
  updatePlayers: (players) => set((state) => {
    const newPlayers: Record<string, any> = {};
    players.forEach(p => { newPlayers[p] = { id: p, isAlive: true }; });
    return { players: newPlayers };
  }),
  setAttendance: (players) => set({ attendance: players }),
  addLog: (log) => set((state) => {
    const exists = state.localLogs.some(l => l.EventID === log.EventID);
    if (exists) return state;
    return { localLogs: [log, ...state.localLogs].slice(0, 20) };
  }),
  addChat: (chat) => set((state) => ({ chatTranscript: [...state.chatTranscript, chat] })),
}));
