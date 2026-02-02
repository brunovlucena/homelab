// FutBoss AI - Match Simulation Engine
// Author: Bruno Lucena (bruno@lucena.cloud)

import type { Player, PlayerAttributes, Match, MatchState, MatchEvent, EventType } from '@futboss/shared';

export type AgentAction = 'PASS' | 'DRIBBLE' | 'SHOOT' | 'TACKLE' | 'HOLD' | 'RUN_FORWARD' | 'FALL_BACK';
export type BallZone = 'defense' | 'midfield' | 'attack';

export interface AgentDecision {
  action: AgentAction;
  reason: string;
}

export interface SimulationConfig {
  matchDuration: number;
  tickInterval: number;
  goalProbabilityBase: number;
}

export const defaultConfig: SimulationConfig = {
  matchDuration: 90,
  matchDuration: 90,
  tickInterval: 1000,
  goalProbabilityBase: 0.02,
};

export class MatchSimulation {
  private state: MatchState;
  private homeTeam: Player[];
  private awayTeam: Player[];
  private config: SimulationConfig;
  private events: MatchEvent[] = [];
  private ballZone: BallZone = 'midfield';
  private ballPossession: 'home' | 'away' = 'home';

  constructor(
    homeTeam: Player[],
    awayTeam: Player[],
    config: Partial<SimulationConfig> = {}
  ) {
    this.homeTeam = homeTeam;
    this.awayTeam = awayTeam;
    this.config = { ...defaultConfig, ...config };
    this.state = this.createInitialState();
  }

  private createInitialState(): MatchState {
    return {
      minute: 0,
      homeScore: 0,
      awayScore: 0,
      possessionHome: 50,
      possessionAway: 50,
      shotsHome: 0,
      shotsAway: 0,
      isPaused: false,
    };
  }

  getState(): MatchState {
    return { ...this.state };
  }

  getEvents(): MatchEvent[] {
    return [...this.events];
  }

  getBallZone(): BallZone {
    return this.ballZone;
  }

  getBallPossession(): 'home' | 'away' {
    return this.ballPossession;
  }

  simulateMinute(): MatchEvent | null {
    if (this.state.minute >= this.config.matchDuration) {
      return null;
    }

    this.state.minute++;
    
    // Get active player (simplified - random from team with ball)
    const team = this.ballPossession === 'home' ? this.homeTeam : this.awayTeam;
    const player = team[Math.floor(Math.random() * team.length)];
    
    // Simulate action outcome
    const event = this.processAction(player);
    
    // Update possession stats
    this.updatePossessionStats();
    
    return event;
  }

  private processAction(player: Player): MatchEvent | null {
    const attrs = player.attributes;
    
    // Determine action based on zone and attributes
    if (this.ballZone === 'attack') {
      return this.processAttackAction(player);
    } else if (this.ballZone === 'midfield') {
      return this.processMidfieldAction(player);
    } else {
      return this.processDefenseAction(player);
    }
  }

  private processAttackAction(player: Player): MatchEvent | null {
    const shotChance = (player.attributes.finishing / 100) * 0.3;
    
    if (Math.random() < shotChance) {
      // Shot!
      const isHome = this.ballPossession === 'home';
      if (isHome) this.state.shotsHome++;
      else this.state.shotsAway++;
      
      const goalChance = this.calculateGoalProbability(player);
      
      if (Math.random() < goalChance) {
        // GOAL!
        if (isHome) this.state.homeScore++;
        else this.state.awayScore++;
        
        this.ballZone = 'midfield';
        this.ballPossession = isHome ? 'away' : 'home';
        
        const event: MatchEvent = {
          minute: this.state.minute,
          eventType: 'goal',
          playerId: player.id,
          teamId: player.teamId || '',
          description: `${player.name} scores!`,
        };
        this.events.push(event);
        return event;
      } else {
        // Save or miss
        this.ballPossession = this.ballPossession === 'home' ? 'away' : 'home';
        this.ballZone = 'defense';
        
        const event: MatchEvent = {
          minute: this.state.minute,
          eventType: 'save',
          playerId: player.id,
          teamId: player.teamId || '',
          description: `${player.name}'s shot saved!`,
        };
        this.events.push(event);
        return event;
      }
    }
    
    // Pass or dribble
    if (Math.random() < 0.5) {
      this.ballZone = Math.random() < 0.3 ? 'midfield' : 'attack';
    }
    
    return null;
  }

  private processMidfieldAction(player: Player): MatchEvent | null {
    const progressChance = (player.attributes.passing + player.attributes.dribbling) / 200;
    
    if (Math.random() < progressChance) {
      // Advance
      this.ballZone = 'attack';
    } else if (Math.random() < 0.2) {
      // Lose possession
      this.ballPossession = this.ballPossession === 'home' ? 'away' : 'home';
      this.ballZone = 'midfield';
    }
    
    // Random foul chance
    if (Math.random() < 0.05) {
      const event: MatchEvent = {
        minute: this.state.minute,
        eventType: 'foul',
        playerId: player.id,
        teamId: player.teamId || '',
        description: `Foul by ${player.name}`,
      };
      this.events.push(event);
      
      // Yellow card chance
      if (Math.random() < 0.2) {
        const cardEvent: MatchEvent = {
          minute: this.state.minute,
          eventType: 'yellow_card',
          playerId: player.id,
          teamId: player.teamId || '',
          description: `Yellow card for ${player.name}`,
        };
        this.events.push(cardEvent);
        return cardEvent;
      }
      return event;
    }
    
    return null;
  }

  private processDefenseAction(player: Player): MatchEvent | null {
    const clearChance = player.attributes.defense / 100;
    
    if (Math.random() < clearChance) {
      // Clear the ball
      this.ballZone = 'midfield';
    } else {
      // Opponent advances
      this.ballPossession = this.ballPossession === 'home' ? 'away' : 'home';
      this.ballZone = 'attack';
    }
    
    return null;
  }

  private calculateGoalProbability(player: Player): number {
    const base = this.config.goalProbabilityBase;
    const finishingBonus = (player.attributes.finishing - 50) / 200;
    const intelligenceBonus = (player.attributes.intelligence - 50) / 400;
    
    return Math.max(0.01, Math.min(0.5, base + finishingBonus + intelligenceBonus));
  }

  private updatePossessionStats(): void {
    const totalMinutes = this.state.minute;
    if (totalMinutes === 0) return;
    
    // Simple possession tracking
    if (this.ballPossession === 'home') {
      this.state.possessionHome = Math.min(70, this.state.possessionHome + 1);
      this.state.possessionAway = 100 - this.state.possessionHome;
    } else {
      this.state.possessionAway = Math.min(70, this.state.possessionAway + 1);
      this.state.possessionHome = 100 - this.state.possessionAway;
    }
  }

  isFinished(): boolean {
    return this.state.minute >= this.config.matchDuration;
  }
}

