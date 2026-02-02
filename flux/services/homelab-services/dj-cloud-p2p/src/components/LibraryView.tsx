import { useState, useEffect } from 'react';
import { useStore, Track } from '../store';
import { libraryService } from '../services/libraryService';

export default function LibraryView() {
  const { tracks, addTracks, setCurrentTrack } = useStore();
  const [scanning, setScanning] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  const handleScanLibrary = async () => {
    setScanning(true);
    try {
      // Por enquanto, vamos simular. Depois implementamos o scan real
      const newTracks = await libraryService.scanDirectory();
      addTracks(newTracks);
    } catch (error) {
      console.error('Erro ao escanear biblioteca:', error);
      alert('Erro ao escanear biblioteca. Veja o console para detalhes.');
    } finally {
      setScanning(false);
    }
  };

  const filteredTracks = tracks.filter((track) => {
    const query = searchQuery.toLowerCase();
    return (
      track.title.toLowerCase().includes(query) ||
      track.artist.toLowerCase().includes(query) ||
      track.album.toLowerCase().includes(query)
    );
  });

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Biblioteca</h2>
          <p className="text-slate-400 mt-1">
            {tracks.length} {tracks.length === 1 ? 'música' : 'músicas'}
          </p>
        </div>
        <button
          onClick={handleScanLibrary}
          disabled={scanning}
          className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {scanning ? 'Escaneando...' : 'Escanear Biblioteca'}
        </button>
      </div>

      {/* Search */}
      <div className="card">
        <input
          type="text"
          placeholder="Buscar músicas..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full bg-slate-700 text-white px-4 py-2 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
        />
      </div>

      {/* Tracks List */}
      <div className="card">
        {filteredTracks.length === 0 ? (
          <div className="text-center py-12 text-slate-400">
            {tracks.length === 0 ? (
              <>
                <p className="text-lg mb-2">Nenhuma música encontrada</p>
                <p className="text-sm">Clique em "Escanear Biblioteca" para começar</p>
              </>
            ) : (
              <p>Nenhuma música encontrada com "{searchQuery}"</p>
            )}
          </div>
        ) : (
          <div className="space-y-2">
            {filteredTracks.map((track) => (
              <div
                key={track.id}
                onClick={() => setCurrentTrack(track)}
                className="flex items-center gap-4 p-3 rounded-lg hover:bg-slate-700 cursor-pointer transition-colors"
              >
                <div className="flex-1 min-w-0">
                  <h3 className="text-white font-medium truncate">{track.title}</h3>
                  <p className="text-slate-400 text-sm truncate">
                    {track.artist} • {track.album}
                  </p>
                </div>
                <div className="text-slate-400 text-sm">
                  {formatDuration(track.duration)}
                </div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setCurrentTrack(track);
                  }}
                  className="btn-primary text-sm"
                >
                  Tocar
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
