// FutBoss AI - Events Tests
// Author: Bruno Lucena (bruno@lucena.cloud)

import { describe, it, expect } from 'vitest';
import {
  getEventIcon,
  getEventImportance,
  generateBasicNarration,
  sortEventsByImportance,
  filterSignificantEvents,
  getMatchHighlights,
} from '../src/events';
import type { MatchEvent } from '@futboss/shared';

function createMockEvent(overrides: Partial<MatchEvent> = {}): MatchEvent {
  return {
    minute: 45,
    eventType: 'goal',
    playerId: 'p1',
    teamId: 't1',
    description: 'Test event',
    ...overrides,
  };
}

describe('getEventIcon', () => {
  it('should return goal icon', () => {
    expect(getEventIcon('goal')).toBe('⚽');
  });

  it('should return yellow card icon', () => {
    expect(getEventIcon('yellow_card')).toBe('🟨');
  });

  it('should return red card icon', () => {
    expect(getEventIcon('red_card')).toBe('🟥');
  });

  it('should return save icon', () => {
    expect(getEventIcon('save')).toBe('🧤');
  });
});

describe('getEventImportance', () => {
  it('should rate goal as most important', () => {
    expect(getEventImportance('goal')).toBe(10);
  });

  it('should rate red card as very important', () => {
    expect(getEventImportance('red_card')).toBe(8);
  });

  it('should rate foul as least important', () => {
    expect(getEventImportance('foul')).toBe(1);
  });
});

describe('generateBasicNarration', () => {
  it('should generate goal narration', () => {
    const event = createMockEvent({ eventType: 'goal', description: 'Player scores!' });
    const narration = generateBasicNarration(event);
    
    expect(narration).toContain('GOOOOL');
    expect(narration).toContain('Player scores!');
  });

  it('should generate save narration', () => {
    const event = createMockEvent({ eventType: 'save', description: 'Keeper saves' });
    const narration = generateBasicNarration(event);
    
    expect(narration).toContain('Great save');
  });

  it('should generate red card narration', () => {
    const event = createMockEvent({ eventType: 'red_card', description: 'Dangerous tackle' });
    const narration = generateBasicNarration(event);
    
    expect(narration).toContain('RED CARD');
  });
});

describe('sortEventsByImportance', () => {
  it('should sort events by importance descending', () => {
    const events: MatchEvent[] = [
      createMockEvent({ eventType: 'foul' }),
      createMockEvent({ eventType: 'goal' }),
      createMockEvent({ eventType: 'yellow_card' }),
    ];

    const sorted = sortEventsByImportance(events);

    expect(sorted[0].eventType).toBe('goal');
    expect(sorted[1].eventType).toBe('yellow_card');
    expect(sorted[2].eventType).toBe('foul');
  });

  it('should not mutate original array', () => {
    const events: MatchEvent[] = [
      createMockEvent({ eventType: 'foul' }),
      createMockEvent({ eventType: 'goal' }),
    ];

    sortEventsByImportance(events);

    expect(events[0].eventType).toBe('foul');
  });
});

describe('filterSignificantEvents', () => {
  it('should filter out low importance events', () => {
    const events: MatchEvent[] = [
      createMockEvent({ eventType: 'goal' }),
      createMockEvent({ eventType: 'foul' }),
      createMockEvent({ eventType: 'yellow_card' }),
      createMockEvent({ eventType: 'substitution' }),
    ];

    const filtered = filterSignificantEvents(events);

    expect(filtered.length).toBe(2);
    expect(filtered.some(e => e.eventType === 'goal')).toBe(true);
    expect(filtered.some(e => e.eventType === 'yellow_card')).toBe(true);
  });

  it('should allow custom minimum importance', () => {
    const events: MatchEvent[] = [
      createMockEvent({ eventType: 'goal' }),
      createMockEvent({ eventType: 'red_card' }),
      createMockEvent({ eventType: 'yellow_card' }),
    ];

    const filtered = filterSignificantEvents(events, 8);

    expect(filtered.length).toBe(2);
  });
});

describe('getMatchHighlights', () => {
  it('should return goals, cards, and saves sorted by minute', () => {
    const events: MatchEvent[] = [
      createMockEvent({ eventType: 'foul', minute: 10 }),
      createMockEvent({ eventType: 'goal', minute: 30 }),
      createMockEvent({ eventType: 'yellow_card', minute: 20 }),
      createMockEvent({ eventType: 'save', minute: 45 }),
      createMockEvent({ eventType: 'substitution', minute: 60 }),
    ];

    const highlights = getMatchHighlights(events);

    expect(highlights.length).toBe(3);
    expect(highlights[0].minute).toBe(20); // yellow card
    expect(highlights[1].minute).toBe(30); // goal
    expect(highlights[2].minute).toBe(45); // save
  });

  it('should limit saves to 3', () => {
    const events: MatchEvent[] = [
      createMockEvent({ eventType: 'save', minute: 10 }),
      createMockEvent({ eventType: 'save', minute: 20 }),
      createMockEvent({ eventType: 'save', minute: 30 }),
      createMockEvent({ eventType: 'save', minute: 40 }),
      createMockEvent({ eventType: 'save', minute: 50 }),
    ];

    const highlights = getMatchHighlights(events);

    expect(highlights.length).toBe(3);
  });
});

