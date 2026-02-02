// FutBoss AI - Player Actions
// Author: Bruno Lucena (bruno@lucena.cloud)

import type { Player, PlayerAttributes } from '@futboss/shared';
import type { AgentAction, AgentDecision, BallZone } from './simulation';

export interface ActionContext {
  player: Player;
  hasBall: boolean;
  ballZone: BallZone;
  minute: number;
  scoreDiff: number; // Positive = winning, negative = losing
}

export function decideAction(ctx: ActionContext): AgentDecision {
  const { player, hasBall, ballZone, scoreDiff } = ctx;
  const attrs = player.attributes;
  const personality = player.personality;

  if (!hasBall) {
    return decideDefensiveAction(attrs, ballZone);
  }

  return decideOffensiveAction(attrs, ballZone, personality, scoreDiff);
}

function decideDefensiveAction(attrs: PlayerAttributes, zone: BallZone): AgentDecision {
  if (attrs.defense > 70) {
    return { action: 'TACKLE', reason: 'Strong defensive ability' };
  }
  
  if (zone === 'defense') {
    return { action: 'FALL_BACK', reason: 'Protecting goal' };
  }
  
  return { action: 'HOLD', reason: 'Maintaining position' };
}

function decideOffensiveAction(
  attrs: PlayerAttributes,
  zone: BallZone,
  personality: { temperament: string; playStyle: string },
  scoreDiff: number
): AgentDecision {
  // In attack zone
  if (zone === 'attack') {
    if (attrs.finishing > 75) {
      return { action: 'SHOOT', reason: 'Good shooting position' };
    }
    if (attrs.creativity > 70 && personality.temperament === 'explosive') {
      return { action: 'DRIBBLE', reason: 'Creating chance with skill' };
    }
    if (attrs.passing > attrs.dribbling) {
      return { action: 'PASS', reason: 'Better passing option' };
    }
  }

  // In midfield
  if (zone === 'midfield') {
    if (attrs.dribbling > 75 && personality.playStyle === 'offensive') {
      return { action: 'DRIBBLE', reason: 'Advancing with ball' };
    }
    if (attrs.passing > 70) {
      return { action: 'PASS', reason: 'Building play' };
    }
    return { action: 'RUN_FORWARD', reason: 'Moving up the pitch' };
  }

  // In defense zone with ball
  if (attrs.passing > 60) {
    return { action: 'PASS', reason: 'Clearing danger' };
  }
  
  return { action: 'HOLD', reason: 'Keeping possession' };
}

export function calculateActionSuccess(
  action: AgentAction,
  attrs: PlayerAttributes
): number {
  const successRates: Record<AgentAction, (a: PlayerAttributes) => number> = {
    PASS: (a) => (a.passing + a.intelligence) / 200,
    DRIBBLE: (a) => (a.dribbling + a.speed) / 200,
    SHOOT: (a) => (a.finishing + a.intelligence) / 200,
    TACKLE: (a) => (a.defense + a.strength) / 200,
    HOLD: () => 0.7, // Base success rate
    RUN_FORWARD: (a) => (a.speed + a.stamina) / 200,
    FALL_BACK: (a) => (a.defense + a.intelligence) / 200,
  };

  return successRates[action](attrs);
}

export function getActionRisk(action: AgentAction): number {
  const riskLevels: Record<AgentAction, number> = {
    SHOOT: 0.7,
    DRIBBLE: 0.5,
    TACKLE: 0.4,
    PASS: 0.2,
    RUN_FORWARD: 0.3,
    HOLD: 0.1,
    FALL_BACK: 0.1,
  };
  
  return riskLevels[action];
}

