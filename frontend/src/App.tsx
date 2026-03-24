// React is imported via Vite transform
import { useGameStore } from './store/gameStore';
import './index.css';

// We will mount the specific phase components here.
import LobbyView from './components/LobbyView';
import GameDashboard from './components/GameDashboard';
import MeetingView from './components/MeetingView';

import { useGameWebSocket } from './hooks/useGameWebSocket';

function App() {
  const { phase } = useGameStore();
  const { sendAction } = useGameWebSocket('ws://localhost:8080', 'alpha_squad');

  return (
    <div className="app-container">
      {phase === 'LOBBY' && <LobbyView />}
      {phase === 'IN_PLAY' && <GameDashboard sendAction={sendAction} />}
      {(phase === 'MEETING_CHAT' || phase === 'MEETING_VOTE') && (
        <>
          <GameDashboard sendAction={sendAction} /> {/* Keep map behind */}
          <MeetingView sendAction={sendAction} />   {/* Overlay modal */}
        </>
      )}
      {phase === 'RESOLVED' && <div className="glass-panel" style={{margin: 'auto'}}><h2>Match Concluded</h2><p>Data batched to 0G DA Layer.</p></div>}
    </div>
  );
}

export default App;
