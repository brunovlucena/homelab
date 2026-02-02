// FutBoss AI - Simulation Tests
// Author: Bruno Lucena (bruno@lucena.cloud)

import { describe, it, expect, beforeEach } from 'vitest';
import { MatchSimulation, defaultConfig } from '../src/simulation';
import type { Player } from '@futboss/shared';

function createMockPlayer(overrides: Partial<Player> = {}): Player {
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
    },
    personality: {
      temperament: 'calm',
      playStyle: 'balanced',
    },
    price: 500,
    isListed: false,
    createdAt: new Date(),
    ...overrides,
  };
}

describe('MatchSimulation', () => {
  let simulation: MatchSimulation;
  let homeTeam: Player[];
  let awayTeam: Player[];

  beforeEach(() => {
    homeTeam = [
      createMockPlayer({ id: 'h1', name: 'Home Player 1', teamId: 'home' }),
      createMockPlayer({ id: 'h2', name: 'Home Player 2', teamId: 'home' }),
    ];
    awayTeam = [
      createMockPlayer({ id: 'a1', name: 'Away Player 1', teamId: 'away' }),
      createMockPlayer({ id: 'a2', name: 'Away Player 2', teamId: 'away' }),
    ];
    simulation = new MatchSimulation(homeTeam, awayTeam);
  });

  it('should create initial state correctly', () => {
    const state = simulation.getState();
    
    expect(state.minute).toBe(0);
    expect(state.homeScore).toBe(0);
    expect(state.awayScore).toBe(0);
    expect(state.possessionHome).toBe(50);
    expect(state.possessionAway).toBe(50);
  });

  it('should increment minute on simulation', () => {
    simulation.simulateMinute();
    expect(simulation.getState().minute).toBe(1);
    
    simulation.simulateMinute();
    expect(simulation.getState().minute).toBe(2);
  });

  it('should track events', () => {
    // Run several minutes to potentially generate events
    for (let i = 0; i < 30; i++) {
      simulation.simulateMinute();
    }
    
    const events = simulation.getEvents();
    expect(Array.isArray(events)).toBe(true);
  });

  it('should finish after match duration', () => {
    expect(simulation.isFinished()).toBe(false);
    
    // Simulate full match
    for (let i = 0; i < 90; i++) {
      simulation.simulateMinute();
    }
    
    expect(simulation.isFinished()).toBe(true);
  });

  it('should not simulate after match is finished', () => {
    // Complete the match
    for (let i = 0; i < 90; i++) {
      simulation.simulateMinute();
    }
    
    const result = simulation.simulateMinute();
    expect(result).toBeNull();
    expect(simulation.getState().minute).toBe(90);
  });

  it('should return ball zone', () => {
    const zone = simulation.getBallZone();
    expect(['defense', 'midfield', 'attack']).toContain(zone);
  });

  it('should return ball possession', () => {
    const possession = simulation.getBallPossession();
    expect(['home', 'away']).toContain(possession);
  });

  it('should use custom config', () => {
    const customSim = new MatchSimulation(homeTeam, awayTeam, {
      matchDuration: 45,
    });
    
    for (let i = 0; i < 45; i++) {
      customSim.simulateMinute();
    }
    
    expect(customSim.isFinished()).toBe(true);
  });
});

