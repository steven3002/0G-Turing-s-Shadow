import { useEffect, useRef, useState } from 'react';
import { useGameStore } from '../store/gameStore';

export function useGameWebSocket(url: string, roomId: string) {
  const ws = useRef<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const { setPhase, updateLocalMap, addLog, addChat } = useGameStore();

  useEffect(() => {
    // 1. Connect with room parameter
    ws.current = new WebSocket(`${url}/lobby/join?room=${roomId}`);

    ws.current.onopen = () => {
      console.log('WebSocket Connection Opened');
    };

    ws.current.onmessage = (event) => {
      try {
        // 2. Handle robustness for concatenated JSON (NDJSON)
        const messages = event.data.split('\n');
        
        messages.forEach((msgStr: string) => {
          if (!msgStr.trim()) return;
          const payload = JSON.parse(msgStr);
          
          // Automated Handshake Logic
          if (payload.system_event === 'HANDSHAKE_CHALLENGE') {
            const { nonce, multiplier } = payload.payload;
            const reversed = nonce.split('').reverse().join('');
            const solution = `${reversed}${multiplier}`;
            
            ws.current?.send(JSON.stringify({
              client_action: 'HANDSHAKE_RESPONSE',
              payload: { solution_string: solution }
            }));
            console.log('Handshake Solved & Sent');
            setIsConnected(true);
            return;
          }

          // 3. Backend Prompt Mapping
          switch (payload.prompt_id || payload.type) {
            case 'PROMPT_01_WELCOME':
              if (payload.players) {
                useGameStore.getState().updatePlayers(payload.players);
              }
              if (payload.room) {
                useGameStore.getState().updateLocalMap(payload.room, []); // Initial room
              }
              if (payload.phase) {
                useGameStore.getState().setPhase(payload.phase);
              }
              break;
            case 'PROMPT_05_STATE_TICK_UPDATE':
              if (payload.activity_log) {
                payload.activity_log.forEach(addLog);
              }
              if (payload.attendance) {
                useGameStore.getState().setAttendance(payload.attendance);
              }
              break;
            case 'PROMPT_06_LOCAL_MAP_UPDATE':
              updateLocalMap(payload.current_node, payload.adjacent_nodes);
              break;
            case 'PROMPT_08_MEETING_START_CONTEXT':
              setPhase('MEETING_CHAT');
              break;
            case 'PROMPT_09_CHAT_EVALUATION_TICK':
            case 'meeting_chat_update':
              addChat({
                sender_id: payload.sender || 'UNKNOWN',
                message_payload: payload.message || payload.text,
                timestamp: payload.timestamp || Date.now()
              });
              break;
            case 'PROMPT_10_VOTING_DEMAND':
              setPhase('MEETING_VOTE');
              break;
            case 'PROMPT_11_MATCH_RESOLUTION':
              setPhase('RESOLVED');
              break;
          }
        });
      } catch (err) {
        console.error('Failed to parse websocket message', err);
      }
    };

    ws.current.onclose = () => {
      console.log('WebSocket Disconnected');
      setIsConnected(false);
    };

    return () => {
      ws.current?.close();
    };
  }, [url, roomId]);

  const sendAction = (action: any) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(action));
    }
  };

  return { isConnected, sendAction };
}
