// FutBoss AI - Player Types
// Author: Bruno Lucena (bruno@lucena.cloud)

export type Position = 'GK' | 'CB' | 'LB' | 'RB' | 'CDM' | 'CM' | 'CAM' | 'LW' | 'RW' | 'ST';
export type Temperament = 'calm' | 'explosive' | 'calculated';
export type PlayStyle = 'offensive' | 'defensive' | 'balanced';

export interface PlayerAttributes {
  speed: number;
  strength: number;
  stamina: number;
  finishing: number;
  passing: number;
  dribbling: number;
  defense: number;
  intelligence: number;
  aggression: number;
  leadership: number;
  creativity: number;
}

export interface PlayerPersonality {
  temperament: Temperament;
  playStyle: PlayStyle;
}

export interface Player {
  id: string;
  name: string;
  position: Position;
  nationality: string;
  age: number;
  attributes: PlayerAttributes;
  personality: PlayerPersonality;
  teamId?: string;
  price: number;
  isListed: boolean;
  createdAt: Date;
}

export interface PlayerCreate {
  name: string;
  position: Position;
  nationality: string;
  age: number;
  attributes?: Partial<PlayerAttributes>;
  personality?: Partial<PlayerPersonality>;
}

export function calculateOverall(attrs: PlayerAttributes): number {
  const values = Object.values(attrs);
  return Math.round(values.reduce((a, b) => a + b, 0) / values.length);
}

export function calculateMarketValue(player: Player): number {
  const baseValue = calculateOverall(player.attributes) * 10;
  let ageModifier = 1.0;
  
  if (player.age < 23) {
    ageModifier = 1.3;
  } else if (player.age > 32) {
    ageModifier = 0.7;
  }
  
  return Math.round(baseValue * ageModifier);
}

export function isDefender(position: Position): boolean {
  return ['CB', 'LB', 'RB'].includes(position);
}

export function isMidfielder(position: Position): boolean {
  return ['CDM', 'CM', 'CAM'].includes(position);
}

export function isAttacker(position: Position): boolean {
  return ['LW', 'RW', 'ST'].includes(position);
}

