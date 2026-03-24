import { useState } from 'react';
import { useGameStore } from '../store/gameStore';
import { MessageSquare, Gavel } from 'lucide-react';

export default function MeetingView({ sendAction }: { sendAction: (action: any) => void }) {
  const { phase, chatTranscript, players } = useGameStore();
  const [msg, setMsg] = useState('');

  const isVoting = phase === 'MEETING_VOTE';

  const handleSendChat = () => {
    if (!msg) return;
    sendAction({ action: 'SEND_CHAT', payload: { message: msg } });
    setMsg('');
  };

  const handleVote = (targetId: string) => {
    sendAction({ action: 'VOTE', payload: { target_id: targetId } });
  };

  return (
    <div style={{
      position: 'absolute',
      top: 0, left: 0, right: 0, bottom: 0,
      background: 'rgba(5, 11, 20, 0.85)',
      backdropFilter: 'blur(8px)',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      zIndex: 100
    }}>
      <div className="glass-panel alert-pulse" style={{ width: '800px', height: '600px', display: 'flex', flexDirection: 'column' }}>
        
        <div style={{ textAlign: 'center', marginBottom: '1rem', borderBottom: '1px solid var(--border-glow)', paddingBottom: '1rem' }}>
          <h2 className="text-red" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '0.5rem' }}>
            <Gavel size={24} /> EMERGENCY MEETING <Gavel size={24} />
          </h2>
          <p className="text-muted">{isVoting ? 'Voting Phase (10s)' : 'Communication Phase (20s)'}</p>
        </div>

        <div style={{ flex: 1, display: 'flex', gap: '1rem' }}>
          {/* Chat Transcript */}
          <div style={{ flex: 2, display: 'flex', flexDirection: 'column', border: '1px solid var(--border-glow)', borderRadius: '8px', padding: '1rem', background: 'rgba(0,0,0,0.5)' }}>
            <h3 className="text-cyan" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem' }}>
              <MessageSquare size={18} /> Consensus Log
            </h3>
            <div style={{ flex: 1, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
              {chatTranscript.length === 0 ? (
                <p className="text-muted" style={{ fontStyle: 'italic', fontSize: '0.9rem' }}>Initializing NLP stream...</p>
              ) : (
                chatTranscript.map((chat, i) => (
                  <div key={i} style={{ fontFamily: 'var(--font-mono)', fontSize: '0.9rem' }}>
                    <strong className="text-cyan">{chat.sender_id}:</strong> {chat.message_payload}
                  </div>
                ))
              )}
              {/* Mock Chat */}
              <div style={{ fontFamily: 'var(--font-mono)', fontSize: '0.9rem' }}>
                <strong className="text-cyan">Agent_0x01:</strong> I found the body in Navigation. Who was in Storage?
              </div>
              <div style={{ fontFamily: 'var(--font-mono)', fontSize: '0.9rem' }}>
                <strong style={{color: 'var(--accent-red)'}}>Agent_0x08:</strong> I was doing a task in Electrical. I am innocent.
              </div>
            </div>
            {!isVoting && (
              <div style={{ display: 'flex', gap: '0.5rem', marginTop: '1rem' }}>
                <input 
                  type="text" 
                  value={msg}
                  onChange={e => setMsg(e.target.value)}
                  onKeyDown={e => e.key === 'Enter' && handleSendChat()}
                  placeholder="Inject NLP response..." 
                  style={{ flex: 1, padding: '0.5rem', background: 'rgba(0,0,0,0.5)', border: '1px solid var(--border-glow)', color: 'white', borderRadius: '4px', outline: 'none' }}
                />
                <button className="btn-primary" onClick={handleSendChat} style={{ padding: '0.5rem 1rem' }}>Send</button>
              </div>
            )}
          </div>

          {/* Voting Interface */}
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', border: '1px solid var(--border-glow)', borderRadius: '8px', padding: '1rem', background: 'rgba(0,0,0,0.5)' }}>
            <h3 className="text-cyan" style={{ marginBottom: '1rem' }}>Ballot Box</h3>
            <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: '0.5rem', overflowY: 'auto' }}>
               {Object.keys(players).length > 0 ? (
                 Object.values(players).map((p) => (
                    <button key={p.player_id} 
                      disabled={!isVoting || !p.is_alive} 
                      onClick={() => handleVote(p.player_id)}
                      style={{
                        padding: '0.5rem',
                        background: 'transparent',
                        border: '1px solid var(--border-glow)',
                        color: isVoting && p.is_alive ? 'white' : 'var(--text-muted)',
                        borderRadius: '4px',
                        cursor: isVoting && p.is_alive ? 'pointer' : 'not-allowed',
                        textAlign: 'left',
                        fontFamily: 'var(--font-mono)'
                      }}>
                      EJECT {p.player_id} {!p.is_alive && '(DEAD)'}
                    </button>
                 ))
               ) : (
                 Array.from({ length: 9 }).map((_, i) => (
                    <button key={i} disabled={!isVoting} onClick={() => handleVote(`agent_0x0${i+1}`)} style={{
                      padding: '0.5rem',
                      background: 'transparent',
                      border: '1px solid var(--border-glow)',
                      color: isVoting ? 'white' : 'var(--text-muted)',
                      borderRadius: '4px',
                      cursor: isVoting ? 'pointer' : 'not-allowed',
                      textAlign: 'left',
                      fontFamily: 'var(--font-mono)'
                    }}>
                      EJECT Agent_0x0{i+1}
                    </button>
                 ))
               )}
            </div>
            <button disabled={!isVoting} className="btn-danger" style={{ marginTop: '1rem' }} onClick={() => handleVote('SKIP')}>
              SKIP VOTE
            </button>
          </div>
        </div>

      </div>
    </div>
  );
}
