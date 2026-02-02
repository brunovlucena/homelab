import { useState } from 'react';
import LibraryView from './components/LibraryView';
import PlayerView from './components/PlayerView';
import P2PConnection from './components/P2PConnection';
import MIDIController from './components/MIDIController';
import { useStore } from './store';

function App() {
  const [activeView, setActiveView] = useState<'library' | 'player' | 'p2p' | 'midi'>('library');
  const currentTrack = useStore((state) => state.currentTrack);

  return (
    <div className="min-h-screen bg-slate-900">
      {/* Header */}
      <header className="bg-slate-800 border-b border-slate-700 px-6 py-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-white">🎵 DJ Cloud P2P</h1>
          <nav className="flex gap-4">
            <button
              onClick={() => setActiveView('library')}
              className={`px-4 py-2 rounded-lg transition-colors ${
                activeView === 'library'
                  ? 'bg-primary-600 text-white'
                  : 'text-slate-300 hover:bg-slate-700'
              }`}
            >
              Biblioteca
            </button>
            <button
              onClick={() => setActiveView('p2p')}
              className={`px-4 py-2 rounded-lg transition-colors ${
                activeView === 'p2p'
                  ? 'bg-primary-600 text-white'
                  : 'text-slate-300 hover:bg-slate-700'
              }`}
            >
              Conexão P2P
            </button>
            <button
              onClick={() => setActiveView('midi')}
              className={`px-4 py-2 rounded-lg transition-colors ${
                activeView === 'midi'
                  ? 'bg-primary-600 text-white'
                  : 'text-slate-300 hover:bg-slate-700'
              }`}
            >
              Controladora
            </button>
            {currentTrack && (
              <button
                onClick={() => setActiveView('player')}
                className={`px-4 py-2 rounded-lg transition-colors ${
                  activeView === 'player'
                    ? 'bg-primary-600 text-white'
                    : 'text-slate-300 hover:bg-slate-700'
                }`}
              >
                Player
              </button>
            )}
          </nav>
        </div>
      </header>

      {/* Main Content */}
      <main className="p-6">
        {activeView === 'library' && <LibraryView />}
        {activeView === 'p2p' && <P2PConnection />}
        {activeView === 'midi' && <MIDIController />}
        {activeView === 'player' && currentTrack && <PlayerView />}
      </main>
    </div>
  );
}

export default App;
