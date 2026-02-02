import { contextBridge, ipcRenderer } from 'electron';

contextBridge.exposeInMainWorld('electronAPI', {
  getVersion: () => ipcRenderer.invoke('get-app-version'),
  selectDirectory: () => ipcRenderer.invoke('select-directory'),
  readDir: (path: string) => ipcRenderer.invoke('read-dir', path),
  getFileStats: (path: string) => ipcRenderer.invoke('get-file-stats', path),
  // MIDI APIs
  listMIDIDevices: () => ipcRenderer.invoke('midi-list-devices'),
  connectMIDI: (deviceName: string) => ipcRenderer.invoke('midi-connect', deviceName),
  disconnectMIDI: () => ipcRenderer.invoke('midi-disconnect'),
  sendMIDIMessage: (message: any) => ipcRenderer.invoke('midi-send-message', message),
  onMIDIMessage: (callback: (message: any) => void) => {
    ipcRenderer.on('midi-message', (_event, message) => callback(message));
  },
});
