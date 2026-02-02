// FutBoss AI - Match Events
// Author: Bruno Lucena (bruno@lucena.cloud)

import type { MatchEvent, EventType } from '@futboss/shared';

export interface EventGenerator {
  generateNarration(event: MatchEvent): string;
  getEventIcon(eventType: EventType): string;
  getEventImportance(eventType: EventType): number;
}

export const eventIcons: Record<EventType, string> = {
  goal: '⚽',
  assist: '🅰️',
  yellow_card: '🟨',
  red_card: '🟥',
  substitution: '🔄',
  injury: '🏥',
  penalty: '⚽🎯',
  save: '🧤',
  foul: '❌',
};

export function getEventIcon(eventType: EventType): string {
  return eventIcons[eventType] || '📢';
}

export function getEventImportance(eventType: EventType): number {
  const importance: Record<EventType, number> = {
    goal: 10,
    red_card: 8,
    penalty: 7,
    yellow_card: 4,
    save: 5,
    assist: 6,
    injury: 6,
    substitution: 2,
    foul: 1,
  };
  
  return importance[eventType] || 1;
}

export function generateBasicNarration(event: MatchEvent): string {
  const templates: Record<EventType, (e: MatchEvent) => string> = {
    goal: (e) => `GOOOOL! ${e.description}`,
    assist: (e) => `Beautiful assist! ${e.description}`,
    yellow_card: (e) => `Yellow card shown. ${e.description}`,
    red_card: (e) => `RED CARD! Player sent off! ${e.description}`,
    substitution: (e) => `Substitution made. ${e.description}`,
    injury: (e) => `Injury concern. ${e.description}`,
    penalty: (e) => `PENALTY! ${e.description}`,
    save: (e) => `Great save! ${e.description}`,
    foul: (e) => `Foul called. ${e.description}`,
  };

  const template = templates[event.eventType];
  return template ? template(event) : event.description;
}

export function sortEventsByImportance(events: MatchEvent[]): MatchEvent[] {
  return [...events].sort((a, b) => {
    const importanceA = getEventImportance(a.eventType);
    const importanceB = getEventImportance(b.eventType);
    return importanceB - importanceA;
  });
}

export function filterSignificantEvents(events: MatchEvent[], minImportance: number = 4): MatchEvent[] {
  return events.filter(e => getEventImportance(e.eventType) >= minImportance);
}

export function getMatchHighlights(events: MatchEvent[]): MatchEvent[] {
  const goals = events.filter(e => e.eventType === 'goal');
  const cards = events.filter(e => e.eventType === 'red_card' || e.eventType === 'yellow_card');
  const saves = events.filter(e => e.eventType === 'save').slice(0, 3);
  
  return [...goals, ...cards, ...saves].sort((a, b) => a.minute - b.minute);
}

