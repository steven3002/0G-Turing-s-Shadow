// React is imported via Vite transform
import { useGameStore } from '../store/gameStore';
import { Radar, Users, Activity } from 'lucide-react';

const ROOM_IMAGES: Record<string, string> = {
  'CAFETERIA': 'cafeteria_background_1774365860155.png',
  'NAVIGATION': 'navigation_background_1774365883171.png',
  'ELECTRICAL': 'electrical_background_1774365902204.png',
  'NUCLEAR_REACTOR': 'reactor_background_1774365925579.png',
  'MEDBAY': 'medbay_background_1774365942408.png',
  'STORAGE': 'storage_background_1774365960091.png'
};

export default function GameDashboard({ sendAction }: { sendAction: (action: any) => void }) {
  const { currentNode, adjacentNodes, localLogs, attendance, players } = useGameStore();
  const playerList = Object.values(players);

  return (
    <div style={{ display: 'flex', width: '100%', gap: '1rem', height: '100%', position: 'relative' }}>
      {/* Left Sidebar: Roster */}
      <div className="glass-panel" style={{ width: '250px', display: 'flex', flexDirection: 'column' }}>
        <h3 className="text-cyan" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem' }}>
          <Users size={18} /> Roster
        </h3>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', flex: 1, overflowY: 'auto' }}>
          {playerList.length === 0 ? (
             <p className="text-muted mono" style={{ fontSize: '0.8rem', padding: '1rem' }}>No agents detected...</p>
          ) : playerList.map((player: any) => {
            const isHere = attendance.includes(player.id);
            return (
              <div key={player.id} style={{ 
                display: 'flex', 
                alignItems: 'center', 
                justifyContent: 'space-between',
                padding: '0.5rem',
                background: isHere ? 'rgba(0, 229, 255, 0.1)' : 'rgba(255,255,255,0.05)',
                border: isHere ? '1px solid var(--accent-cyan)' : '1px solid transparent',
                borderRadius: '6px',
                transition: 'all 0.3s ease'
              }}>
                <span className="mono" style={{ fontSize: '0.85rem' }}>{player.id}</span>
                <span style={{ 
                  width: '8px', 
                  height: '8px', 
                  borderRadius: '50%', 
                  background: isHere ? 'var(--accent-cyan)' : 'var(--accent-green)', // Cyan means "I see them here", Green means "They are alive elsewhere"
                  boxShadow: `0 0 5px ${isHere ? 'var(--accent-cyan)' : 'var(--accent-green)'}`
                }}></span>
              </div>
            );
          })}
        </div>
        
        <button className="btn-danger alert-pulse" style={{ marginTop: '1rem' }} onClick={() => sendAction({ action: 'PANIC_BUTTON', payload: {} })}>
          EMERGENCY PANIC
        </button>
      </div>

      {/* Center: Strategic Map */}
      <div className="glass-panel room-active" style={{ flex: 1, position: 'relative', overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
        <div className="room-bg" style={{ backgroundImage: `url(/${ROOM_IMAGES[currentNode] || ''})` }}></div>
        
        <div style={{ zIndex: 1, padding: '1rem', borderBottom: '1px solid var(--border-glow)' }}>
          <h2 className="text-cyan" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Radar size={24} /> {currentNode}
          </h2>
        </div>

        <div style={{ flex: 1, position: 'relative', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          {/* Agent Markers In Room */}
          <div style={{ 
            position: 'absolute', 
            top: '50%', left: '50%', 
            transform: 'translate(-50%, -50%)',
            display: 'flex', flexWrap: 'wrap', gap: '10px', maxWidth: '300px',
            justifyContent: 'center', zIndex: 2
          }}>
            {attendance.map((id) => (
              <div key={id} className="agent-marker" style={{
                width: '40px', height: '40px',
                background: 'rgba(0, 229, 255, 0.2)',
                border: '2px solid var(--accent-cyan)',
                borderRadius: '50%',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
                fontSize: '0.6rem', fontWeight: 'bold', color: 'white',
                backdropFilter: 'blur(5px)',
                boxShadow: '0 0 15px var(--accent-cyan)'
              }}>
                {id.slice(-4).toUpperCase()}
              </div>
            ))}
          </div>

          <div style={{ position: 'relative', width: '200px', height: '60px', background: 'var(--accent-cyan)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#000', fontWeight: 'bold', borderRadius: '4px', zIndex: 1, boxShadow: '0 0 30px var(--accent-cyan)' }}>
             YOU ARE HERE
          </div>

          {adjacentNodes.map((node, i) => (
            <div key={node} 
              onClick={() => sendAction({ action: 'MOVE', payload: { destination: node } })}
              style={{
                position: 'absolute',
                padding: '1rem 2rem',
                background: 'rgba(5, 11, 20, 0.9)',
                border: '1px solid var(--accent-cyan)',
                color: 'var(--text-main)',
                borderRadius: '8px',
                cursor: 'pointer',
                transition: 'all 0.3s ease',
                transform: `translate(${i === 0 ? '-220px' : i === 1 ? '220px' : '0'}, ${i === 2 ? '-120px' : i === 3 ? '120px' : '0'})`,
                zIndex: 2
              }}
              onMouseEnter={(e) => e.currentTarget.style.boxShadow = '0 0 20px var(--accent-cyan)'}
              onMouseLeave={(e) => e.currentTarget.style.boxShadow = 'none'}
            >
              MOVE TO {node}
            </div>
          ))}
        </div>
      </div>

      {/* Right Sidebar: Logs */}
      <div className="glass-panel" style={{ width: '300px', display: 'flex', flexDirection: 'column' }}>
        <h3 className="text-green" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem' }}>
          <Activity size={18} /> Room Decaying Logs
        </h3>
        <div style={{ flex: 1, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          {localLogs.length === 0 ? (
            <p className="text-muted" style={{ fontStyle: 'italic', fontSize: '0.9rem', padding: '1rem' }}>Awaiting fast-decay events...</p>
          ) : (
            localLogs.map((log) => (
              <div key={log.EventID} className="agent-marker" style={{
                fontFamily: 'var(--font-mono)',
                fontSize: '0.8rem',
                padding: '0.6rem',
                borderLeft: log.Action === 'ASSASSINATED' ? '3px solid var(--accent-red)' : '3px solid var(--accent-cyan)',
                background: 'rgba(0,0,0,0.4)',
                borderRadius: '0 4px 4px 0'
              }}>
                <span className="text-muted" style={{fontSize: '0.7rem'}}>
                  [{new Date(log.ExpiresAt).toISOString().split('T')[1].slice(0, 8)}]
                </span><br/>
                <span style={{ color: log.Action === 'ASSASSINATED' ? 'var(--accent-red)' : 'var(--accent-green)' }}>
                  {log.ActorID}
                </span> {log.Action}
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
