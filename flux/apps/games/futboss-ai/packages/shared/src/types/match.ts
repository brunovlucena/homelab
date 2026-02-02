// FutBoss AI - Match Types
// Author: Bruno Lucena (bruno@lucena.cloud)

export type MatchStatus = 'pending' | 'in_progress' | 'finished' | 'cancelled';
export type EventType = 'goal' | 'assist' | 'yellow_card' | 'red_card' | 'substitution' | 'injury' | 'penalty' | 'save' | 'foul';

export interface MatchEvent {
  minute: number;
  eventType: EventType;
  playerId: string;
  teamId: string;
  description: string;
  aiNarration?: string;
}

export interface MatchState {
  minute: number;
  homeScore: number;
  awayScore: number;
  possessionHome: number;
  possessionAway: number;
  shotsHome: number;
  shotsAway: number;
  isPaused: boolean;
}

export interface Match {
  id: string;
  homeTeamId: string;
  awayTeamId: string;
  homeScore: number;
  awayScore: number;
  status: MatchStatus;
  events: MatchEvent[];
  state: MatchState;
  startedAt?: Date;
  finishedAt?: Date;
  createdAt: Date;
}

export interface MatchCreate {
  homeTeamId: string;
  awayTeamId: string;
}

export function isHomeWinner(match: Match): boolean | null {
  if (match.status !== 'finished') return null;
  if (match.homeScore === match.awayScore) return null;
  return match.homeScore > match.awayScore;
}

export function isDraw(match: Match): boolean {
  return match.status === 'finished' && match.homeScore === match.awayScore;
}

export function totalGoals(match: Match): number {
  return match.homeScore + match.awayScore;
}

