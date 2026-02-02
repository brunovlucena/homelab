// FutBoss AI - Actions Tests
// Author: Bruno Lucena (bruno@lucena.cloud)

import { describe, it, expect } from 'vitest';
import { decideAction, calculateActionSuccess, getActionRisk } from '../src/actions';
import type { Player, PlayerAttributes } from '@futboss/shared';

function createMockPlayer(attrsOverride: Partial<PlayerAttributes> = {}): Player {
  return {
    id: '1',
    name: 'Test Player',
    position: 'CM',
    nationality: 'Brasil',
    age: 25,
    attributes: {
      speed: 70,
      strength: 70,
      stamina: 70,
      finishing: 70,
      passing: 70,
      dribbling: 70,
      defense: 70,
      intelligence: 70,
      aggression: 50,
      leadership: 50,
      creativity: 70,
      ...attrsOverride,
    },
    personality: {
      temperament: 'calm',
      playStyle: 'balanced',
    },
    price: 500,
    isListed: false,
    createdAt: new Date(),
  };
}

describe('decideAction', () => {
  it('should tackle when defender without ball', () => {
    const player = createMockPlayer({ defense: 85 });
    const result = decideAction({
      player,
      hasBall: false,
      ballZone: 'midfield',
      minute: 30,
      scoreDiff: 0,
    });
    
    expect(result.action).toBe('TACKLE');
  });

  it('should fall back in defense zone without ball', () => {
    const player = createMockPlayer({ defense: 50 });
    const result = decideAction({
      player,
      hasBall: false,
      ballZone: 'defense',
      minute: 30,
      scoreDiff: 0,
    });
    
    expect(result.action).toBe('FALL_BACK');
  });

  it('should shoot when good finisher in attack', () => {
    const player = createMockPlayer({ finishing: 85 });
    const result = decideAction({
      player,
      hasBall: true,
      ballZone: 'attack',
      minute: 30,
      scoreDiff: 0,
    });
    
    expect(result.action).toBe('SHOOT');
  });

  it('should pass when good passer in midfield', () => {
    const player = createMockPlayer({ passing: 85, dribbling: 50 });
    const result = decideAction({
      player,
      hasBall: true,
      ballZone: 'midfield',
      minute: 30,
      scoreDiff: 0,
    });
    
    expect(result.action).toBe('PASS');
  });

  it('should dribble when skilled dribbler with offensive style', () => {
    const player = createMockPlayer({ dribbling: 85, passing: 60 });
    player.personality.playStyle = 'offensive';
    
    const result = decideAction({
      player,
      hasBall: true,
      ballZone: 'midfield',
      minute: 30,
      scoreDiff: 0,
    });
    
    expect(result.action).toBe('DRIBBLE');
  });
});

describe('calculateActionSuccess', () => {
  it('should calculate pass success based on passing and intelligence', () => {
    const attrs = createMockPlayer({ passing: 80, intelligence: 80 }).attributes;
    const success = calculateActionSuccess('PASS', attrs);
    
    expect(success).toBe(0.8); // (80 + 80) / 200
  });

  it('should calculate shoot success based on finishing and intelligence', () => {
    const attrs = createMockPlayer({ finishing: 90, intelligence: 70 }).attributes;
    const success = calculateActionSuccess('SHOOT', attrs);
    
    expect(success).toBe(0.8); // (90 + 70) / 200
  });

  it('should calculate tackle success based on defense and strength', () => {
    const attrs = createMockPlayer({ defense: 85, strength: 75 }).attributes;
    const success = calculateActionSuccess('TACKLE', attrs);
    
    expect(success).toBe(0.8); // (85 + 75) / 200
  });

  it('should have base success for HOLD action', () => {
    const attrs = createMockPlayer().attributes;
    const success = calculateActionSuccess('HOLD', attrs);
    
    expect(success).toBe(0.7);
  });
});

describe('getActionRisk', () => {
  it('should return high risk for SHOOT', () => {
    expect(getActionRisk('SHOOT')).toBe(0.7);
  });

  it('should return medium risk for DRIBBLE', () => {
    expect(getActionRisk('DRIBBLE')).toBe(0.5);
  });

  it('should return low risk for HOLD', () => {
    expect(getActionRisk('HOLD')).toBe(0.1);
  });

  it('should return low risk for FALL_BACK', () => {
    expect(getActionRisk('FALL_BACK')).toBe(0.1);
  });
});

