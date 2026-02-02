import { useEffect, useRef } from 'react';
import { useStore } from '../store';
import { midiService } from '../services/midiService';
import { Howl } from 'howler';

/**
 * Hook to integrate MIDI controller with player
 */
export function useMIDIController(soundRef: React.MutableRefObject<Howl | null>) {
  const {
    setPlaying,
    setCurrentTime,
    currentTrack,
    tracks,
    setCurrentTrack,
  } = useStore();

  useEffect(() => {
    // Setup MIDI event handlers
    const handleControl = (data: { control: string; value: number; raw: any }) => {
      const { control, value } = data;
      const sound = soundRef.current;

      console.log(`🎛️ MIDI Control: ${control} = ${value}`);

      // Deck 1 Controls
      if (control === 'deck1_play') {
        if (sound) {
          if (value > 0) {
            if (sound.playing()) {
              sound.pause();
              setPlaying(false);
            } else {
              sound.play();
              setPlaying(true);
            }
          }
        }
      }

      if (control === 'deck1_cue') {
        if (sound && value > 0) {
          sound.seek(0);
          setCurrentTime(0);
        }
      }

      if (control === 'deck1_pitch') {
        if (sound) {
          // Pitch control: value is 0-127, map to 0.5-2.0 rate
          const rate = 0.5 + (value / 127) * 1.5;
          sound.rate(rate);
        }
      }

      if (control === 'deck1_pitch_bend_up') {
        if (sound && value > 0) {
          const currentRate = sound.rate();
          sound.rate(Math.min(2.0, currentRate + 0.01));
        }
      }

      if (control === 'deck1_pitch_bend_down') {
        if (sound && value > 0) {
          const currentRate = sound.rate();
          sound.rate(Math.max(0.5, currentRate - 0.01));
        }
      }

      if (control === 'deck1_jog_rotate') {
        if (sound) {
          // Jog wheel: value is relative, positive = forward, negative = backward
          const currentTime = sound.seek() as number;
          const newTime = currentTime + (value - 64) * 0.1; // Adjust sensitivity
          sound.seek(Math.max(0, Math.min(newTime, sound.duration())));
          setCurrentTime(sound.seek() as number);
        }
      }

      // Performance Pads - Hot Cues
      if (control.startsWith('deck1_pad_')) {
        const padNum = parseInt(control.split('_')[2]);
        if (value > 0 && currentTrack) {
          // For now, just log - can implement hot cues later
          console.log(`Hot Cue ${padNum} triggered`);
        }
      }

      // Load track (deck1_load)
      if (control === 'deck1_load' && value > 0) {
        if (tracks.length > 0) {
          // Load next track or first track
          const currentIndex = tracks.findIndex((t) => t.id === currentTrack?.id);
          const nextIndex = currentIndex >= 0 && currentIndex < tracks.length - 1
            ? currentIndex + 1
            : 0;
          setCurrentTrack(tracks[nextIndex]);
        }
      }

      // EQ Controls
      if (control === 'deck1_eq_high') {
        // Map value 0-127 to -12dB to +12dB
        const db = ((value / 127) * 24) - 12;
        console.log(`High EQ: ${db.toFixed(1)}dB`);
        // Howler doesn't support EQ directly, would need Web Audio API
      }

      if (control === 'deck1_eq_mid') {
        const db = ((value / 127) * 24) - 12;
        console.log(`Mid EQ: ${db.toFixed(1)}dB`);
      }

      if (control === 'deck1_eq_low') {
        const db = ((value / 127) * 24) - 12;
        console.log(`Low EQ: ${db.toFixed(1)}dB`);
      }

      // Filter
      if (control === 'deck1_filter') {
        // Map value 0-127 to filter frequency
        const filterValue = value / 127;
        console.log(`Filter: ${(filterValue * 100).toFixed(0)}%`);
        // Would need Web Audio API for actual filtering
      }
    };

    midiService.on('control', handleControl);

    return () => {
      midiService.off('control', handleControl);
    };
  }, [soundRef, setPlaying, setCurrentTime, currentTrack, tracks, setCurrentTrack]);
}
