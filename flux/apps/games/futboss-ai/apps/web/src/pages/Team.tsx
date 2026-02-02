// FutBoss AI - Team Page
// Author: Bruno Lucena (bruno@lucena.cloud)

import { useState } from 'react';
import { useStore } from '../store';
import type { Player, Position } from '@futboss/shared';
import { calculateOverall } from '@futboss/shared';

// Mock players for demo
const mockPlayers: Player[] = [
  { id: '1', name: 'Roberto Carlos', position: 'GK', nationality: 'Brasil', age: 28, attributes: { speed: 45, strength: 65, stamina: 70, finishing: 20, passing: 65, dribbling: 35, defense: 85, intelligence: 70, aggression: 40, leadership: 60, creativity: 30 }, personality: { temperament: 'calm', playStyle: 'defensive' }, price: 500, isListed: false, createdAt: new Date() },
  { id: '2', name: 'Marcelo Silva', position: 'CB', nationality: 'Brasil', age: 26, attributes: { speed: 60, strength: 80, stamina: 75, finishing: 30, passing: 55, dribbling: 40, defense: 82, intelligence: 68, aggression: 65, leadership: 70, creativity: 35 }, personality: { temperament: 'calculated', playStyle: 'defensive' }, price: 450, isListed: false, createdAt: new Date() },
  { id: '3', name: 'Bruno Fernandes', position: 'CM', nationality: 'Portugal', age: 27, attributes: { speed: 70, strength: 65, stamina: 80, finishing: 75, passing: 88, dribbling: 78, defense: 55, intelligence: 85, aggression: 55, leadership: 80, creativity: 88 }, personality: { temperament: 'calculated', playStyle: 'offensive' }, price: 800, isListed: false, createdAt: new Date() },
  { id: '4', name: 'Neymar Jr', position: 'LW', nationality: 'Brasil', age: 29, attributes: { speed: 90, strength: 55, stamina: 75, finishing: 85, passing: 80, dribbling: 95, defense: 30, intelligence: 82, aggression: 45, leadership: 60, creativity: 95 }, personality: { temperament: 'explosive', playStyle: 'offensive' }, price: 1500, isListed: false, createdAt: new Date() },
  { id: '5', name: 'Pedro Striker', position: 'ST', nationality: 'Brasil', age: 24, attributes: { speed: 82, strength: 70, stamina: 78, finishing: 90, passing: 65, dribbling: 75, defense: 25, intelligence: 72, aggression: 70, leadership: 55, creativity: 70 }, personality: { temperament: 'explosive', playStyle: 'offensive' }, price: 1200, isListed: false, createdAt: new Date() },
];

function getPositionColor(position: Position): string {
  if (['GK'].includes(position)) return 'text-yellow-400';
  if (['CB', 'LB', 'RB'].includes(position)) return 'text-blue-400';
  if (['CDM', 'CM', 'CAM'].includes(position)) return 'text-green-400';
  return 'text-red-400';
}

function AttrBar({ value }: { value: number }) {
  const color = value >= 80 ? 'bg-green-500' : value >= 60 ? 'bg-yellow-500' : 'bg-red-500';
  return (
    <div className="attr-bar">
      <div className={`attr-bar-fill ${color}`} style={{ width: `${value}%` }} />
    </div>
  );
}

export default function Team() {
  const [selectedPlayer, setSelectedPlayer] = useState<Player | null>(null);
  const players = mockPlayers;

  return (
    <div>
      <h1 className="text-3xl font-bold mb-8">⚽ Team Management</h1>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Player List */}
        <div className="lg:col-span-2 card">
          <h2 className="text-xl font-bold mb-4">Squad ({players.length} players)</h2>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="text-left text-gray-400 border-b border-gray-700">
                  <th className="pb-2">Name</th>
                  <th className="pb-2">Pos</th>
                  <th className="pb-2">Age</th>
                  <th className="pb-2">OVR</th>
                  <th className="pb-2">Value</th>
                </tr>
              </thead>
              <tbody>
                {players.map((player) => (
                  <tr
                    key={player.id}
                    className={`border-b border-gray-800 cursor-pointer hover:bg-primary/10 ${
                      selectedPlayer?.id === player.id ? 'bg-primary/20' : ''
                    }`}
                    onClick={() => setSelectedPlayer(player)}
                  >
                    <td className="py-3 font-medium">{player.name}</td>
                    <td className={`py-3 ${getPositionColor(player.position)}`}>{player.position}</td>
                    <td className="py-3">{player.age}</td>
                    <td className="py-3 font-bold">{calculateOverall(player.attributes)}</td>
                    <td className="py-3 text-accent">{player.price} FTC</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Player Details */}
        <div className="card">
          {selectedPlayer ? (
            <>
              <h2 className="text-xl font-bold mb-4">{selectedPlayer.name}</h2>
              <div className="mb-4">
                <span className={`${getPositionColor(selectedPlayer.position)} font-bold`}>
                  {selectedPlayer.position}
                </span>
                <span className="text-gray-400 ml-2">| {selectedPlayer.nationality}</span>
              </div>

              <div className="mb-4">
                <p className="text-gray-400 text-sm">Overall</p>
                <p className="text-3xl font-bold text-primary">
                  {calculateOverall(selectedPlayer.attributes)}
                </p>
              </div>

              <div className="space-y-3 mb-6">
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span>Speed</span>
                    <span>{selectedPlayer.attributes.speed}</span>
                  </div>
                  <AttrBar value={selectedPlayer.attributes.speed} />
                </div>
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span>Finishing</span>
                    <span>{selectedPlayer.attributes.finishing}</span>
                  </div>
                  <AttrBar value={selectedPlayer.attributes.finishing} />
                </div>
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span>Passing</span>
                    <span>{selectedPlayer.attributes.passing}</span>
                  </div>
                  <AttrBar value={selectedPlayer.attributes.passing} />
                </div>
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span>Defense</span>
                    <span>{selectedPlayer.attributes.defense}</span>
                  </div>
                  <AttrBar value={selectedPlayer.attributes.defense} />
                </div>
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span>Creativity</span>
                    <span>{selectedPlayer.attributes.creativity}</span>
                  </div>
                  <AttrBar value={selectedPlayer.attributes.creativity} />
                </div>
              </div>

              <div className="mb-4 p-3 bg-dark rounded-lg">
                <p className="text-gray-400 text-sm">Personality</p>
                <p className="capitalize">
                  {selectedPlayer.personality.temperament} / {selectedPlayer.personality.playStyle}
                </p>
              </div>

              <button className="btn-secondary w-full">List for Sale</button>
            </>
          ) : (
            <div className="text-center text-gray-400 py-8">
              Select a player to view details
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

