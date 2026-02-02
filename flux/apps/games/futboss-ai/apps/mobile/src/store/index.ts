// FutBoss AI - Mobile Store (Zustand)
// Author: Bruno Lucena (bruno@lucena.cloud)

import { create } from 'zustand';
import type { User, Team, Player } from '@futboss/shared';

interface GameState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  team: Team | null;
  players: Player[];
  balance: number;
  
  setUser: (user: User | null, token?: string) => void;
  setTeam: (team: Team | null) => void;
  setPlayers: (players: Player[]) => void;
  setBalance: (balance: number) => void;
  logout: () => void;
}

export const useStore = create<GameState>((set) => ({
  user: null,
  token: null,
  isAuthenticated: false,
  team: null,
  players: [],
  balance: 1000,
  
  setUser: (user, token) => {
    set({ user, token, isAuthenticated: !!user });
  },
  
  setTeam: (team) => set({ team }),
  setPlayers: (players) => set({ players }),
  setBalance: (balance) => set({ balance }),
  
  logout: () => {
    set({
      user: null,
      token: null,
      isAuthenticated: false,
      team: null,
      players: [],
    });
  },
}));

