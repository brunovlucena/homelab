// FutBoss AI - Global Store (Zustand)
// Author: Bruno Lucena (bruno@lucena.cloud)

import { create } from 'zustand';
import type { User, Team, Player } from '@futboss/shared';

interface GameState {
  // Auth
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  
  // Game data
  team: Team | null;
  players: Player[];
  balance: number;
  
  // Actions
  setUser: (user: User | null, token?: string) => void;
  setTeam: (team: Team | null) => void;
  setPlayers: (players: Player[]) => void;
  setBalance: (balance: number) => void;
  logout: () => void;
}

export const useStore = create<GameState>((set) => ({
  // Initial state
  user: null,
  token: localStorage.getItem('token'),
  isAuthenticated: !!localStorage.getItem('token'),
  team: null,
  players: [],
  balance: 1000,
  
  // Actions
  setUser: (user, token) => {
    if (token) {
      localStorage.setItem('token', token);
    }
    set({ user, token, isAuthenticated: !!user });
  },
  
  setTeam: (team) => set({ team }),
  
  setPlayers: (players) => set({ players }),
  
  setBalance: (balance) => set({ balance }),
  
  logout: () => {
    localStorage.removeItem('token');
    set({
      user: null,
      token: null,
      isAuthenticated: false,
      team: null,
      players: [],
    });
  },
}));

