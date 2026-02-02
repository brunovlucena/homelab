// FutBoss AI - Team Types
// Author: Bruno Lucena (bruno@lucena.cloud)

export type Formation = '4-4-2' | '4-3-3' | '3-5-2' | '4-5-1' | '3-4-3' | '5-3-2';

export interface Team {
  id: string;
  name: string;
  ownerId: string;
  formation: Formation;
  playerIds: string[];
  wins: number;
  draws: number;
  losses: number;
  goalsFor: number;
  goalsAgainst: number;
  createdAt: Date;
}

export interface TeamCreate {
  name: string;
  formation?: Formation;
}

export interface TeamResponse extends Team {
  points: number;
  goalDifference: number;
}

export function calculatePoints(team: Team): number {
  return team.wins * 3 + team.draws;
}

export function calculateGoalDifference(team: Team): number {
  return team.goalsFor - team.goalsAgainst;
}

export function calculateMatchesPlayed(team: Team): number {
  return team.wins + team.draws + team.losses;
}

export function calculateWinRate(team: Team): number {
  const matches = calculateMatchesPlayed(team);
  if (matches === 0) return 0;
  return (team.wins / matches) * 100;
}

