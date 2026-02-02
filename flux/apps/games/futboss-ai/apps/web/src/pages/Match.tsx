// FutBoss AI - Match Page
// Author: Bruno Lucena (bruno@lucena.cloud)

import { useState, useEffect } from 'react';

type MatchState = 'lobby' | 'playing' | 'finished';

interface MatchEvent {
  minute: number;
  type: 'goal' | 'save' | 'foul' | 'card';
  team: 'home' | 'away';
  description: string;
}

export default function Match() {
  const [matchState, setMatchState] = useState<MatchState>('lobby');
  const [minute, setMinute] = useState(0);
  const [homeScore, setHomeScore] = useState(0);
  const [awayScore, setAwayScore] = useState(0);
  const [events, setEvents] = useState<MatchEvent[]>([]);
  const [possession, setPossession] = useState(50);

  useEffect(() => {
    if (matchState !== 'playing') return;

    const interval = setInterval(() => {
      setMinute((m) => {
        if (m >= 90) {
          setMatchState('finished');
          return 90;
        }
        
        // Random events
        if (Math.random() < 0.1) {
          const isHome = Math.random() < 0.5;
          if (Math.random() < 0.3) {
            // Goal!
            if (isHome) setHomeScore(s => s + 1);
            else setAwayScore(s => s + 1);
            setEvents(e => [...e, {
              minute: m,
              type: 'goal',
              team: isHome ? 'home' : 'away',
              description: `GOAL! ${isHome ? 'Home' : 'Away'} team scores!`
            }]);
          }
        }
        
        setPossession(50 + Math.floor(Math.random() * 20 - 10));
        return m + 1;
      });
    }, 500);

    return () => clearInterval(interval);
  }, [matchState]);

  const startMatch = () => {
    setMatchState('playing');
    setMinute(0);
    setHomeScore(0);
    setAwayScore(0);
    setEvents([]);
  };

  const getEventIcon = (type: string) => {
    switch (type) {
      case 'goal': return '⚽';
      case 'save': return '🧤';
      case 'foul': return '❌';
      case 'card': return '🟨';
      default: return '📢';
    }
  };

  return (
    <div>
      <h1 className="text-3xl font-bold mb-8">🎮 Match Center</h1>

      {matchState === 'lobby' && (
        <div className="card text-center py-12">
          <h2 className="text-2xl font-bold mb-4">Ready to Play?</h2>
          <p className="text-gray-400 mb-8">Start a new match against an AI opponent</p>
          <button onClick={startMatch} className="btn-primary text-xl px-8 py-4">
            Start Match vs AI
          </button>
        </div>
      )}

      {(matchState === 'playing' || matchState === 'finished') && (
        <>
          {/* Scoreboard */}
          <div className="card mb-6">
            <div className="flex justify-between items-center">
              <div className="text-center flex-1">
                <p className="text-xl font-bold">My Team FC</p>
                <p className="text-5xl font-bold text-primary mt-2">{homeScore}</p>
              </div>
              
              <div className="text-center px-8">
                <p className="text-gray-400 text-sm">
                  {matchState === 'finished' ? 'FULL TIME' : 'LIVE'}
                </p>
                <p className="text-3xl font-bold text-accent">{minute}'</p>
              </div>
              
              <div className="text-center flex-1">
                <p className="text-xl font-bold">AI United</p>
                <p className="text-5xl font-bold text-secondary mt-2">{awayScore}</p>
              </div>
            </div>

            {/* Possession bar */}
            <div className="mt-6">
              <div className="flex justify-between text-sm mb-1">
                <span>{possession}%</span>
                <span>Possession</span>
                <span>{100 - possession}%</span>
              </div>
              <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
                <div 
                  className="h-full bg-primary transition-all"
                  style={{ width: `${possession}%` }}
                />
              </div>
            </div>
          </div>

          {/* Match Events */}
          <div className="card">
            <h3 className="text-xl font-bold mb-4">Match Events</h3>
            <div className="space-y-2 max-h-64 overflow-y-auto">
              {events.length === 0 ? (
                <p className="text-gray-400">Waiting for action...</p>
              ) : (
                events.slice().reverse().map((event, i) => (
                  <div 
                    key={i}
                    className={`p-3 rounded-lg ${
                      event.team === 'home' ? 'bg-primary/10 border-l-4 border-primary' : 'bg-secondary/10 border-l-4 border-secondary'
                    }`}
                  >
                    <span className="mr-2">{getEventIcon(event.type)}</span>
                    <span className="font-bold">{event.minute}'</span>
                    <span className="ml-2">{event.description}</span>
                  </div>
                ))
              )}
            </div>
          </div>

          {matchState === 'finished' && (
            <div className="card mt-6 text-center">
              <h2 className="text-2xl font-bold mb-4">
                {homeScore > awayScore ? '🎉 Victory!' : homeScore < awayScore ? '😢 Defeat' : '🤝 Draw'}
              </h2>
              <button onClick={startMatch} className="btn-primary">
                Play Again
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
}

