// FutBoss AI - Dashboard Page
// Author: Bruno Lucena (bruno@lucena.cloud)

import { useStore } from '../store';
import { Link } from 'react-router-dom';

export default function Dashboard() {
  const { team, balance } = useStore();

  return (
    <div>
      <h1 className="text-3xl font-bold mb-8">Welcome to FutBoss AI</h1>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="card">
          <p className="text-gray-400 text-sm">Team</p>
          <p className="text-2xl font-bold">{team?.name || 'Create a team'}</p>
          {team && (
            <p className="text-primary mt-2">
              {team.wins}W - {team.draws}D - {team.losses}L
            </p>
          )}
        </div>
        
        <div className="card">
          <p className="text-gray-400 text-sm">Balance</p>
          <p className="text-2xl font-bold text-accent">{balance.toLocaleString()} FTC</p>
          <Link to="/wallet" className="text-primary text-sm hover:underline">
            Buy more tokens →
          </Link>
        </div>
        
        <div className="card">
          <p className="text-gray-400 text-sm">Ranking</p>
          <p className="text-2xl font-bold">#--</p>
          <p className="text-gray-400 text-sm">Play matches to rank up</p>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="card mb-8">
        <h2 className="text-xl font-bold mb-4">Quick Actions</h2>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <Link to="/team" className="btn-primary text-center">
            👥 Manage Team
          </Link>
          <Link to="/market" className="btn-secondary text-center">
            🏪 Transfer Market
          </Link>
          <Link to="/match" className="btn-primary text-center">
            🎮 New Match
          </Link>
          <Link to="/wallet" className="btn-secondary text-center">
            💰 Buy Tokens
          </Link>
        </div>
      </div>

      {/* Info */}
      <div className="card">
        <h2 className="text-xl font-bold mb-4">How to Play</h2>
        <ul className="space-y-2 text-gray-300">
          <li>⚽ Build your dream team by purchasing players from the market</li>
          <li>🤖 Each player has unique AI attributes that affect their decisions</li>
          <li>🎮 Challenge other players in real-time multiplayer matches</li>
          <li>💰 Win matches to earn FutCoins and climb the rankings</li>
          <li>🧠 AI agents run locally on Ollama for smart decision-making</li>
        </ul>
      </div>
    </div>
  );
}

