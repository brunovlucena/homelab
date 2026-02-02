import { create } from 'zustand';
import { v4 as uuidv4 } from 'uuid';

export interface Track {
  id: string;
  filePath: string;
  fileName: string;
  title: string;
  artist: string;
  album: string;
  duration: number;
  bpm?: number;
  key?: string;
  metadata: {
    bitrate: number;
    sampleRate: number;
    format: string;
    size: number;
  };
  lastModified: Date;
  hash: string;
}

export interface Playlist {
  id: string;
  name: string;
  tracks: string[];
  createdAt: Date;
  updatedAt: Date;
}

interface AppState {
  tracks: Track[];
  playlists: Playlist[];
  currentTrack: Track | null;
  isPlaying: boolean;
  currentTime: number;
  p2pConnected: boolean;
  p2pPeerId: string | null;
  // MIDI Controller state
  midiConnected: boolean;
  midiDeviceName: string | null;
  midiDevices: Array<{ name: string; inputIndex: number; outputIndex: number; manufacturer?: string }>;
  
  // Actions
  addTracks: (tracks: Track[]) => void;
  setCurrentTrack: (track: Track | null) => void;
  setPlaying: (playing: boolean) => void;
  setCurrentTime: (time: number) => void;
  createPlaylist: (name: string) => void;
  addTrackToPlaylist: (playlistId: string, trackId: string) => void;
  setP2PConnected: (connected: boolean) => void;
  setP2PPeerId: (peerId: string | null) => void;
  setMIDIConnected: (connected: boolean) => void;
  setMIDIDeviceName: (name: string | null) => void;
  setMIDIDevices: (devices: Array<{ name: string; inputIndex: number; outputIndex: number; manufacturer?: string }>) => void;
}

export const useStore = create<AppState>((set) => ({
  tracks: [],
  playlists: [],
  currentTrack: null,
  isPlaying: false,
  currentTime: 0,
  p2pConnected: false,
  p2pPeerId: null,
  midiConnected: false,
  midiDeviceName: null,
  midiDevices: [],

  addTracks: (tracks) =>
    set((state) => ({
      tracks: [...state.tracks, ...tracks],
    })),

  setCurrentTrack: (track) =>
    set({ currentTrack: track }),

  setPlaying: (playing) =>
    set({ isPlaying: playing }),

  setCurrentTime: (time) =>
    set({ currentTime: time }),

  createPlaylist: (name) =>
    set((state) => ({
      playlists: [
        ...state.playlists,
        {
          id: uuidv4(),
          name,
          tracks: [],
          createdAt: new Date(),
          updatedAt: new Date(),
        },
      ],
    })),

  addTrackToPlaylist: (playlistId, trackId) =>
    set((state) => ({
      playlists: state.playlists.map((playlist) =>
        playlist.id === playlistId
          ? {
              ...playlist,
              tracks: [...playlist.tracks, trackId],
              updatedAt: new Date(),
            }
          : playlist
      ),
    })),

  setP2PConnected: (connected) =>
    set({ p2pConnected: connected }),

  setP2PPeerId: (peerId) =>
    set({ p2pPeerId: peerId }),

  setMIDIConnected: (connected) =>
    set({ midiConnected: connected }),

  setMIDIDeviceName: (name) =>
    set({ midiDeviceName: name }),

  setMIDIDevices: (devices) =>
    set({ midiDevices: devices }),
}));
