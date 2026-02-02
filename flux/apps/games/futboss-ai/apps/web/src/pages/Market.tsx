// FutBoss AI - Market Page
// Author: Bruno Lucena (bruno@lucena.cloud)

import { useState } from 'react';
import { useStore } from '../store';
import type { Player, Position } from '@futboss/shared';
import { calculateOverall } from '@futboss/shared';

// Mock market players
const marketPlayers: Player[] = [
  { id: 'm1', name: 'Ronaldo Silva', position: 'ST', nationality: 'Brasil', age: 23, attributes: { speed: 88, strength: 75, stamina: 82, finishing: 92, passing: 70, dribbling: 85, defense: 30, intelligence: 78, aggression: 65, leadership: 55, creativity: 80 }, personality: { temperament: 'explosive', playStyle: 'offensive' }, price: 1500, isListed: true, createdAt: new Date() },
  { id: 'm2', name: 'Casemiro Jr', position: 'CDM', nationality: 'Brasil', age: 25, attributes: { speed: 65, strength: 85, stamina: 88, finishing: 55, passing: 75, dribbling: 60, defense: 88, intelligence: 82, aggression: 75, leadership: 80, creativity: 55 }, personality: { temperament: 'calculated', playStyle: 'defensive' }, price: 1200, isListed: true, createdAt: new Date() },
  { id: 'm3', name: 'Raphinha Costa', position: 'RW', nationality: 'Brasil', age: 26, attributes: { speed: 88, strength: 60, stamina: 78, finishing: 78, passing: 75, dribbling: 85, defense: 40, intelligence: 75, aggression: 55, leadership: 50, creativity: 82 }, personality: { temperament: 'explosive', playStyle: 'offensive' }, price: 950, isListed: true, createdAt: new Date() },
  { id: 'm4', name: 'Alisson Becker', position: 'GK', nationality: 'Brasil', age: 30, attributes: { speed: 50, strength: 70, stamina: 75, finishing: 25, passing: 45, dribbling: 30, defense: 90, intelligence: 85, aggression: 35, leadership: 75, creativity: 40 }, personality: { temperament: 'calm', playStyle: 'defensive' }, price: 1400, isListed: true, createdAt: new Date() },
  { id: 'm5', name: 'Endrick Felipe', position: 'ST', nationality: 'Brasil', age: 18, attributes: { speed: 85, strength: 65, stamina: 78, finishing: 82, passing: 55, dribbling: 78, defense: 25, intelligence: 68, aggression: 70, leadership: 45, creativity: 75 }, personality: { temperament: 'explosive', playStyle: 'offensive' }, price: 800, isListed: true, createdAt: new Date() },
];

const positions: Position[] = ['GK', 'CB', 'LB', 'RB', 'CDM', 'CM', 'CAM', 'LW', 'RW', 'ST'];

export default function Market() {
  const { balance } = useStore();
  const [filter, setFilter] = useState<Position | 'all'>('all');
  const [selectedPlayer, setSelectedPlayer] = useState<Player | null>(null);

  const filteredPlayers = filter === 'all' 
    ? marketPlayers 
    : marketPlayers.filter(p => p.position === filter);

  const handleBuy = (player: Player) => {
    if (balance < player.price) {
      alert('Insufficient balance!');
      return;
    }
    alert(`Buying ${player.name} for ${player.price} FTC...`);
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">🏪 Transfer Market</h1>
        <div className="text-xl">
          Balance: <span className="text-accent font-bold">{balance.toLocaleString()} FTC</span>
        </div>
      </div>

      {/* Filters */}
      <div className="card mb-6">
        <div className="flex flex-wrap gap-2">
          <button
            onClick={() => setFilter('all')}
            className={`px-4 py-2 rounded-lg ${
              filter === 'all' ? 'bg-primary text-dark' : 'bg-dark hover:bg-primary/20'
            }`}
          >
            All
          </button>
          {positions.map((pos) => (
            <button
              key={pos}
              onClick={() => setFilter(pos)}
              className={`px-4 py-2 rounded-lg ${
                filter === pos ? 'bg-primary text-dark' : 'bg-dark hover:bg-primary/20'
              }`}
            >
              {pos}
            </button>
          ))}
        </div>
      </div>

      {/* Player Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredPlayers.map((player) => (
          <div
            key={player.id}
            className={`card cursor-pointer transition-all hover:border-primary ${
              selectedPlayer?.id === player.id ? 'border-primary' : ''
            }`}
            onClick={() => setSelectedPlayer(player)}
          >
            <div className="flex justify-between items-start mb-3">
              <div>
                <h3 className="font-bold text-lg">{player.name}</h3>
                <p className="text-gray-400">{player.nationality} | {player.age} yrs</p>
              </div>
              <span className="text-xl font-bold bg-secondary/20 text-secondary px-2 py-1 rounded">
                {calculateOverall(player.attributes)}
              </span>
            </div>

            <div className="flex justify-between items-center">
              <span className="text-primary font-medium">{player.position}</span>
              <span className="text-accent font-bold">{player.price} FTC</span>
            </div>

            {selectedPlayer?.id === player.id && (
              <button
                onClick={() => handleBuy(player)}
                className="btn-primary w-full mt-4"
                disabled={balance < player.price}
              >
                {balance >= player.price ? 'Buy Player' : 'Insufficient Funds'}
              </button>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

