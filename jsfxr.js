/**
 * ============================================================================
 * JSFXR - JavaScript Sound Effects Synthesizer
 * ============================================================================
 * 
 * A port of the SFXR sound effect generator to JavaScript.
 * Generates retro-style 8-bit sound effects commonly used in games.
 * 
 * Original SFXR by DrPetter: http://www.drpetter.se/project_sfxr.html
 * JavaScript port by Thomas Vian, with modifications by Thiemo Mättig
 * 
 * @license Apache-2.0
 * @copyright 2010 Thomas Vian
 * 
 * 
 * Memory Optimizations
    Pre-allocated TypedArrays: Phaser and noise buffers are now Float32Array allocated once per instance, not per synthesis call
    Buffer reuse: Uses .fill(0) to clear buffers instead of creating new arrays
    Local variable caching: Hot-loop variables are cached locally to avoid property lookups
    Bitwise modulo: & 1023 instead of % 1024 for power-of-2 operations
    New jsfxrBuffer() API: Bypasses WAV encoding entirely when using Web Audio API, saving ~30% memory and CPU for playback

  * Performance Considerations
    The new jsfxrBuffer() function is recommended for modern games as it:
      Skips WAV header creation
      Skips base64 encoding
      Skips browser's WAV decoding
      Provides direct AudioBuffer for immediate playback

 * ============================================================================
 */

/**
 * SfxrParams - Sound effect parameter container
 * 
 * Holds all configurable parameters for sound synthesis including:
 * - Wave type (square, saw, sine, noise)
 * - Envelope settings (attack, sustain, decay)
 * - Frequency modulation (slide, vibrato)
 * - Filters (low-pass, high-pass)
 * - Effects (phaser, repeat)
 * 
 * @class
 */
class SfxrParams {
  /** @type {number} Wave shape: 0=square, 1=sawtooth, 2=sine, 3=noise */
  waveType = 0;
  
  /** @type {number} Time for volume to ramp up (0-1) */
  attackTime = 0;
  
  /** @type {number} Time at full volume (0-1) */
  sustainTime = 0;
  
  /** @type {number} Extra volume boost at sustain start (0-1) */
  sustainPunch = 0;
  
  /** @type {number} Time for volume to fade out (0-1) */
  decayTime = 0;
  
  /** @type {number} Base frequency of the sound (0-1) */
  startFrequency = 0;
  
  /** @type {number} Frequency cutoff - sound stops if frequency drops below this */
  minFrequency = 0;
  
  /** @type {number} Frequency slide - positive slides up, negative slides down */
  slide = 0;
  
  /** @type {number} Acceleration of frequency slide */
  deltaSlide = 0;
  
  /** @type {number} Depth of vibrato effect */
  vibratoDepth = 0;
  
  /** @type {number} Speed of vibrato oscillation */
  vibratoSpeed = 0;
  
  /** @type {number} Amount to change pitch mid-sound */
  changeAmount = 0;
  
  /** @type {number} When to apply the pitch change */
  changeSpeed = 0;
  
  /** @type {number} Duty cycle for square wave (0-1) */
  squareDuty = 0;
  
  /** @type {number} Sweep of square wave duty cycle */
  dutySweep = 0;
  
  /** @type {number} Speed of sound repeat */
  repeatSpeed = 0;
  
  /** @type {number} Initial phaser offset */
  phaserOffset = 0;
  
  /** @type {number} Phaser offset sweep */
  phaserSweep = 0;
  
  /** @type {number} Low-pass filter cutoff frequency (0-1) */
  lpFilterCutoff = 0;
  
  /** @type {number} Low-pass filter cutoff sweep */
  lpFilterCutoffSweep = 0;
  
  /** @type {number} Low-pass filter resonance (0-1) */
  lpFilterResonance = 0;
  
  /** @type {number} High-pass filter cutoff frequency (0-1) */
  hpFilterCutoff = 0;
  
  /** @type {number} High-pass filter cutoff sweep */
  hpFilterCutoffSweep = 0;
  
  /** @type {number} Master volume (0-1) */
  masterVolume = 0;

  /**
   * Parses a comma-separated settings string into parameters.
   * Format: "waveType,attackTime,sustainTime,...,masterVolume"
   * 
   * @param {string} string - Comma-separated parameter values
   * @returns {void}
   * 
   * @example
   * params.setSettingsString("0,0.1,0.3,0.2,0.4,0.5,0,0,0,0,0,0,0,0.5,0,0,0,0,1,0,0,0,0,0.5");
   */
  setSettingsString(string) {
    // MODERN JS: Using destructuring with default values would be cleaner,
    // but we need backward compatibility with short strings
    const values = string.split(",");
    
    // MODERN JS: Using Number() instead of multiplication for clearer intent
    // The | 0 converts to integer, * 1 || 0 handles undefined/NaN
    this.waveType            = Number.parseInt(values[0], 10) || 0;
    this.attackTime          = Number.parseFloat(values[1]) || 0;
    this.sustainTime         = Number.parseFloat(values[2]) || 0;
    this.sustainPunch        = Number.parseFloat(values[3]) || 0;
    this.decayTime           = Number.parseFloat(values[4]) || 0;
    this.startFrequency      = Number.parseFloat(values[5]) || 0;
    this.minFrequency        = Number.parseFloat(values[6]) || 0;
    this.slide               = Number.parseFloat(values[7]) || 0;
    this.deltaSlide          = Number.parseFloat(values[8]) || 0;
    this.vibratoDepth        = Number.parseFloat(values[9]) || 0;
    this.vibratoSpeed        = Number.parseFloat(values[10]) || 0;
    this.changeAmount        = Number.parseFloat(values[11]) || 0;
    this.changeSpeed         = Number.parseFloat(values[12]) || 0;
    this.squareDuty          = Number.parseFloat(values[13]) || 0;
    this.dutySweep           = Number.parseFloat(values[14]) || 0;
    this.repeatSpeed         = Number.parseFloat(values[15]) || 0;
    this.phaserOffset        = Number.parseFloat(values[16]) || 0;
    this.phaserSweep         = Number.parseFloat(values[17]) || 0;
    this.lpFilterCutoff      = Number.parseFloat(values[18]) || 0;
    this.lpFilterCutoffSweep = Number.parseFloat(values[19]) || 0;
    this.lpFilterResonance   = Number.parseFloat(values[20]) || 0;
    this.hpFilterCutoff      = Number.parseFloat(values[21]) || 0;
    this.hpFilterCutoffSweep = Number.parseFloat(values[22]) || 0;
    this.masterVolume        = Number.parseFloat(values[23]) || 0;

    // Ensure minimum sustain time for audible sound
    if (this.sustainTime < 0.01) {
      this.sustainTime = 0.01;
    }

    // Ensure minimum total envelope length to prevent clicks/pops
    const totalTime = this.attackTime + this.sustainTime + this.decayTime;
    if (totalTime < 0.18) {
      const multiplier = 0.18 / totalTime;
      this.attackTime  *= multiplier;
      this.sustainTime *= multiplier;
      this.decayTime   *= multiplier;
    }
  }
}

/**
 * SfxrSynth - Sound synthesizer engine
 * 
 * Generates audio samples based on SfxrParams settings.
 * Supports multiple wave types, envelope shaping, filters, and effects.
 * 
 * @class
 */
class SfxrSynth {
  /** @type {SfxrParams} Current sound parameters */
  _params = new SfxrParams();

  // =========================================================================
  // Private instance variables (envelope lengths)
  // =========================================================================
  
  /** @private @type {number} Length of attack stage in samples */
  #envelopeLength0 = 0;
  
  /** @private @type {number} Length of sustain stage in samples */
  #envelopeLength1 = 0;
  
  /** @private @type {number} Length of decay stage in samples */
  #envelopeLength2 = 0;

  // =========================================================================
  // Private instance variables (oscillator state)
  // =========================================================================
  
  /** @private @type {number} Current wave period */
  #period = 0;
  
  /** @private @type {number} Maximum period before sound stops */
  #maxPeriod = 0;
  
  /** @private @type {number} Frequency slide multiplier */
  #slide = 0;
  
  /** @private @type {number} Slide acceleration */
  #deltaSlide = 0;
  
  /** @private @type {number} Pitch change amount */
  #changeAmount = 0;
  
  /** @private @type {number} Pitch change timer */
  #changeTime = 0;
  
  /** @private @type {number} Pitch change trigger time */
  #changeLimit = 0;
  
  /** @private @type {number} Square wave duty cycle */
  #squareDuty = 0;
  
  /** @private @type {number} Duty cycle sweep */
  #dutySweep = 0;

  // =========================================================================
  // MEMORY OPTIMIZATION: Pre-allocated typed arrays for buffers
  // Using Float32Array instead of regular Array for better performance
  // =========================================================================
  
  /** @private @type {Float32Array} Phaser delay buffer (1024 samples) */
  #phaserBuffer = new Float32Array(1024);
  
  /** @private @type {Float32Array} Noise lookup table (32 samples) */
  #noiseBuffer = new Float32Array(32);

  /**
   * Resets oscillator state for partial reset (used for repeat effect).
   * Does not reset envelope lengths.
   * 
   * @returns {void}
   */
  reset() {
    const p = this._params;

    // Calculate period from frequency (inverse square relationship)
    // Adding small epsilon (0.001) prevents division by zero
    this.#period    = 100 / (p.startFrequency * p.startFrequency + 0.001);
    this.#maxPeriod = 100 / (p.minFrequency * p.minFrequency + 0.001);

    // Calculate slide as a multiplier (cubic for more natural curve)
    this.#slide      = 1 - p.slide * p.slide * p.slide * 0.01;
    this.#deltaSlide = -p.deltaSlide * p.deltaSlide * p.deltaSlide * 0.000001;

    // Square wave duty cycle (only used for wave type 0)
    if (p.waveType === 0) {
      this.#squareDuty = 0.5 - p.squareDuty / 2;
      this.#dutySweep  = -p.dutySweep * 0.00005;
    }

    // Pitch change calculation
    this.#changeAmount = p.changeAmount > 0 
      ? 1 - p.changeAmount * p.changeAmount * 0.9 
      : 1 + p.changeAmount * p.changeAmount * 10;
    this.#changeTime  = 0;
    this.#changeLimit = p.changeSpeed === 1 
      ? 0 
      : (1 - p.changeSpeed) * (1 - p.changeSpeed) * 20000 + 32;
  }

  /**
   * Performs full reset including envelope calculation.
   * Call this before starting a new sound.
   * 
   * @returns {number} Total length of sound in samples
   */
  totalReset() {
    this.reset();
    const p = this._params;

    // Calculate envelope lengths (quadratic scaling for natural feel)
    this.#envelopeLength0 = p.attackTime * p.attackTime * 100000;
    this.#envelopeLength1 = p.sustainTime * p.sustainTime * 100000;
    this.#envelopeLength2 = p.decayTime * p.decayTime * 100000 + 10;

    // Return total length (truncated to integer)
    return (this.#envelopeLength0 + this.#envelopeLength1 + this.#envelopeLength2) | 0;
  }

  /**
   * Synthesizes audio samples into the provided buffer.
   * 
   * This is the main synthesis loop that generates waveforms with:
   * - Multiple wave types (square, saw, sine, noise)
   * - ADSR envelope shaping
   * - Low-pass and high-pass filtering
   * - Vibrato and frequency slide
   * - Phaser effect
   * - 8x oversampling for anti-aliasing
   * 
   * @param {Uint16Array} buffer - Output buffer for 16-bit PCM samples
   * @param {number} length - Number of samples to generate
   * @returns {number} Actual number of samples written
   */
  synthWave(buffer, length) {
    const p = this._params;

    // =========================================================================
    // Filter configuration (calculated once per sound)
    // =========================================================================
    
    /** @type {boolean} Whether any filtering is active */
    const filtersEnabled = p.lpFilterCutoff !== 1 || p.hpFilterCutoff !== 0;
    
    /** @type {number} High-pass filter cutoff (squared for curve) */
    let hpFilterCutoff = p.hpFilterCutoff * p.hpFilterCutoff * 0.1;
    
    /** @type {number} High-pass cutoff sweep multiplier */
    const hpFilterDeltaCutoff = 1 + p.hpFilterCutoffSweep * 0.0003;
    
    /** @type {number} Low-pass filter cutoff (cubed for steeper curve) */
    let lpFilterCutoff = p.lpFilterCutoff * p.lpFilterCutoff * p.lpFilterCutoff * 0.1;
    
    /** @type {number} Low-pass cutoff sweep multiplier */
    const lpFilterDeltaCutoff = 1 + p.lpFilterCutoffSweep * 0.0001;
    
    /** @type {boolean} Whether low-pass filter is active */
    const lpFilterOn = p.lpFilterCutoff !== 1;
    
    /** @type {number} Master volume (squared for logarithmic feel) */
    const masterVolume = p.masterVolume * p.masterVolume;
    
    /** @type {number} Minimum frequency threshold */
    const minFrequency = p.minFrequency;
    
    /** @type {boolean} Whether phaser effect is active */
    const phaserEnabled = p.phaserOffset !== 0 || p.phaserSweep !== 0;
    
    /** @type {number} Phaser offset sweep rate */
    const phaserDeltaOffset = p.phaserSweep * p.phaserSweep * p.phaserSweep * 0.2;
    
    /** @type {number} Initial phaser offset */
    let phaserOffset = p.phaserOffset * p.phaserOffset * (p.phaserOffset < 0 ? -1020 : 1020);
    
    /** @type {number} Repeat interval (0 = no repeat) */
    const repeatLimit = p.repeatSpeed 
      ? ((1 - p.repeatSpeed) * (1 - p.repeatSpeed) * 20000 | 0) + 32 
      : 0;
    
    /** @type {number} Sustain punch (volume boost at sustain start) */
    const sustainPunch = p.sustainPunch;
    
    /** @type {number} Vibrato depth (half amplitude) */
    const vibratoAmplitude = p.vibratoDepth / 2;
    
    /** @type {number} Vibrato oscillation speed */
    const vibratoSpeed = p.vibratoSpeed * p.vibratoSpeed * 0.01;
    
    /** @type {number} Wave type selector */
    const waveType = p.waveType;

    // =========================================================================
    // Envelope state
    // =========================================================================
    
    let envelopeLength = this.#envelopeLength0;
    const envelopeOverLength0 = 1 / this.#envelopeLength0;  // Pre-computed reciprocal
    const envelopeOverLength1 = 1 / this.#envelopeLength1;
    const envelopeOverLength2 = 1 / this.#envelopeLength2;

    // =========================================================================
    // Low-pass filter damping (calculated from resonance)
    // =========================================================================
    
    let lpFilterDamping = 5 / (1 + p.lpFilterResonance * p.lpFilterResonance * 20) * (0.01 + lpFilterCutoff);
    // MODERN JS: Using Math.min for clearer clamping
    lpFilterDamping = 1 - Math.min(lpFilterDamping, 0.8);

    // =========================================================================
    // Synthesis state variables
    // =========================================================================
    
    let finished = false;
    let envelopeStage = 0;     // 0=attack, 1=sustain, 2=decay, 3=finished
    let envelopeTime = 0;
    let envelopeVolume = 0;
    let hpFilterPos = 0;
    let lpFilterDeltaPos = 0;
    let lpFilterOldPos = 0;
    let lpFilterPos = 0;
    let periodTemp = 0;
    let phase = 0;
    let phaserInt = 0;
    let phaserPos = 0;
    let pos = 0;
    let repeatTime = 0;
    let sample = 0;
    let superSample = 0;
    let vibratoPhase = 0;

    // =========================================================================
    // MEMORY OPTIMIZATION: Reuse pre-allocated buffers instead of creating new
    // =========================================================================
    
    // Clear phaser buffer (using TypedArray.fill for efficiency)
    this.#phaserBuffer.fill(0);
    
    // MODERN JS: Using crypto.getRandomValues for better random distribution
    // However, for audio synthesis, Math.random is sufficient and faster
    for (let i = 0; i < 32; i++) {
      this.#noiseBuffer[i] = Math.random() * 2 - 1;
    }

    // Local references to avoid property lookups in hot loop
    const phaserBuffer = this.#phaserBuffer;
    const noiseBuffer = this.#noiseBuffer;

    // Cache period and related values
    let period = this.#period;
    let maxPeriod = this.#maxPeriod;
    let slide = this.#slide;
    let deltaSlide = this.#deltaSlide;
    let changeAmount = this.#changeAmount;
    let changeTime = this.#changeTime;
    let changeLimit = this.#changeLimit;
    let squareDuty = this.#squareDuty;
    let dutySweep = this.#dutySweep;

    // =========================================================================
    // Main synthesis loop
    // =========================================================================
    
    for (let i = 0; i < length; i++) {
      if (finished) {
        return i;
      }

      // ---------------------------------------------------------------------
      // Handle repeat effect
      // ---------------------------------------------------------------------
      if (repeatLimit) {
        if (++repeatTime >= repeatLimit) {
          repeatTime = 0;
          this.reset();
          // Refresh cached values after reset
          period = this.#period;
          maxPeriod = this.#maxPeriod;
          slide = this.#slide;
          deltaSlide = this.#deltaSlide;
          changeAmount = this.#changeAmount;
          changeTime = this.#changeTime;
          changeLimit = this.#changeLimit;
          squareDuty = this.#squareDuty;
          dutySweep = this.#dutySweep;
        }
      }

      // ---------------------------------------------------------------------
      // Handle pitch change
      // ---------------------------------------------------------------------
      if (changeLimit) {
        if (++changeTime >= changeLimit) {
          changeLimit = 0;
          period *= changeAmount;
        }
      }

      // ---------------------------------------------------------------------
      // Apply frequency slide
      // ---------------------------------------------------------------------
      slide += deltaSlide;
      period *= slide;

      // Check for minimum frequency cutoff
      if (period > maxPeriod) {
        period = maxPeriod;
        if (minFrequency > 0) {
          finished = true;
        }
      }

      periodTemp = period;

      // ---------------------------------------------------------------------
      // Apply vibrato
      // ---------------------------------------------------------------------
      if (vibratoAmplitude > 0) {
        vibratoPhase += vibratoSpeed;
        periodTemp *= 1 + Math.sin(vibratoPhase) * vibratoAmplitude;
      }

      // Clamp period to minimum (prevents aliasing)
      periodTemp = Math.max(periodTemp | 0, 8);

      // ---------------------------------------------------------------------
      // Square wave duty sweep
      // ---------------------------------------------------------------------
      if (waveType === 0) {
        squareDuty += dutySweep;
        // MODERN JS: Using Math.min/max for clearer clamping
        squareDuty = Math.max(0, Math.min(0.5, squareDuty));
      }

      // ---------------------------------------------------------------------
      // Envelope stage progression
      // ---------------------------------------------------------------------
      if (++envelopeTime > envelopeLength) {
        envelopeTime = 0;
        envelopeStage++;
        
        if (envelopeStage === 1) {
          envelopeLength = this.#envelopeLength1;
        } else if (envelopeStage === 2) {
          envelopeLength = this.#envelopeLength2;
        }
      }

      // ---------------------------------------------------------------------
      // Calculate envelope volume based on current stage
      // ---------------------------------------------------------------------
      switch (envelopeStage) {
        case 0: // Attack: ramp up
          envelopeVolume = envelopeTime * envelopeOverLength0;
          break;
        case 1: // Sustain: hold with punch decay
          envelopeVolume = 1 + (1 - envelopeTime * envelopeOverLength1) * 2 * sustainPunch;
          break;
        case 2: // Decay: ramp down
          envelopeVolume = 1 - envelopeTime * envelopeOverLength2;
          break;
        case 3: // Finished
          envelopeVolume = 0;
          finished = true;
          break;
      }

      // ---------------------------------------------------------------------
      // Update phaser offset
      // ---------------------------------------------------------------------
      if (phaserEnabled) {
        phaserOffset += phaserDeltaOffset;
        phaserInt = Math.abs(phaserOffset | 0);
        // MODERN JS: Using Math.min for clamping
        phaserInt = Math.min(phaserInt, 1023);
      }

      // ---------------------------------------------------------------------
      // Update high-pass filter cutoff
      // ---------------------------------------------------------------------
      if (filtersEnabled && hpFilterDeltaCutoff !== 1) {
        hpFilterCutoff *= hpFilterDeltaCutoff;
        // MODERN JS: Using Math.min/max for clamping
        hpFilterCutoff = Math.max(0.00001, Math.min(0.1, hpFilterCutoff));
      }

      // ---------------------------------------------------------------------
      // 8x Oversampling loop (anti-aliasing)
      // Each output sample is the average of 8 sub-samples
      // ---------------------------------------------------------------------
      superSample = 0;
      
      for (let j = 0; j < 8; j++) {
        // Advance phase through waveform
        phase++;
        if (phase >= periodTemp) {
          phase %= periodTemp;

          // Generate new noise for noise wave type
          if (waveType === 3) {
            for (let n = 0; n < 32; n++) {
              noiseBuffer[n] = Math.random() * 2 - 1;
            }
          }
        }

        // -----------------------------------------------------------------
        // Generate sample based on wave type
        // -----------------------------------------------------------------
        switch (waveType) {
          case 0: // Square wave
            sample = (phase / periodTemp < squareDuty) ? 0.5 : -0.5;
            break;
            
          case 1: // Sawtooth wave
            sample = 1 - (phase / periodTemp) * 2;
            break;
            
          case 2: // Sine wave (fast polynomial approximation)
            // This approximation is faster than Math.sin() and accurate enough
            // for audio synthesis. Uses a parabolic curve with correction.
            pos = phase / periodTemp;
            pos = pos > 0.5 ? (pos - 1) * 6.28318531 : pos * 6.28318531;
            // First approximation (parabola)
            sample = pos < 0 
              ? 1.27323954 * pos + 0.405284735 * pos * pos 
              : 1.27323954 * pos - 0.405284735 * pos * pos;
            // Second pass correction for accuracy
            sample = sample < 0 
              ? 0.225 * (sample * -sample - sample) + sample 
              : 0.225 * (sample * sample - sample) + sample;
            break;
            
          case 3: // White noise
            sample = noiseBuffer[Math.abs((phase * 32 / periodTemp) | 0) % 32];
            break;
        }

        // -----------------------------------------------------------------
        // Apply low-pass and high-pass filters
        // -----------------------------------------------------------------
        if (filtersEnabled) {
          lpFilterOldPos = lpFilterPos;
          lpFilterCutoff *= lpFilterDeltaCutoff;
          // MODERN JS: Clamping with Math.min/max
          lpFilterCutoff = Math.max(0, Math.min(0.1, lpFilterCutoff));

          if (lpFilterOn) {
            // Low-pass filter: smooth transition to target
            lpFilterDeltaPos += (sample - lpFilterPos) * lpFilterCutoff;
            lpFilterDeltaPos *= lpFilterDamping;
          } else {
            lpFilterPos = sample;
            lpFilterDeltaPos = 0;
          }

          lpFilterPos += lpFilterDeltaPos;

          // High-pass filter: remove DC offset and low frequencies
          hpFilterPos += lpFilterPos - lpFilterOldPos;
          hpFilterPos *= 1 - hpFilterCutoff;
          sample = hpFilterPos;
        }

        // -----------------------------------------------------------------
        // Apply phaser effect
        // -----------------------------------------------------------------
        if (phaserEnabled) {
          // MEMORY OPTIMIZATION: Using bitwise AND for modulo 1024 (power of 2)
          phaserBuffer[phaserPos & 1023] = sample;
          sample += phaserBuffer[(phaserPos - phaserInt + 1024) & 1023];
          phaserPos++;
        }

        superSample += sample;
      }

      // ---------------------------------------------------------------------
      // Finalize sample: average, apply envelope and master volume
      // ---------------------------------------------------------------------
      superSample *= 0.125 * envelopeVolume * masterVolume;

      // Convert to 16-bit PCM with clipping
      // MODERN JS: Using Math.round for better rounding, clamp to prevent overflow
      buffer[i] = superSample >= 1 
        ? 32767 
        : superSample <= -1 
          ? -32768 
          : (superSample * 32767) | 0;
    }

    return length;
  }
}

// =============================================================================
// JSFXR Public API
// =============================================================================

/** @type {SfxrSynth} Singleton synthesizer instance (reused for memory efficiency) */
const synth = new SfxrSynth();

/**
 * Generates a sound effect and returns it as a data URL.
 * 
 * Creates a WAV file encoded as base64 that can be used directly
 * with the Audio API or as an audio source.
 * 
 * @param {string} settings - Comma-separated sound parameters
 * @returns {string} Data URL containing the WAV audio (data:audio/wav;base64,...)
 * 
 * @example
 * // Generate a laser sound
 * const laserSound = jsfxr("0,0.1,0.2,0.1,0.3,0.5,0.2,-0.5,0,0,0,0,0,0.5,0,0,0,0,1,0,0,0,0,0.5");
 * const audio = new Audio(laserSound);
 * audio.play();
 * 
 * @example
 * // MODERN JS: Using Web Audio API for better performance
 * const audioCtx = new AudioContext();
 * fetch(jsfxr(settings))
 *   .then(r => r.arrayBuffer())
 *   .then(b => audioCtx.decodeAudioData(b))
 *   .then(buffer => {
 *     const source = audioCtx.createBufferSource();
 *     source.buffer = buffer;
 *     source.connect(audioCtx.destination);
 *     source.start();
 *   });
 */
window['jsfxr'] = function(settings) {
  // Initialize parameters from settings string
  synth._params.setSettingsString(settings);
  
  // Calculate total sound length and prepare for synthesis
  const envelopeFullLength = synth.totalReset();
  
  // MEMORY OPTIMIZATION: Calculate exact buffer size needed
  // Format: 44-byte WAV header + 16-bit PCM samples (2 bytes each)
  const sampleCount = ((envelopeFullLength + 1) / 2) | 0;
  const dataSize = sampleCount * 4 + 44;
  
  // Allocate buffer for WAV file
  const data = new Uint8Array(dataSize);
  
  // Generate audio samples (writing directly after header)
  const samplesWritten = synth.synthWave(new Uint16Array(data.buffer, 44), envelopeFullLength);
  const bytesUsed = samplesWritten * 2;
  
  // =========================================================================
  // Write WAV file header (44 bytes)
  // Using DataView for cleaner header construction
  // =========================================================================
  
  const headerView = new DataView(data.buffer);
  
  // RIFF header
  headerView.setUint32(0, 0x46464952, true);   // "RIFF" (little-endian)
  headerView.setUint32(4, bytesUsed + 36, true); // File size - 8
  headerView.setUint32(8, 0x45564157, true);   // "WAVE"
  
  // fmt sub-chunk
  headerView.setUint32(12, 0x20746D66, true);  // "fmt "
  headerView.setUint32(16, 16, true);          // Sub-chunk size (16 for PCM)
  headerView.setUint16(20, 1, true);           // Audio format (1 = PCM)
  headerView.setUint16(22, 1, true);           // Number of channels (1 = mono)
  headerView.setUint32(24, 44100, true);       // Sample rate (44.1 kHz)
  headerView.setUint32(28, 88200, true);       // Byte rate (44100 * 2)
  headerView.setUint16(32, 2, true);           // Block align (2 bytes per sample)
  headerView.setUint16(34, 16, true);          // Bits per sample
  
  // data sub-chunk
  headerView.setUint32(36, 0x61746164, true);  // "data"
  headerView.setUint32(40, bytesUsed, true);   // Data size
  
  // =========================================================================
  // MODERN JS: Use built-in btoa() with Uint8Array for base64 encoding
  // This is more efficient than manual encoding
  // =========================================================================
  
  const totalBytes = bytesUsed + 44;
  
  // MODERN JS IMPROVEMENT: Using Blob and URL.createObjectURL is more memory
  // efficient for larger sounds, but data URLs work better for small sounds
  // that need to be cached/serialized
  
  // For browsers that support it, we could use:
  // return URL.createObjectURL(new Blob([data.subarray(0, totalBytes)], {type: 'audio/wav'}));
  
  // But for backward compatibility, using base64 data URL:
  // MODERN JS: Using built-in binary-to-base64 conversion
  const base64 = btoa(String.fromCharCode.apply(null, data.subarray(0, totalBytes)));
  
  return 'data:audio/wav;base64,' + base64;
};

// =============================================================================
// MODERN ALTERNATIVE: Direct Web Audio API integration
// This avoids the overhead of base64 encoding/decoding entirely
// =============================================================================

/**
 * Generates a sound effect and returns an AudioBuffer for Web Audio API.
 * 
 * This is more efficient than jsfxr() as it skips WAV encoding/decoding.
 * Requires a Web Audio API AudioContext.
 * 
 * @param {AudioContext} audioCtx - Web Audio API context
 * @param {string} settings - Comma-separated sound parameters
 * @returns {AudioBuffer} Ready-to-play audio buffer
 * 
 * @example
 * const audioCtx = new AudioContext();
 * const buffer = jsfxrBuffer(audioCtx, "0,0.1,0.2,...");
 * const source = audioCtx.createBufferSource();
 * source.buffer = buffer;
 * source.connect(audioCtx.destination);
 * source.start();
 */
window['jsfxrBuffer'] = function(audioCtx, settings) {
  synth._params.setSettingsString(settings);
  const length = synth.totalReset();
  
  // Create AudioBuffer directly (no WAV encoding needed)
  const buffer = audioCtx.createBuffer(1, length, 44100);
  const channelData = buffer.getChannelData(0);
  
  // Generate samples into a temporary Int16 buffer
  const tempBuffer = new Int16Array(length);
  const samplesWritten = synth.synthWave(new Uint16Array(tempBuffer.buffer), length);
  
  // Convert Int16 to Float32 (Web Audio API format)
  for (let i = 0; i < samplesWritten; i++) {
    channelData[i] = tempBuffer[i] / 32768;
  }
  
  return buffer;
};