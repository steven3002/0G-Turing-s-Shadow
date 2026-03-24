import { useState } from 'react';
import { useGameStore } from '../store/gameStore';

export default function LobbyView() {
  const { setPhase } = useGameStore();
  const [key, setKey] = useState('');

  const handleConnect = (e: React.FormEvent) => {
    e.preventDefault();
    if (!key) return;
    
    // In MVP, we just mock the handshake and transition
    setPhase('IN_PLAY');
  };

  return (
    <div className="glass-panel" style={{ margin: 'auto', maxWidth: '400px', width: '100%', textAlign: 'center' }}>
      <h1 className="text-cyan" style={{ marginBottom: '1rem' }}>0G-Turing-s-Shadow</h1>
      <p className="text-muted" style={{ marginBottom: '2rem', fontSize: '0.9rem' }}>
        Averifiable benchmark for AI strategic planning.<br/>
        Initiate cryptographic handshake protocol.
      </p>

      <form onSubmit={handleConnect} style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
        <input 
          type="text" 
          value={key}
          onChange={(e) => setKey(e.target.value)}
          placeholder="Enter Access Key..." 
          style={{
            padding: '1rem',
            background: 'rgba(0,0,0,0.5)',
            border: '1px solid var(--border-glow)',
            color: 'var(--accent-cyan)',
            fontFamily: 'var(--font-mono)',
            borderRadius: '8px',
            outline: 'none'
          }}
        />
        <button type="submit" className="btn-primary">
          Initiate Handshake
        </button>
      </form>
    </div>
  );
}
