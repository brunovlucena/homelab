import { useEffect, useRef, useState } from 'react';
import { useStore } from '../store';
import { Howl } from 'howler';
import { useMIDIController } from '../hooks/useMIDIController';

export default function PlayerView() {
  const { currentTrack, isPlaying, setPlaying, setCurrentTime } = useStore();
  const [sound, setSound] = useState<Howl | null>(null);
  const soundRef = useRef<Howl | null>(null);
  const progressIntervalRef = useRef<NodeJS.Timeout | null>(null);

  // Integrate MIDI controller
  useMIDIController(soundRef);

  useEffect(() => {
    if (!currentTrack) return;

    // Criar nova instância Howl
    const newSound = new Howl({
      src: [`file://${currentTrack.filePath}`],
      html5: true,
      onplay: () => setPlaying(true),
      onpause: () => setPlaying(false),
      onend: () => {
        setPlaying(false);
        setCurrentTime(0);
      },
    });

    setSound(newSound);
    soundRef.current = newSound;

    // Atualizar progresso
    progressIntervalRef.current = setInterval(() => {
      if (newSound.playing()) {
        setCurrentTime(newSound.seek() as number);
      }
    }, 100);

    return () => {
      if (progressIntervalRef.current) {
        clearInterval(progressIntervalRef.current);
      }
      newSound.unload();
      soundRef.current = null;
    };
  }, [currentTrack]);

  const handlePlayPause = () => {
    if (!sound) return;

    if (isPlaying) {
      sound.pause();
    } else {
      sound.play();
    }
  };

  const handleSeek = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!sound) return;
    const newTime = parseFloat(e.target.value);
    sound.seek(newTime);
    setCurrentTime(newTime);
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  if (!currentTrack) {
    return (
      <div className="card text-center py-12">
        <p className="text-slate-400">Nenhuma música selecionada</p>
      </div>
    );
  }

  const progress = currentTrack.duration > 0 
    ? (currentTime / currentTrack.duration) * 100 
    : 0;

  return (
    <div className="card max-w-2xl mx-auto">
      <div className="text-center mb-6">
        <h2 className="text-2xl font-bold text-white mb-2">{currentTrack.title}</h2>
        <p className="text-slate-400">{currentTrack.artist}</p>
        <p className="text-slate-500 text-sm">{currentTrack.album}</p>
      </div>

      {/* Progress Bar */}
      <div className="mb-6">
        <input
          type="range"
          min="0"
          max={currentTrack.duration}
          value={currentTime}
          onChange={handleSeek}
          className="w-full h-2 bg-slate-700 rounded-lg appearance-none cursor-pointer accent-primary-500"
        />
        <div className="flex justify-between text-sm text-slate-400 mt-1">
          <span>{formatTime(currentTime)}</span>
          <span>{formatTime(currentTrack.duration)}</span>
        </div>
      </div>

      {/* Controls */}
      <div className="flex items-center justify-center gap-4">
        <button
          onClick={handlePlayPause}
          className="w-16 h-16 rounded-full bg-primary-600 hover:bg-primary-700 flex items-center justify-center text-white text-2xl transition-colors"
        >
          {isPlaying ? '⏸' : '▶'}
        </button>
      </div>

      {/* Metadata */}
      <div className="mt-6 pt-6 border-t border-slate-700">
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-slate-400">Formato:</span>
            <span className="text-white ml-2">{currentTrack.metadata.format}</span>
          </div>
          <div>
            <span className="text-slate-400">Bitrate:</span>
            <span className="text-white ml-2">{currentTrack.metadata.bitrate} kbps</span>
          </div>
          {currentTrack.bpm && (
            <div>
              <span className="text-slate-400">BPM:</span>
              <span className="text-white ml-2">{currentTrack.bpm}</span>
            </div>
          )}
          {currentTrack.key && (
            <div>
              <span className="text-slate-400">Key:</span>
              <span className="text-white ml-2">{currentTrack.key}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
