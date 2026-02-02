export interface Score {
  userId: string;
  sessionId: string;
  points: number;
  transitions: number;
  perfectSyncs: number;
  duration: number; // seconds
  timestamp: number;
}

export interface Achievement {
  id: string;
  name: string;
  description: string;
  icon: string;
  points: number;
  unlockedAt?: number;
}

export interface UserStats {
  userId: string;
  totalPoints: number;
  totalSessions: number;
  totalDuration: number; // seconds
  achievements: Achievement[];
  rank: number;
  level: number;
  xp: number;
}

export interface LeaderboardEntry {
  userId: string;
  username: string;
  avatar?: string;
  totalPoints: number;
  rank: number;
  level: number;
}
