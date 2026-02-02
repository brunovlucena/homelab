/**
 * DDJ-REV5 MIDI Control Mapping
 * Mapeamento dos controles MIDI da Pioneer DDJ-REV5
 * Baseado na documentação oficial da Pioneer
 */

export interface MIDIControl {
  type: 'note' | 'cc' | 'pitchbend';
  channel: number;
  control: number;
  name: string;
  description: string;
}

export interface ControlMapping {
  [key: string]: MIDIControl;
}

// Mapeamento de controles DDJ-REV5
export const DDJ_REV5_MAPPING: ControlMapping = {
  // Deck 1 Controls
  'deck1_play': { type: 'note', channel: 0, control: 0x10, name: 'Deck 1 Play', description: 'Play/Pause Deck 1' },
  'deck1_cue': { type: 'note', channel: 0, control: 0x11, name: 'Deck 1 Cue', description: 'Cue Deck 1' },
  'deck1_sync': { type: 'note', channel: 0, control: 0x12, name: 'Deck 1 Sync', description: 'Sync Deck 1' },
  'deck1_load': { type: 'note', channel: 0, control: 0x13, name: 'Deck 1 Load', description: 'Load track to Deck 1' },
  
  // Deck 1 Jog Wheel
  'deck1_jog_touch': { type: 'note', channel: 0, control: 0x20, name: 'Deck 1 Jog Touch', description: 'Jog wheel touch' },
  'deck1_jog_rotate': { type: 'cc', channel: 0, control: 0x10, name: 'Deck 1 Jog Rotate', description: 'Jog wheel rotation' },
  
  // Deck 1 Pitch
  'deck1_pitch': { type: 'cc', channel: 0, control: 0x11, name: 'Deck 1 Pitch', description: 'Pitch control' },
  'deck1_pitch_bend_up': { type: 'note', channel: 0, control: 0x14, name: 'Deck 1 Pitch Bend Up', description: 'Pitch bend up' },
  'deck1_pitch_bend_down': { type: 'note', channel: 0, control: 0x15, name: 'Deck 1 Pitch Bend Down', description: 'Pitch bend down' },
  
  // Deck 1 EQ
  'deck1_eq_high': { type: 'cc', channel: 0, control: 0x12, name: 'Deck 1 EQ High', description: 'High EQ' },
  'deck1_eq_mid': { type: 'cc', channel: 0, control: 0x13, name: 'Deck 1 EQ Mid', description: 'Mid EQ' },
  'deck1_eq_low': { type: 'cc', channel: 0, control: 0x14, name: 'Deck 1 EQ Low', description: 'Low EQ' },
  
  // Deck 1 Filter
  'deck1_filter': { type: 'cc', channel: 0, control: 0x15, name: 'Deck 1 Filter', description: 'Filter knob' },
  
  // Deck 1 Performance Pads
  'deck1_pad_1': { type: 'note', channel: 0, control: 0x30, name: 'Deck 1 Pad 1', description: 'Performance pad 1' },
  'deck1_pad_2': { type: 'note', channel: 0, control: 0x31, name: 'Deck 1 Pad 2', description: 'Performance pad 2' },
  'deck1_pad_3': { type: 'note', channel: 0, control: 0x32, name: 'Deck 1 Pad 3', description: 'Performance pad 3' },
  'deck1_pad_4': { type: 'note', channel: 0, control: 0x33, name: 'Deck 1 Pad 4', description: 'Performance pad 4' },
  'deck1_pad_5': { type: 'note', channel: 0, control: 0x34, name: 'Deck 1 Pad 5', description: 'Performance pad 5' },
  'deck1_pad_6': { type: 'note', channel: 0, control: 0x35, name: 'Deck 1 Pad 6', description: 'Performance pad 6' },
  'deck1_pad_7': { type: 'note', channel: 0, control: 0x36, name: 'Deck 1 Pad 7', description: 'Performance pad 7' },
  'deck1_pad_8': { type: 'note', channel: 0, control: 0x37, name: 'Deck 1 Pad 8', description: 'Performance pad 8' },
  
  // Deck 1 Loop Controls
  'deck1_loop_in': { type: 'note', channel: 0, control: 0x40, name: 'Deck 1 Loop In', description: 'Set loop in point' },
  'deck1_loop_out': { type: 'note', channel: 0, control: 0x41, name: 'Deck 1 Loop Out', description: 'Set loop out point' },
  'deck1_loop_exit': { type: 'note', channel: 0, control: 0x42, name: 'Deck 1 Loop Exit', description: 'Exit loop' },
  'deck1_loop_half': { type: 'note', channel: 0, control: 0x43, name: 'Deck 1 Loop Half', description: 'Half loop' },
  'deck1_loop_double': { type: 'note', channel: 0, control: 0x44, name: 'Deck 1 Loop Double', description: 'Double loop' },
  
  // Deck 2 Controls (similar, channel 1)
  'deck2_play': { type: 'note', channel: 1, control: 0x10, name: 'Deck 2 Play', description: 'Play/Pause Deck 2' },
  'deck2_cue': { type: 'note', channel: 1, control: 0x11, name: 'Deck 2 Cue', description: 'Cue Deck 2' },
  'deck2_sync': { type: 'note', channel: 1, control: 0x12, name: 'Deck 2 Sync', description: 'Sync Deck 2' },
  'deck2_load': { type: 'note', channel: 1, control: 0x13, name: 'Deck 2 Load', description: 'Load track to Deck 2' },
  
  'deck2_jog_touch': { type: 'note', channel: 1, control: 0x20, name: 'Deck 2 Jog Touch', description: 'Jog wheel touch' },
  'deck2_jog_rotate': { type: 'cc', channel: 1, control: 0x10, name: 'Deck 2 Jog Rotate', description: 'Jog wheel rotation' },
  
  'deck2_pitch': { type: 'cc', channel: 1, control: 0x11, name: 'Deck 2 Pitch', description: 'Pitch control' },
  'deck2_pitch_bend_up': { type: 'note', channel: 1, control: 0x14, name: 'Deck 2 Pitch Bend Up', description: 'Pitch bend up' },
  'deck2_pitch_bend_down': { type: 'note', channel: 1, control: 0x15, name: 'Deck 2 Pitch Bend Down', description: 'Pitch bend down' },
  
  'deck2_eq_high': { type: 'cc', channel: 1, control: 0x12, name: 'Deck 2 EQ High', description: 'High EQ' },
  'deck2_eq_mid': { type: 'cc', channel: 1, control: 0x13, name: 'Deck 2 EQ Mid', description: 'Mid EQ' },
  'deck2_eq_low': { type: 'cc', channel: 1, control: 0x14, name: 'Deck 2 EQ Low', description: 'Low EQ' },
  
  'deck2_filter': { type: 'cc', channel: 1, control: 0x15, name: 'Deck 2 Filter', description: 'Filter knob' },
  
  'deck2_pad_1': { type: 'note', channel: 1, control: 0x30, name: 'Deck 2 Pad 1', description: 'Performance pad 1' },
  'deck2_pad_2': { type: 'note', channel: 1, control: 0x31, name: 'Deck 2 Pad 2', description: 'Performance pad 2' },
  'deck2_pad_3': { type: 'note', channel: 1, control: 0x32, name: 'Deck 2 Pad 3', description: 'Performance pad 3' },
  'deck2_pad_4': { type: 'note', channel: 1, control: 0x33, name: 'Deck 2 Pad 4', description: 'Performance pad 4' },
  'deck2_pad_5': { type: 'note', channel: 1, control: 0x34, name: 'Deck 2 Pad 5', description: 'Performance pad 5' },
  'deck2_pad_6': { type: 'note', channel: 1, control: 0x35, name: 'Deck 2 Pad 6', description: 'Performance pad 6' },
  'deck2_pad_7': { type: 'note', channel: 1, control: 0x36, name: 'Deck 2 Pad 7', description: 'Performance pad 7' },
  'deck2_pad_8': { type: 'note', channel: 1, control: 0x37, name: 'Deck 2 Pad 8', description: 'Performance pad 8' },
  
  'deck2_loop_in': { type: 'note', channel: 1, control: 0x40, name: 'Deck 2 Loop In', description: 'Set loop in point' },
  'deck2_loop_out': { type: 'note', channel: 1, control: 0x41, name: 'Deck 2 Loop Out', description: 'Set loop out point' },
  'deck2_loop_exit': { type: 'note', channel: 1, control: 0x42, name: 'Deck 2 Loop Exit', description: 'Exit loop' },
  'deck2_loop_half': { type: 'note', channel: 1, control: 0x43, name: 'Deck 2 Loop Half', description: 'Half loop' },
  'deck2_loop_double': { type: 'note', channel: 1, control: 0x44, name: 'Deck 2 Loop Double', description: 'Double loop' },
  
  // Mixer Controls
  'mixer_crossfader': { type: 'cc', channel: 2, control: 0x10, name: 'Crossfader', description: 'Crossfader position' },
  'mixer_gain_1': { type: 'cc', channel: 2, control: 0x11, name: 'Gain Channel 1', description: 'Gain control channel 1' },
  'mixer_gain_2': { type: 'cc', channel: 2, control: 0x12, name: 'Gain Channel 2', description: 'Gain control channel 2' },
  'mixer_master': { type: 'cc', channel: 2, control: 0x13, name: 'Master Volume', description: 'Master volume' },
  
  // Effects
  'fx_1_on': { type: 'note', channel: 2, control: 0x50, name: 'FX 1 On', description: 'Effect 1 on/off' },
  'fx_2_on': { type: 'note', channel: 2, control: 0x51, name: 'FX 2 On', description: 'Effect 2 on/off' },
  'fx_3_on': { type: 'note', channel: 2, control: 0x52, name: 'FX 3 On', description: 'Effect 3 on/off' },
  'fx_1_param': { type: 'cc', channel: 2, control: 0x20, name: 'FX 1 Parameter', description: 'Effect 1 parameter' },
  'fx_2_param': { type: 'cc', channel: 2, control: 0x21, name: 'FX 2 Parameter', description: 'Effect 2 parameter' },
  'fx_3_param': { type: 'cc', channel: 2, control: 0x22, name: 'FX 3 Parameter', description: 'Effect 3 parameter' },
};

/**
 * Reverse mapping: MIDI message -> control name
 */
export function getControlFromMIDI(
  type: 'note' | 'cc' | 'pitchbend',
  channel: number,
  control: number
): string | null {
  for (const [name, mapping] of Object.entries(DDJ_REV5_MAPPING)) {
    if (
      mapping.type === type &&
      mapping.channel === channel &&
      mapping.control === control
    ) {
      return name;
    }
  }
  return null;
}

/**
 * Check if a device name matches DDJ-REV5
 */
export function isDDJRev5(deviceName: string): boolean {
  const name = deviceName.toLowerCase();
  return (
    name.includes('ddj-rev5') ||
    name.includes('ddj rev5') ||
    name.includes('rev5') ||
    (name.includes('pioneer') && name.includes('ddj'))
  );
}
