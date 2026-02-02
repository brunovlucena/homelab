export interface MIDIDevice {
  name: string;
  inputIndex: number;
  outputIndex: number;
  manufacturer?: string;
}

export interface MIDIMessage {
  type: 'note' | 'cc' | 'pitchbend';
  channel: number;
  control: number;
  value: number;
  timestamp: number;
}

export interface ElectronAPI {
  getVersion: () => Promise<string>;
  selectDirectory: () => Promise<string | null>;
  readFile: (path: string) => Promise<Buffer>;
  readDir: (path: string) => Promise<string[]>;
  getFileStats: (path: string) => Promise<{
    size: number;
    mtime: Date;
    isDirectory: boolean;
  }>;
  // MIDI APIs
  listMIDIDevices: () => Promise<MIDIDevice[]>;
  connectMIDI: (deviceName: string) => Promise<boolean>;
  disconnectMIDI: () => Promise<void>;
  sendMIDIMessage: (message: MIDIMessage) => Promise<void>;
  onMIDIMessage: (callback: (message: MIDIMessage) => void) => void;
}

declare global {
  interface Window {
    electronAPI?: ElectronAPI;
  }
}
