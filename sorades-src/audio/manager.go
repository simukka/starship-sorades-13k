package audio

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/simukka/starship-sorades-13k/common"
)

// AudioManager manages sound effects using the Web Audio API.
type AudioManager struct {
	ctx          *js.Object
	masterGain   *js.Object
	buffers      map[int]*js.Object
	ready        bool
	canvasHeight float64
	AudioCtx     *js.Object // Exposed for state checking
	musicSource  *js.Object
	musicGain    *js.Object
	titleSource  *js.Object
	titleGain    *js.Object

	// Reverb effect chain
	reverb     *js.Object // ConvolverNode
	reverbGain *js.Object // Gain for wet signal

	// Synth music state
	synthGain     *js.Object
	synthOscs     []*js.Object // Active oscillators
	synthGains    []*js.Object // Gains for each oscillator
	synthPlaying  bool
	synthArpIndex int
	synthArpTimer *js.Object
	synthRNG      *common.SeededRNG // Seeded RNG for deterministic music

	// Enemy-reactive synth layers
	tensionOsc    *js.Object // High tension oscillator
	tensionGain   *js.Object // Tension volume
	tensionFilter *js.Object // Tension filter for sweeps
	pulseOsc      *js.Object // Rhythmic pulse oscillator
	pulseGain     *js.Object // Pulse volume
	pulseLFO      *js.Object // Pulse rate modulator

	// Blade Runner-style layers
	reactiveGain   *js.Object   // Master gain for all reactive layers
	reactiveReverb *js.Object   // Extra reverb send for reactive layers
	shimmerOscs    []*js.Object // Detuned shimmer oscillators
	shimmerGain    *js.Object   // Shimmer volume
	shimmerFilter  *js.Object   // Shimmer filter
	sirenOsc       *js.Object   // Danger siren oscillator
	sirenGain      *js.Object   // Siren volume
	sirenLFO       *js.Object   // Siren vibrato
	subBassOsc     *js.Object   // Sub bass for danger
	subBassGain    *js.Object   // Sub bass volume
	padOscs        []*js.Object // Atmospheric pad oscillators
	padGain        *js.Object   // Pad volume
	padFilter      *js.Object   // Pad filter for sweeps

	// Control panel
	controlPanel *js.Object // The control panel DOM element
	editingSfxID int        // Currently editing sound effect ID (-1 = none)

	// Level music preset
	currentPreset *LevelMusicPreset

	// Shield filter - low-pass filter for muffling external sounds when inside base
	shieldFilter     *js.Object // BiquadFilterNode for low-pass filtering
	shieldFilterGain *js.Object // Pre-filter gain for attenuation
	insideShield     bool       // Track if ship is currently inside shield
}

// GetCurrentPresetName returns the name of the currently active music preset
func (am *AudioManager) GetCurrentPresetName() string {
	if am.currentPreset == nil {
		return "None"
	}
	return am.currentPreset.Name
}

// NewAudioManager creates a new audio manager.
func NewAudioManager(seed *common.SeededRNG, height float64) *AudioManager {
	return &AudioManager{
		buffers:      make(map[int]*js.Object),
		editingSfxID: -1,
		synthRNG:     seed,
		canvasHeight: height,
	}
}

// Init initializes the Web Audio context.
func (am *AudioManager) Init() []string {
	if am.ctx != nil {
		return nil
	}

	// Try to create AudioContext
	audioCtx := js.Global.Get("AudioContext")
	if audioCtx == nil || audioCtx == js.Undefined {
		audioCtx = js.Global.Get("webkitAudioContext")
	}
	if audioCtx == nil || audioCtx == js.Undefined {
		return nil
	}

	am.ctx = audioCtx.New()
	am.AudioCtx = am.ctx // Expose for state checking
	am.masterGain = am.ctx.Call("createGain")
	am.masterGain.Call("connect", am.ctx.Get("destination"))
	am.masterGain.Get("gain").Set("value", AudioConfig.MasterVolume)

	// Initialize reverb effect chain
	am.initReverb()

	// Initialize shield filter for muffling external sounds
	am.initShieldFilter()

	am.ready = true

	sounds := make([]string, len(SfxData))

	// Load all sound effects using pure Go jsfxr implementation
	for i, sfxParams := range SfxData {
		sounds[i] = GenerateStereoWavDataURL(sfxParams, 0.0)
	}

	return sounds
}

// LoadSound loads and decodes a sound effect from a jsfxr data URL.
func (am *AudioManager) LoadSound(id int, dataURL string) {
	if am.ctx == nil {
		return
	}

	// Fetch and decode the audio data
	fetchPromise := js.Global.Call("fetch", dataURL)
	fetchPromise.Call("then", func(response *js.Object) {
		arrayBufferPromise := response.Call("arrayBuffer")
		arrayBufferPromise.Call("then", func(arrayBuffer *js.Object) {
			decodePromise := am.ctx.Call("decodeAudioData", arrayBuffer)
			decodePromise.Call("then", func(audioBuffer *js.Object) {
				am.buffers[id] = audioBuffer
			})
		})
	})
}

// Play plays a sound effect by ID.
func (am *AudioManager) Play(id int) {
	if !am.ready {
		return
	}
	buffer, ok := am.buffers[id]
	if !ok || buffer == nil {
		return
	}

	// Resume context if suspended
	if am.ctx.Get("state").String() == "suspended" {
		am.ctx.Call("resume")
	}

	// Create and play buffer source
	source := am.ctx.Call("createBufferSource")
	source.Set("buffer", buffer)
	source.Call("connect", am.masterGain)
	source.Call("start", 0)
}

// PlayWithPan plays a sound effect by ID with stereo panning and volume control.
// pan: -1.0 = full left, 0.0 = center, 1.0 = full right
// volume: 0.0 = silent, 1.0 = full volume (can exceed 1.0 for boost)
// Also sends the signal to the reverb bus at lower volume.
// Note: This routes through the shield filter for the muffling effect.
func (am *AudioManager) PlayWithPan(id int, pan, volume float64) {
	am.playWithPanInternal(id, pan, volume, true) // Route through shield filter
}

// PlayLocal plays a sound effect for the local player (not filtered by shield).
// Used for player's own actions like firing, pickups, etc.
func (am *AudioManager) PlayLocal(id int, volume float64) {
	am.playWithPanInternal(id, 0, volume, false) // Direct to master, no filter
}

// playWithPanInternal is the internal implementation for playing sounds.
// useShieldFilter: if true, routes through shield filter (for external sounds)
func (am *AudioManager) playWithPanInternal(id int, pan, volume float64, useShieldFilter bool) {
	if !am.ready {
		return
	}
	buffer, ok := am.buffers[id]
	if !ok || buffer == nil {
		return
	}

	// Resume context if suspended
	if am.ctx.Get("state").String() == "suspended" {
		am.ctx.Call("resume")
	}

	// Clamp pan to valid range
	if pan < -1 {
		pan = -1
	}
	if pan > 1 {
		pan = 1
	}

	// Clamp volume to reasonable range
	if volume < 0 {
		volume = 0
	}
	if volume > 2 {
		volume = 2
	}

	// Create stereo panner node
	panner := am.ctx.Call("createStereoPanner")
	panner.Get("pan").Set("value", pan)

	// Create gain node for volume control
	gainNode := am.ctx.Call("createGain")
	gainNode.Get("gain").Set("value", volume)

	// Create and play buffer source
	source := am.ctx.Call("createBufferSource")
	source.Set("buffer", buffer)
	source.Call("connect", gainNode)
	gainNode.Call("connect", panner)

	// Route based on whether this is an external sound
	if useShieldFilter && am.shieldFilterGain != nil {
		// External sound: route through shield filter (muffled when inside shield)
		panner.Call("connect", am.shieldFilterGain)
	} else {
		// Local/player sound: route directly to master (never muffled)
		panner.Call("connect", am.masterGain)
	}

	// Wet path: panner -> reverb (if available)
	if am.reverb != nil {
		panner.Call("connect", am.reverb)
	}

	source.Call("start", 0)
}

// PlayWithPanLoud plays a sound effect with increased volume and heavy reverb.
// Used for boss sounds that need large presence in the mix.
// pan: -1.0 = full left, 0.0 = center, 1.0 = full right
func (am *AudioManager) PlayWithPanLoud(id int, pan float64) {
	if !am.ready {
		return
	}
	buffer, ok := am.buffers[id]
	if !ok || buffer == nil {
		return
	}

	// Resume context if suspended
	if am.ctx.Get("state").String() == "suspended" {
		am.ctx.Call("resume")
	}

	// Clamp pan to valid range (reduced range for boss - more centered)
	if pan < -0.5 {
		pan = -0.5
	}
	if pan > 0.5 {
		pan = 0.5
	}

	// Create stereo panner node
	panner := am.ctx.Call("createStereoPanner")
	panner.Get("pan").Set("value", pan)

	// Boost gain for larger presence
	boostGain := am.ctx.Call("createGain")
	boostGain.Get("gain").Set("value", 2.5) // 2.5x volume boost

	// Create and play buffer source
	source := am.ctx.Call("createBufferSource")
	source.Set("buffer", buffer)
	source.Call("connect", boostGain)
	boostGain.Call("connect", panner)

	// Dry path: panner -> masterGain
	panner.Call("connect", am.masterGain)

	// Heavy wet path: extra reverb send for boss sounds
	if am.reverb != nil {
		reverbSend := am.ctx.Call("createGain")
		reverbSend.Get("gain").Set("value", 1.5) // Extra reverb
		panner.Call("connect", reverbSend)
		reverbSend.Call("connect", am.reverb)
	}

	source.Call("start", 0)
}

// initReverb creates a simple algorithmic reverb using a ConvolverNode.
func (am *AudioManager) initReverb() {
	// Create reverb gain (wet level)
	am.reverbGain = am.ctx.Call("createGain")
	am.reverbGain.Get("gain").Set("value", AudioConfig.ReverbMix)
	am.reverbGain.Call("connect", am.masterGain)

	// Create convolver for reverb
	am.reverb = am.ctx.Call("createConvolver")
	am.reverb.Call("connect", am.reverbGain)

	// Generate impulse response algorithmically
	am.generateImpulseResponse(AudioConfig.ReverbTime, AudioConfig.ReverbDecay)
}

// generateImpulseResponse creates a synthetic impulse response for reverb.
// duration: length in seconds, decay: how quickly it fades (0-1)
func (am *AudioManager) generateImpulseResponse(duration, decay float64) {
	sampleRate := am.ctx.Get("sampleRate").Int()
	length := int(float64(sampleRate) * duration)

	// Create stereo buffer
	impulse := am.ctx.Call("createBuffer", 2, length, sampleRate)

	// Fill with decaying noise
	for channel := 0; channel < 2; channel++ {
		channelData := impulse.Call("getChannelData", channel)

		for i := 0; i < length; i++ {
			// Random noise with exponential decay
			noise := js.Global.Get("Math").Call("random").Float()*2 - 1
			progress := float64(i) / float64(length)
			envelope := js.Global.Get("Math").Call("pow", 1-progress, decay).Float()
			channelData.SetIndex(i, noise*envelope)
		}
	}

	am.reverb.Set("buffer", impulse)
}

// initShieldFilter creates a low-pass filter for muffling external sounds when inside base shield.
// This simulates the shield blocking/dampening high frequencies from outside.
func (am *AudioManager) initShieldFilter() {
	// Create a biquad filter node for low-pass filtering
	am.shieldFilter = am.ctx.Call("createBiquadFilter")
	am.shieldFilter.Set("type", "lowpass")
	// Start with high cutoff (no filtering)
	am.shieldFilter.Get("frequency").Set("value", 20000)
	am.shieldFilter.Get("Q").Set("value", 0.7) // Gentle rolloff

	// Create gain node for additional attenuation
	am.shieldFilterGain = am.ctx.Call("createGain")
	am.shieldFilterGain.Get("gain").Set("value", 1.0)

	// Route: shieldFilterGain -> shieldFilter -> masterGain
	am.shieldFilterGain.Call("connect", am.shieldFilter)
	am.shieldFilter.Call("connect", am.masterGain)

	am.insideShield = false
}

// SetShieldMode enables or disables the shield audio filter effect.
// When inside=true, external sounds are muffled with a low-pass filter.
// The transition is smooth to avoid audio pops.
func (am *AudioManager) SetShieldMode(inside bool) {
	if !am.ready || am.shieldFilter == nil {
		return
	}

	// Don't retrigger if state hasn't changed
	if am.insideShield == inside {
		return
	}
	am.insideShield = inside

	currentTime := am.ctx.Get("currentTime").Float()
	transitionTime := 0.15 // Smooth 150ms transition

	if inside {
		// Inside shield - apply low-pass filter and slight volume reduction
		// Cutoff at 800Hz muffles most high frequencies while letting bass through
		am.shieldFilter.Get("frequency").Call("linearRampToValueAtTime", 800, currentTime+transitionTime)
		am.shieldFilter.Get("Q").Call("linearRampToValueAtTime", 1.5, currentTime+transitionTime) // Slight resonance at cutoff
		am.shieldFilterGain.Get("gain").Call("linearRampToValueAtTime", 0.6, currentTime+transitionTime)
	} else {
		// Outside shield - full frequency response
		am.shieldFilter.Get("frequency").Call("linearRampToValueAtTime", 20000, currentTime+transitionTime)
		am.shieldFilter.Get("Q").Call("linearRampToValueAtTime", 0.7, currentTime+transitionTime)
		am.shieldFilterGain.Get("gain").Call("linearRampToValueAtTime", 1.0, currentTime+transitionTime)
	}
}

// SetVolume sets the master volume (0.0 to 1.0).
func (am *AudioManager) SetVolume(volume float64) {
	if am.masterGain == nil {
		return
	}
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}
	am.masterGain.Get("gain").Set("value", volume)
}

// PlayTitle plays the title screen music loop.
func (am *AudioManager) PlayTitle() {
	if !am.ready {
		return
	}
	// Play intro sound
	am.Play(21)
}

// FadeOutTitle fades out the title music.
func (am *AudioManager) FadeOutTitle() {
	if !am.ready || am.titleGain == nil {
		return
	}
	// Fade out title gain
	currentTime := am.ctx.Get("currentTime").Float()
	am.titleGain.Get("gain").Call("linearRampToValueAtTime", 0, currentTime+0.5)
}

// PlayMusic starts the game music.
func (am *AudioManager) PlayMusic() {
	// Game music would be loaded separately
	// For now, this is a stub
}

// StopMusic stops the game music.
func (am *AudioManager) StopMusic() {
	if !am.ready || am.musicSource == nil {
		return
	}
	am.musicSource.Call("stop")
}

// StartSynthMusic starts the Blade Runner-style synth music.
// seed: game seed for deterministic randomness in the music
// level: game level (determines music preset/key)
func (am *AudioManager) StartSynthMusic(seed uint32, level int) {
	if !am.ready || am.synthPlaying {
		return
	}

	// Resume context if suspended
	if am.ctx.Get("state").String() == "suspended" {
		am.ctx.Call("resume")
	}

	am.synthPlaying = true

	// Get the music preset for this level
	am.currentPreset = GetLevelPreset(level)

	// Create synth master gain
	am.synthGain = am.ctx.Call("createGain")
	am.synthGain.Get("gain").Set("value", AudioConfig.SynthVolume)
	am.synthGain.Call("connect", am.masterGain)

	// Also route through reverb for atmosphere
	if am.reverb != nil {
		synthReverb := am.ctx.Call("createGain")
		synthReverb.Get("gain").Set("value", AudioConfig.SynthReverbSend)
		synthReverb.Call("connect", am.reverb)
		am.synthGain.Call("connect", synthReverb)
	}

	// Start the drone pad
	am.startDronePad()

	// Start the bass line
	am.startBassLine()

	// Start the slow arpeggio
	am.startArpeggio()

	// Start enemy-reactive layers
	am.startEnemyReactiveLayers()
}

// startDronePad creates atmospheric drone pads.
func (am *AudioManager) startDronePad() {
	// Use preset frequencies (dark atmospheric pad)
	baseFreqs := am.currentPreset.DroneFreqs

	for _, freq := range baseFreqs {
		// Main oscillator - sawtooth for rich harmonics
		osc := am.ctx.Call("createOscillator")
		osc.Set("type", "sawtooth")
		osc.Get("frequency").Set("value", freq)

		// Slight detune for thickness (seeded)
		osc.Get("detune").Set("value", (am.synthRNG.Random()-0.5)*AudioConfig.DroneDetune)

		// Low-pass filter for warmth
		filter := am.ctx.Call("createBiquadFilter")
		filter.Set("type", "lowpass")
		filter.Get("frequency").Set("value", AudioConfig.DroneFilterBase+freq*2)
		filter.Get("Q").Set("value", 1)

		// LFO for slow filter sweep (seeded rate)
		lfo := am.ctx.Call("createOscillator")
		lfo.Set("type", "sine")
		lfo.Get("frequency").Set("value", AudioConfig.DroneLFORate+am.synthRNG.Random()*AudioConfig.DroneLFORate)

		lfoGain := am.ctx.Call("createGain")
		lfoGain.Get("gain").Set("value", AudioConfig.DroneFilterMod)

		lfo.Call("connect", lfoGain)
		lfoGain.Call("connect", filter.Get("frequency"))
		lfo.Call("start")

		// Gain for this voice
		gain := am.ctx.Call("createGain")
		gain.Get("gain").Set("value", AudioConfig.DronePadVolume)

		osc.Call("connect", filter)
		filter.Call("connect", gain)
		gain.Call("connect", am.synthGain)

		osc.Call("start")

		am.synthOscs = append(am.synthOscs, osc, lfo)
		am.synthGains = append(am.synthGains, gain)
	}

	// Add sub bass drone
	subOsc := am.ctx.Call("createOscillator")
	subOsc.Set("type", "sine")
	subOsc.Get("frequency").Set("value", am.currentPreset.SubBassFreq)

	subGain := am.ctx.Call("createGain")
	subGain.Get("gain").Set("value", AudioConfig.SubBassVolume)

	subOsc.Call("connect", subGain)
	subGain.Call("connect", am.synthGain)
	subOsc.Call("start")

	am.synthOscs = append(am.synthOscs, subOsc)
	am.synthGains = append(am.synthGains, subGain)
}

// startBassLine creates a moving bass line synced to the arpeggio tempo.
func (am *AudioManager) startBassLine() {
	// Bass notes from preset (following chord progression)
	bassNotes := am.currentPreset.BassNotes
	bassTempo := am.currentPreset.BassTempo

	bassIndex := 0

	// Create main bass oscillator (always running, we modulate its frequency)
	bassOsc := am.ctx.Call("createOscillator")
	bassOsc.Set("type", "sawtooth")
	bassOsc.Get("frequency").Set("value", bassNotes[0])

	// Second oscillator for thickness (sub octave)
	bassOsc2 := am.ctx.Call("createOscillator")
	bassOsc2.Set("type", "sine")
	bassOsc2.Get("frequency").Set("value", bassNotes[0]/2) // Octave below

	// Third oscillator slightly detuned for analog warmth
	bassOsc3 := am.ctx.Call("createOscillator")
	bassOsc3.Set("type", "sawtooth")
	bassOsc3.Get("frequency").Set("value", bassNotes[0])
	bassOsc3.Get("detune").Set("value", AudioConfig.BassDetune)

	// Low-pass filter for that classic synth bass sound
	bassFilter := am.ctx.Call("createBiquadFilter")
	bassFilter.Set("type", "lowpass")
	bassFilter.Get("frequency").Set("value", AudioConfig.BassFilterBase)
	bassFilter.Get("Q").Set("value", AudioConfig.BassFilterQ)

	// Filter envelope LFO (slow sweep for movement)
	filterLFO := am.ctx.Call("createOscillator")
	filterLFO.Set("type", "sine")
	filterLFO.Get("frequency").Set("value", AudioConfig.BassLFORate)
	filterLFOGain := am.ctx.Call("createGain")
	filterLFOGain.Get("gain").Set("value", 150)
	filterLFO.Call("connect", filterLFOGain)
	filterLFOGain.Call("connect", bassFilter.Get("frequency"))
	filterLFO.Call("start")

	// Main bass gain
	bassGain := am.ctx.Call("createGain")
	bassGain.Get("gain").Set("value", AudioConfig.BassVolume)

	// Sub bass gain (quieter)
	subBassGain := am.ctx.Call("createGain")
	subBassGain.Get("gain").Set("value", AudioConfig.BassSubVolume)

	// Connect oscillators through filter
	bassOsc.Call("connect", bassFilter)
	bassOsc3.Call("connect", bassFilter)
	bassFilter.Call("connect", bassGain)
	bassGain.Call("connect", am.synthGain)

	// Sub bass direct (no filter needed for pure sine)
	bassOsc2.Call("connect", subBassGain)
	subBassGain.Call("connect", am.synthGain)

	bassOsc.Call("start")
	bassOsc2.Call("start")
	bassOsc3.Call("start")

	am.synthOscs = append(am.synthOscs, bassOsc, bassOsc2, bassOsc3, filterLFO)
	am.synthGains = append(am.synthGains, bassGain, subBassGain)

	// Bass note change timer (synced to bassTempo = 2x arpeggio rate for half-time feel)
	var playBassNote func()
	playBassNote = func() {
		if !am.synthPlaying {
			return
		}

		// Get current bass note
		freq := bassNotes[bassIndex]
		currentTime := am.ctx.Get("currentTime").Float()

		// Smooth glide to new note (portamento)
		bassOsc.Get("frequency").Call("linearRampToValueAtTime", freq, currentTime+AudioConfig.BassGlideTime)
		bassOsc2.Get("frequency").Call("linearRampToValueAtTime", freq/2, currentTime+AudioConfig.BassGlideTime)
		bassOsc3.Get("frequency").Call("linearRampToValueAtTime", freq, currentTime+AudioConfig.BassGlideTime)

		// Filter envelope - open up on note change, then close
		bassFilter.Get("frequency").Call("setValueAtTime", AudioConfig.BassFilterPeak, currentTime)
		bassFilter.Get("frequency").Call("linearRampToValueAtTime", AudioConfig.BassFilterBase-50, currentTime+0.8)

		// Advance to next note
		if am.synthRNG.Random() > AudioConfig.BassHoldChance {
			bassIndex = (bassIndex + 1) % len(bassNotes)
		} else {
			// Occasionally hold the note or skip
			if am.synthRNG.Random() > AudioConfig.BassSkipChance {
				bassIndex = (bassIndex + 2) % len(bassNotes)
			}
		}

		// Schedule next note (bassTempo +/- 12.5% variation)
		variation := float64(bassTempo) * 0.125
		nextDelay := float64(bassTempo) - variation + am.synthRNG.Random()*2*variation
		js.Global.Call("setTimeout", playBassNote, nextDelay)
	}

	// Start bass line after a delay (let drone establish first)
	js.Global.Call("setTimeout", playBassNote, 3200)
}

// startArpeggio creates a slow, haunting arpeggio.
func (am *AudioManager) startArpeggio() {
	// Arpeggio notes from preset (minor pentatonic with tension)
	arpNotes := am.currentPreset.ArpNotes
	arpTempo := am.currentPreset.ArpTempo

	am.synthArpIndex = 0

	// Play arpeggio note at preset tempo (slow, atmospheric)
	var playArpNote func()
	playArpNote = func() {
		if !am.synthPlaying {
			return
		}

		// Get current note
		freq := arpNotes[am.synthArpIndex]

		// Oscillator for arp note
		osc := am.ctx.Call("createOscillator")
		osc.Set("type", "triangle") // Softer tone

		// Slight random detune for analog feel (seeded)
		osc.Get("frequency").Set("value", freq*(1+(am.synthRNG.Random()-0.5)*AudioConfig.ArpDetune))

		// Envelope gain
		gain := am.ctx.Call("createGain")
		currentTime := am.ctx.Get("currentTime").Float()

		// ADSR-like envelope: slow attack, long decay
		gain.Get("gain").Set("value", 0)
		gain.Get("gain").Call("linearRampToValueAtTime", AudioConfig.ArpPeakVolume, currentTime+AudioConfig.ArpAttackTime)
		gain.Get("gain").Call("linearRampToValueAtTime", AudioConfig.ArpSustainVolume, currentTime+0.3)
		gain.Get("gain").Call("linearRampToValueAtTime", 0, currentTime+AudioConfig.ArpReleaseTime)

		// Filter for warmth
		filter := am.ctx.Call("createBiquadFilter")
		filter.Set("type", "lowpass")
		filter.Get("frequency").Set("value", AudioConfig.ArpFilterCutoff)

		osc.Call("connect", filter)
		filter.Call("connect", gain)
		gain.Call("connect", am.synthGain)

		osc.Call("start")
		osc.Call("stop", currentTime+AudioConfig.ArpReleaseTime+0.1)

		// Advance to next note (with seeded randomness for variation)
		if am.synthRNG.Random() > AudioConfig.ArpRandomChance {
			am.synthArpIndex = (am.synthArpIndex + 1) % len(arpNotes)
		} else {
			// Occasionally jump to a random note
			am.synthArpIndex = int(am.synthRNG.Random() * float64(len(arpNotes)))
		}

		// Schedule next note (arpTempo +/- 25% variation for organic feel)
		variation := float64(arpTempo) * 0.25
		nextDelay := float64(arpTempo) - variation + am.synthRNG.Random()*2*variation
		am.synthArpTimer = js.Global.Call("setTimeout", playArpNote, nextDelay)
	}

	// Start after a short delay
	am.synthArpTimer = js.Global.Call("setTimeout", playArpNote, AudioConfig.ArpDelayMs)
}

// startEnemyReactiveLayers creates synth layers that respond to enemy state.
// Inspired by Vangelis' Blade Runner soundtrack - dark, atmospheric, analog.
// All layers route through a master gain with heavy reverb for cohesion.
func (am *AudioManager) startEnemyReactiveLayers() {
	// === MASTER ROUTING ===
	// Create master gain for all reactive layers (lower overall volume)
	am.reactiveGain = am.ctx.Call("createGain")
	am.reactiveGain.Get("gain").Set("value", AudioConfig.ReactiveVolume)
	am.reactiveGain.Call("connect", am.synthGain)

	// Heavy reverb send for atmospheric depth
	if am.reverb != nil {
		am.reactiveReverb = am.ctx.Call("createGain")
		am.reactiveReverb.Get("gain").Set("value", AudioConfig.ReactiveReverbSend)
		am.reactiveReverb.Call("connect", am.reverb)
		am.reactiveGain.Call("connect", am.reactiveReverb)
	}

	// === TENSION LAYER ===
	// Dissonant high oscillator that increases with enemy count
	am.tensionOsc = am.ctx.Call("createOscillator")
	am.tensionOsc.Set("type", "sawtooth")
	am.tensionOsc.Get("frequency").Set("value", am.currentPreset.TensionBaseFreq)

	am.tensionFilter = am.ctx.Call("createBiquadFilter")
	am.tensionFilter.Set("type", "lowpass")
	am.tensionFilter.Get("frequency").Set("value", AudioConfig.TensionFilterBase)
	am.tensionFilter.Get("Q").Set("value", AudioConfig.TensionFilterQ)

	am.tensionGain = am.ctx.Call("createGain")
	am.tensionGain.Get("gain").Set("value", 0)

	am.tensionOsc.Call("connect", am.tensionFilter)
	am.tensionFilter.Call("connect", am.tensionGain)
	am.tensionGain.Call("connect", am.reactiveGain)
	am.tensionOsc.Call("start")
	am.synthOscs = append(am.synthOscs, am.tensionOsc)

	// === SHIMMER LAYER ===
	// CS-80 style detuned oscillators for ethereal shimmer
	am.shimmerGain = am.ctx.Call("createGain")
	am.shimmerGain.Get("gain").Set("value", 0)

	am.shimmerFilter = am.ctx.Call("createBiquadFilter")
	am.shimmerFilter.Set("type", "bandpass")
	am.shimmerFilter.Get("frequency").Set("value", AudioConfig.ShimmerFilterFreq)
	am.shimmerFilter.Get("Q").Set("value", AudioConfig.ShimmerFilterQ)

	// Create detuned oscillators for rich shimmer (from preset)
	shimmerFreqs := am.currentPreset.ShimmerFreqs
	detunes := []float64{-12, 7, -5, 15}
	for i, freq := range shimmerFreqs {
		osc := am.ctx.Call("createOscillator")
		osc.Set("type", "sine")
		osc.Get("frequency").Set("value", freq)
		osc.Get("detune").Set("value", detunes[i])

		// Slow vibrato synced to tempo
		vibrato := am.ctx.Call("createOscillator")
		vibrato.Set("type", "sine")
		vibrato.Get("frequency").Set("value", AudioConfig.ShimmerVibratoRate+am.synthRNG.Random()*AudioConfig.ShimmerVibratoRate)
		vibratoGain := am.ctx.Call("createGain")
		vibratoGain.Get("gain").Set("value", AudioConfig.ShimmerVibratoDepth+am.synthRNG.Random()*1.5)
		vibrato.Call("connect", vibratoGain)
		vibratoGain.Call("connect", osc.Get("detune"))
		vibrato.Call("start")

		osc.Call("connect", am.shimmerFilter)
		osc.Call("start")
		am.shimmerOscs = append(am.shimmerOscs, osc, vibrato)
		am.synthOscs = append(am.synthOscs, osc, vibrato)
	}
	am.shimmerFilter.Call("connect", am.shimmerGain)
	am.shimmerGain.Call("connect", am.reactiveGain)

	// === DANGER SIREN ===
	// Haunting siren that rises with extreme danger
	am.sirenOsc = am.ctx.Call("createOscillator")
	am.sirenOsc.Set("type", "sine")
	am.sirenOsc.Get("frequency").Set("value", am.currentPreset.SirenBaseFreq)

	// Slow vibrato synced to tempo subdivision
	am.sirenLFO = am.ctx.Call("createOscillator")
	am.sirenLFO.Set("type", "sine")
	am.sirenLFO.Get("frequency").Set("value", AudioConfig.SirenVibratoRate)
	sirenLFOGain := am.ctx.Call("createGain")
	sirenLFOGain.Get("gain").Set("value", AudioConfig.SirenVibratoDepth)
	am.sirenLFO.Call("connect", sirenLFOGain)
	sirenLFOGain.Call("connect", am.sirenOsc.Get("frequency"))

	sirenFilter := am.ctx.Call("createBiquadFilter")
	sirenFilter.Set("type", "lowpass")
	sirenFilter.Get("frequency").Set("value", AudioConfig.SirenFilterCutoff)
	sirenFilter.Get("Q").Set("value", AudioConfig.SirenFilterQ)

	am.sirenGain = am.ctx.Call("createGain")
	am.sirenGain.Get("gain").Set("value", 0)

	am.sirenOsc.Call("connect", sirenFilter)
	sirenFilter.Call("connect", am.sirenGain)
	am.sirenGain.Call("connect", am.reactiveGain)
	am.sirenOsc.Call("start")
	am.sirenLFO.Call("start")
	am.synthOscs = append(am.synthOscs, am.sirenOsc, am.sirenLFO)

	// === SUB BASS DRONE ===
	// Deep rumbling bass that intensifies with danger
	am.subBassOsc = am.ctx.Call("createOscillator")
	am.subBassOsc.Set("type", "sine")
	am.subBassOsc.Get("frequency").Set("value", am.currentPreset.SubBassReactiveFreq)

	// Subtle pitch wobble at tempo
	subLFO := am.ctx.Call("createOscillator")
	subLFO.Set("type", "sine")
	subLFO.Get("frequency").Set("value", AudioConfig.SubBassLFORate)
	subLFOGain := am.ctx.Call("createGain")
	subLFOGain.Get("gain").Set("value", AudioConfig.SubBassLFODepth)
	subLFO.Call("connect", subLFOGain)
	subLFOGain.Call("connect", am.subBassOsc.Get("frequency"))

	am.subBassGain = am.ctx.Call("createGain")
	am.subBassGain.Get("gain").Set("value", 0)

	am.subBassOsc.Call("connect", am.subBassGain)
	am.subBassGain.Call("connect", am.reactiveGain)
	am.subBassOsc.Call("start")
	subLFO.Call("start")
	am.synthOscs = append(am.synthOscs, am.subBassOsc, subLFO)

	// === ATMOSPHERIC PAD ===
	// Dark, evolving pad chord (from preset)
	am.padGain = am.ctx.Call("createGain")
	am.padGain.Get("gain").Set("value", 0)

	am.padFilter = am.ctx.Call("createBiquadFilter")
	am.padFilter.Set("type", "lowpass")
	am.padFilter.Get("frequency").Set("value", AudioConfig.PadFilterBase)
	am.padFilter.Get("Q").Set("value", AudioConfig.PadFilterQ)

	// Pad chord from preset
	padFreqs := am.currentPreset.PadFreqs
	for _, freq := range padFreqs {
		for _, detune := range []float64{-AudioConfig.PadDetuneRange, AudioConfig.PadDetuneRange} {
			osc := am.ctx.Call("createOscillator")
			osc.Set("type", "sawtooth")
			osc.Get("frequency").Set("value", freq)
			osc.Get("detune").Set("value", detune+am.synthRNG.Random()*3-1.5)
			osc.Call("connect", am.padFilter)
			osc.Call("start")
			am.padOscs = append(am.padOscs, osc)
			am.synthOscs = append(am.synthOscs, osc)
		}
	}
	am.padFilter.Call("connect", am.padGain)
	am.padGain.Call("connect", am.reactiveGain)

	// === PULSE LAYER ===
	// Rhythmic pulse synced to background tempo (based on arpTempo)
	am.pulseOsc = am.ctx.Call("createOscillator")
	am.pulseOsc.Set("type", "triangle")                                       // Softer than square
	am.pulseOsc.Get("frequency").Set("value", am.currentPreset.DroneFreqs[0]) // Use root note

	// Calculate tempo LFO frequency from preset (1000/arpTempo Hz)
	pulseTempoHz := 1000.0 / float64(am.currentPreset.ArpTempo)
	am.pulseLFO = am.ctx.Call("createOscillator")
	am.pulseLFO.Set("type", "sine") // Smoother pulse
	am.pulseLFO.Get("frequency").Set("value", pulseTempoHz)

	pulseLFOGain := am.ctx.Call("createGain")
	pulseLFOGain.Get("gain").Set("value", AudioConfig.PulseLFODepth)

	am.pulseGain = am.ctx.Call("createGain")
	am.pulseGain.Get("gain").Set("value", 0)

	am.pulseLFO.Call("connect", pulseLFOGain)
	pulseLFOGain.Call("connect", am.pulseGain.Get("gain"))

	pulseFilter := am.ctx.Call("createBiquadFilter")
	pulseFilter.Set("type", "lowpass")
	pulseFilter.Get("frequency").Set("value", AudioConfig.PulseFilterCutoff)

	am.pulseOsc.Call("connect", pulseFilter)
	pulseFilter.Call("connect", am.pulseGain)
	am.pulseGain.Call("connect", am.reactiveGain)

	am.pulseOsc.Call("start")
	am.pulseLFO.Call("start")
	am.synthOscs = append(am.synthOscs, am.pulseOsc, am.pulseLFO)
}

// UpdateEnemySynth updates the enemy-reactive synth layers based on enemy state.
// enemyCount: number of active enemies
// avgEnemyY: average Y position of enemies (0 = top, HEIGHT = bottom)
// closestDist: distance to closest enemy from player
func (am *AudioManager) UpdateEnemySynth(enemyCount int, avgEnemyY, closestDist float64) {
	if !am.synthPlaying || am.tensionGain == nil || am.currentPreset == nil {
		return
	}

	currentTime := am.ctx.Get("currentTime").Float()

	// Normalize values
	normY := avgEnemyY / am.canvasHeight
	threat := 1 - (closestDist / 500)
	if threat < 0 {
		threat = 0
	}
	if threat > 1 {
		threat = 1
	}

	// === TENSION LAYER ===
	tensionLevel := float64(enemyCount) * AudioConfig.TensionPerEnemy
	if tensionLevel > AudioConfig.TensionMaxVolume {
		tensionLevel = AudioConfig.TensionMaxVolume
	}
	tensionFilterFreq := AudioConfig.TensionFilterBase + (1-normY)*800
	tensionFreq := am.currentPreset.TensionBaseFreq + threat*AudioConfig.TensionFreqMod

	am.tensionGain.Get("gain").Call("linearRampToValueAtTime", tensionLevel, currentTime+0.2)
	am.tensionFilter.Get("frequency").Call("linearRampToValueAtTime", tensionFilterFreq, currentTime+0.2)
	am.tensionOsc.Get("frequency").Call("linearRampToValueAtTime", tensionFreq, currentTime+0.2)

	// === SHIMMER LAYER ===
	shimmerLevel := float64(enemyCount) * AudioConfig.ShimmerPerEnemy
	if shimmerLevel > AudioConfig.ShimmerMaxVolume {
		shimmerLevel = AudioConfig.ShimmerMaxVolume
	}
	shimmerFilterFreq := AudioConfig.ShimmerFilterFreq - 500 + (1-normY)*1200
	am.shimmerGain.Get("gain").Call("linearRampToValueAtTime", shimmerLevel, currentTime+0.3)
	am.shimmerFilter.Get("frequency").Call("linearRampToValueAtTime", shimmerFilterFreq, currentTime+0.4)

	// === DANGER SIREN ===
	sirenLevel := 0.0
	sirenFreq := am.currentPreset.SirenBaseFreq
	if threat > AudioConfig.SirenThreatThreshold {
		sirenLevel = (threat - AudioConfig.SirenThreatThreshold) * (AudioConfig.SirenMaxVolume / (1 - AudioConfig.SirenThreatThreshold))
		sirenFreq = am.currentPreset.SirenBaseFreq + (threat-AudioConfig.SirenThreatThreshold)*AudioConfig.SirenFreqMod
	}
	am.sirenGain.Get("gain").Call("linearRampToValueAtTime", sirenLevel, currentTime+0.15)
	am.sirenOsc.Get("frequency").Call("linearRampToValueAtTime", sirenFreq, currentTime+0.15)

	// === SUB BASS ===
	subLevel := threat * AudioConfig.SubBassReactiveMax
	subFreq := am.currentPreset.SubBassReactiveFreq + threat*AudioConfig.SubBassFreqMod
	am.subBassGain.Get("gain").Call("linearRampToValueAtTime", subLevel, currentTime+0.25)
	am.subBassOsc.Get("frequency").Call("linearRampToValueAtTime", subFreq, currentTime+0.25)

	// === ATMOSPHERIC PAD ===
	padLevel := float64(enemyCount) * AudioConfig.PadPerEnemy
	if padLevel > AudioConfig.PadMaxVolume {
		padLevel = AudioConfig.PadMaxVolume
	}
	padFilterFreq := AudioConfig.PadFilterBase - 100 + threat*600 + (1-normY)*300
	am.padGain.Get("gain").Call("linearRampToValueAtTime", padLevel, currentTime+0.4)
	am.padFilter.Get("frequency").Call("linearRampToValueAtTime", padFilterFreq, currentTime+0.3)

	// === PULSE LAYER ===
	basePulseHz := 1000.0 / float64(am.currentPreset.ArpTempo)
	if threat > AudioConfig.PulseThreatThreshold {
		// Tempo multipliers: 1x, 2x, 4x base tempo
		var pulseRate float64
		if threat > 0.75 {
			pulseRate = basePulseHz * 4 // 4x tempo
		} else if threat > 0.5 {
			pulseRate = basePulseHz * 2 // 2x tempo
		} else {
			pulseRate = basePulseHz // Base tempo
		}
		pulseLevel := 0.02 + threat*(AudioConfig.PulseMaxVolume-0.02)
		if pulseLevel > AudioConfig.PulseMaxVolume {
			pulseLevel = AudioConfig.PulseMaxVolume
		}
		am.pulseLFO.Get("frequency").Call("linearRampToValueAtTime", pulseRate, currentTime+0.15)
		am.pulseGain.Get("gain").Call("linearRampToValueAtTime", pulseLevel, currentTime+0.15)
	} else {
		// Very slow, quiet pulse when safe (half tempo)
		am.pulseLFO.Get("frequency").Call("linearRampToValueAtTime", basePulseHz*0.5, currentTime+0.6)
		am.pulseGain.Get("gain").Call("linearRampToValueAtTime", AudioConfig.PulseMinVolume, currentTime+0.6)
	}
}

// StopSynthMusic stops the synth music with a fade out.
func (am *AudioManager) StopSynthMusic() {
	if !am.ready || !am.synthPlaying {
		return
	}

	am.synthPlaying = false

	// Clear arp timer
	if am.synthArpTimer != nil {
		js.Global.Call("clearTimeout", am.synthArpTimer)
	}

	// Fade out and stop all oscillators
	currentTime := am.ctx.Get("currentTime").Float()

	if am.synthGain != nil {
		am.synthGain.Get("gain").Call("linearRampToValueAtTime", 0, currentTime+2)
	}

	// Stop oscillators after fade
	for _, osc := range am.synthOscs {
		if osc != nil {
			osc.Call("stop", currentTime+2.1)
		}
	}

	// Clear references
	am.synthOscs = nil
	am.synthGains = nil
}

// SetMusicPreset changes the music preset for level transitions.
// This smoothly transitions reactive layer frequencies to match the new key.
func (am *AudioManager) SetMusicPreset(level int) {
	if !am.synthPlaying || am.currentPreset == nil {
		return
	}

	newPreset := GetLevelPreset(level)
	if newPreset == am.currentPreset {
		return // No change needed
	}

	am.currentPreset = newPreset
	currentTime := am.ctx.Get("currentTime").Float()
	transitionTime := 2.0 // 2 second smooth transition

	// Update tension oscillator frequency
	if am.tensionOsc != nil {
		am.tensionOsc.Get("frequency").Call("linearRampToValueAtTime",
			newPreset.TensionBaseFreq, currentTime+transitionTime)
	}

	// Update siren frequency
	if am.sirenOsc != nil {
		am.sirenOsc.Get("frequency").Call("linearRampToValueAtTime",
			newPreset.SirenBaseFreq, currentTime+transitionTime)
	}

	// Update sub bass frequency
	if am.subBassOsc != nil {
		am.subBassOsc.Get("frequency").Call("linearRampToValueAtTime",
			newPreset.SubBassReactiveFreq, currentTime+transitionTime)
	}

	// Update shimmer oscillator frequencies
	if len(am.shimmerOscs) > 0 && len(newPreset.ShimmerFreqs) > 0 {
		for i := 0; i < len(am.shimmerOscs) && i/2 < len(newPreset.ShimmerFreqs); i += 2 {
			// shimmerOscs contains pairs: [osc, vibrato, osc, vibrato, ...]
			if am.shimmerOscs[i] != nil {
				am.shimmerOscs[i].Get("frequency").Call("linearRampToValueAtTime",
					newPreset.ShimmerFreqs[i/2], currentTime+transitionTime)
			}
		}
	}

	// Update pad oscillator frequencies
	if len(am.padOscs) > 0 && len(newPreset.PadFreqs) > 0 {
		// padOscs has 2 oscillators per frequency (detuned pair)
		for i := 0; i < len(am.padOscs) && i/2 < len(newPreset.PadFreqs); i++ {
			if am.padOscs[i] != nil {
				am.padOscs[i].Get("frequency").Call("linearRampToValueAtTime",
					newPreset.PadFreqs[i/2], currentTime+transitionTime)
			}
		}
	}

	// Update pulse oscillator to new root note
	if am.pulseOsc != nil && len(newPreset.DroneFreqs) > 0 {
		am.pulseOsc.Get("frequency").Call("linearRampToValueAtTime",
			newPreset.DroneFreqs[0], currentTime+transitionTime)
	}

	// Update pulse LFO to new tempo
	if am.pulseLFO != nil {
		newPulseHz := 1000.0 / float64(newPreset.ArpTempo)
		am.pulseLFO.Get("frequency").Call("linearRampToValueAtTime",
			newPulseHz, currentTime+transitionTime)
	}
}
