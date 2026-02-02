/**
 * MIDI Handler for Electron Main Process
 * Handles MIDI device communication using easymidi
 */

const easymidi = require('easymidi');

interface MIDIDevice {
  name: string;
  inputIndex: number;
  outputIndex: number;
  manufacturer?: string;
}

interface MIDIMessage {
  type: 'note' | 'cc' | 'pitchbend';
  channel: number;
  control: number;
  value: number;
  timestamp: number;
}

class MIDIHandler {
  private inputs: any[] = [];
  private outputs: any[] = [];
  private connectedInput: any = null;
  private connectedOutput: any = null;
  private messageCallback: ((message: MIDIMessage) => void) | null = null;

  constructor() {
    this.refreshDevices();
  }

  /**
   * Refresh list of available MIDI devices
   */
  refreshDevices(): MIDIDevice[] {
    try {
      const inputNames = easymidi.getInputs();
      const outputNames = easymidi.getOutputs();

      const devices: MIDIDevice[] = [];

      inputNames.forEach((name: string, index: number) => {
        devices.push({
          name,
          inputIndex: index,
          outputIndex: outputNames.indexOf(name),
          manufacturer: this.extractManufacturer(name),
        });
      });

      this.inputs = inputNames;
      this.outputs = outputNames;

      return devices;
    } catch (error) {
      console.error('Error refreshing MIDI devices:', error);
      return [];
    }
  }

  /**
   * Extract manufacturer from device name
   */
  private extractManufacturer(name: string): string | undefined {
    if (name.toLowerCase().includes('pioneer')) return 'Pioneer';
    if (name.toLowerCase().includes('native instruments')) return 'Native Instruments';
    if (name.toLowerCase().includes('numark')) return 'Numark';
    return undefined;
  }

  /**
   * List available MIDI devices
   */
  listDevices(): MIDIDevice[] {
    return this.refreshDevices();
  }

  /**
   * Connect to a MIDI device
   */
  connect(deviceName: string): boolean {
    try {
      // Disconnect existing connection
      this.disconnect();

      // Find device
      const inputIndex = this.inputs.indexOf(deviceName);
      const outputIndex = this.outputs.indexOf(deviceName);

      if (inputIndex === -1) {
        console.error(`MIDI input device not found: ${deviceName}`);
        return false;
      }

      // Connect to input
      try {
        this.connectedInput = new easymidi.Input(deviceName);
        this.connectedInput.on('noteon', (msg: any) => {
          this.handleMessage('note', msg.channel, msg.note, msg.velocity);
        });
        this.connectedInput.on('noteoff', (msg: any) => {
          this.handleMessage('note', msg.channel, msg.note, 0);
        });
        this.connectedInput.on('cc', (msg: any) => {
          this.handleMessage('cc', msg.channel, msg.controller, msg.value);
        });
        this.connectedInput.on('pitch', (msg: any) => {
          this.handleMessage('pitchbend', msg.channel, 0, msg.value);
        });
      } catch (error) {
        console.error(`Error connecting to MIDI input ${deviceName}:`, error);
        return false;
      }

      // Connect to output if available
      if (outputIndex !== -1) {
        try {
          this.connectedOutput = new easymidi.Output(deviceName);
        } catch (error) {
          console.warn(`Could not connect to MIDI output ${deviceName}:`, error);
        }
      }

      console.log(`✅ Connected to MIDI device: ${deviceName}`);
      return true;
    } catch (error) {
      console.error('Error connecting to MIDI device:', error);
      return false;
    }
  }

  /**
   * Disconnect from MIDI device
   */
  disconnect(): void {
    if (this.connectedInput) {
      try {
        this.connectedInput.close();
      } catch (error) {
        console.error('Error closing MIDI input:', error);
      }
      this.connectedInput = null;
    }

    if (this.connectedOutput) {
      try {
        this.connectedOutput.close();
      } catch (error) {
        console.error('Error closing MIDI output:', error);
      }
      this.connectedOutput = null;
    }
  }

  /**
   * Handle incoming MIDI message
   */
  private handleMessage(
    type: 'note' | 'cc' | 'pitchbend',
    channel: number,
    control: number,
    value: number
  ): void {
    const message: MIDIMessage = {
      type,
      channel,
      control,
      value,
      timestamp: Date.now(),
    };

    if (this.messageCallback) {
      this.messageCallback(message);
    }
  }

  /**
   * Set callback for MIDI messages
   */
  onMessage(callback: (message: MIDIMessage) => void): void {
    this.messageCallback = callback;
  }

  /**
   * Send MIDI message
   */
  sendMessage(message: MIDIMessage): void {
    if (!this.connectedOutput) {
      console.warn('No MIDI output connected');
      return;
    }

    try {
      if (message.type === 'note') {
        if (message.value > 0) {
          this.connectedOutput.send('noteon', {
            channel: message.channel,
            note: message.control,
            velocity: message.value,
          });
        } else {
          this.connectedOutput.send('noteoff', {
            channel: message.channel,
            note: message.control,
            velocity: 0,
          });
        }
      } else if (message.type === 'cc') {
        this.connectedOutput.send('cc', {
          channel: message.channel,
          controller: message.control,
          value: message.value,
        });
      } else if (message.type === 'pitchbend') {
        this.connectedOutput.send('pitch', {
          channel: message.channel,
          value: message.value,
        });
      }
    } catch (error) {
      console.error('Error sending MIDI message:', error);
    }
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.connectedInput !== null;
  }
}

export const midiHandler = new MIDIHandler();
