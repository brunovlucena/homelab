export interface DJ {
  id: string;
  username: string;
  avatar?: string;
  isHost: boolean;
  connectionState: 'connecting' | 'connected' | 'disconnected';
}

export interface Track {
  id: string;
  title: string;
  artist: string;
  duration: number;
  bpm?: number;
  key?: string;
  waveform?: Float32Array;
  filePath?: string;
  ipfsHash?: string;
}

export interface SessionState {
  sessionId: string;
  participants: DJ[];
  currentTrack?: Track;
  bpm: number;
  key: string;
  position: number; // Position in track (seconds)
  isPlaying: boolean;
  volume: number;
  effects: Record<string, any>;
  createdAt: number;
  updatedAt: number;
}

export interface CollaborationSession {
  id: string;
  host: DJ;
  participants: DJ[];
  state: SessionState;
  createdAt: number;
  endedAt?: number;
}

export interface SessionUpdate {
  type: 'state' | 'track' | 'bpm' | 'key' | 'position' | 'play' | 'pause' | 'effect';
  sessionId: string;
  data: any;
  timestamp: number;
  userId: string;
}
