import { EventEmitter } from 'events';
import { getControlFromMIDI, isDDJRev5 } from './ddj-rev5-mapping';

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

class MIDIService extends EventEmitter {
  private devices: MIDIDevice[] = [];
  private connectedDevice: MIDIDevice | null = null;
  private isConnected: boolean = false;

  constructor() {
    super();
  }

  /**
   * List available MIDI devices
   */
  async listDevices(): Promise<MIDIDevice[]> {
    try {
      // Call Electron IPC to get MIDI devices
      if (window.electronAPI?.listMIDIDevices) {
        const devices = await window.electronAPI.listMIDIDevices();
        this.devices = devices;
        return devices;
      }
      return [];
    } catch (error) {
      console.error('Error listing MIDI devices:', error);
      return [];
    }
  }

  /**
   * Connect to a MIDI device
   */
  async connect(deviceName?: string): Promise<boolean> {
    try {
      // If device name not provided, try to find DDJ-REV5
      if (!deviceName) {
        const devices = await this.listDevices();
        const ddjRev5 = devices.find((d) => isDDJRev5(d.name));
        if (ddjRev5) {
          deviceName = ddjRev5.name;
        } else if (devices.length > 0) {
          deviceName = devices[0].name;
        } else {
          throw new Error('No MIDI devices found');
        }
      }

      // Connect via Electron IPC
      if (window.electronAPI?.connectMIDI) {
        const success = await window.electronAPI.connectMIDI(deviceName);
        if (success) {
          const devices = await this.listDevices();
          this.connectedDevice = devices.find((d) => d.name === deviceName) || null;
          this.isConnected = true;
          this.emit('connected', this.connectedDevice);
          return true;
        }
      }
      return false;
    } catch (error) {
      console.error('Error connecting to MIDI device:', error);
      this.emit('error', error);
      return false;
    }
  }

  /**
   * Disconnect from MIDI device
   */
  async disconnect(): Promise<void> {
    try {
      if (window.electronAPI?.disconnectMIDI) {
        await window.electronAPI.disconnectMIDI();
      }
      this.connectedDevice = null;
      this.isConnected = false;
      this.emit('disconnected');
    } catch (error) {
      console.error('Error disconnecting from MIDI device:', error);
    }
  }

  /**
   * Handle MIDI message from Electron
   */
  handleMIDIMessage(message: MIDIMessage): void {
    const controlName = getControlFromMIDI(
      message.type,
      message.channel,
      message.control
    );

    if (controlName) {
      this.emit('control', {
        control: controlName,
        value: message.value,
        raw: message,
      });
    } else {
      // Emit raw message for unknown controls
      this.emit('raw', message);
    }
  }

  /**
   * Send MIDI message (for feedback/leds)
   */
  async sendMessage(message: MIDIMessage): Promise<void> {
    try {
      if (window.electronAPI?.sendMIDIMessage) {
        await window.electronAPI.sendMIDIMessage(message);
      }
    } catch (error) {
      console.error('Error sending MIDI message:', error);
    }
  }

  getConnectedDevice(): MIDIDevice | null {
    return this.connectedDevice;
  }

  getIsConnected(): boolean {
    return this.isConnected;
  }
}

export const midiService = new MIDIService();

// Listen for MIDI messages from Electron
if (window.electronAPI?.onMIDIMessage) {
  window.electronAPI.onMIDIMessage((message: MIDIMessage) => {
    midiService.handleMIDIMessage(message);
  });
}
