package audio

type Config struct {
	// Master settings
	MasterVolume float64 // 0.0 - 1.0
	ReverbMix    float64 // 0.0 - 1.0, wet/dry mix
	ReverbTime   float64 // Reverb duration in seconds
	ReverbDecay  float64 // Reverb decay rate (0-1)

	// Synth music settings
	SynthVolume     float64 // Main synth bus volume
	SynthReverbSend float64 // Reverb send level for synth

	// Drone pad settings
	DronePadVolume  float64 // Individual drone voice volume
	DroneFilterBase float64 // Base lowpass filter frequency
	DroneFilterMod  float64 // Filter modulation range
	DroneLFORate    float64 // LFO rate for filter sweep
	DroneDetune     float64 // Max random detune (cents)
	SubBassVolume   float64 // Sub bass drone volume

	// Bass line settings
	BassVolume     float64 // Main bass volume
	BassSubVolume  float64 // Sub octave bass volume
	BassDetune     float64 // Detuned oscillator detune (cents)
	BassFilterBase float64 // Base filter cutoff
	BassFilterPeak float64 // Filter peak on note attack
	BassFilterQ    float64 // Filter resonance
	BassLFORate    float64 // Filter LFO rate
	BassGlideTime  float64 // Portamento time in seconds
	BassHoldChance float64 // Chance to hold/skip note
	BassSkipChance float64 // Chance to skip to next note

	// Arpeggio settings
	ArpAttackTime    float64 // Attack time in seconds
	ArpPeakVolume    float64 // Peak envelope volume
	ArpSustainVolume float64 // Sustain envelope volume
	ArpReleaseTime   float64 // Release time in seconds
	ArpFilterCutoff  float64 // Lowpass filter cutoff
	ArpDetune        float64 // Max random detune (fraction)
	ArpRandomChance  float64 // Chance to jump to random note
	ArpDelayMs       int     // Initial delay before starting

	// Reactive layer master settings
	ReactiveVolume     float64 // Master volume for all reactive layers
	ReactiveReverbSend float64 // Reverb send for reactive layers

	// Tension layer settings
	TensionFilterBase float64 // Base filter cutoff
	TensionFilterQ    float64 // Filter resonance
	TensionPerEnemy   float64 // Volume increase per enemy
	TensionMaxVolume  float64 // Maximum tension volume
	TensionFreqMod    float64 // Frequency modulation with threat

	// Shimmer layer settings
	ShimmerFilterFreq   float64 // Bandpass center frequency
	ShimmerFilterQ      float64 // Filter Q
	ShimmerPerEnemy     float64 // Volume per enemy
	ShimmerMaxVolume    float64 // Maximum shimmer volume
	ShimmerVibratoRate  float64 // Vibrato LFO base rate
	ShimmerVibratoDepth float64 // Vibrato depth in cents

	// Danger siren settings
	SirenThreatThreshold float64 // Threat level to activate siren
	SirenMaxVolume       float64 // Maximum siren volume
	SirenFreqMod         float64 // Frequency shift with threat
	SirenVibratoRate     float64 // Vibrato rate
	SirenVibratoDepth    float64 // Vibrato depth (Hz)
	SirenFilterCutoff    float64 // Lowpass filter cutoff
	SirenFilterQ         float64 // Filter resonance

	// Sub bass reactive settings
	SubBassReactiveMax float64 // Maximum reactive sub bass volume
	SubBassFreqMod     float64 // Frequency modulation with threat
	SubBassLFORate     float64 // LFO rate for pitch wobble
	SubBassLFODepth    float64 // LFO depth (Hz)

	// Atmospheric pad settings
	PadFilterBase  float64 // Base lowpass filter cutoff
	PadFilterQ     float64 // Filter resonance
	PadPerEnemy    float64 // Volume per enemy
	PadMaxVolume   float64 // Maximum pad volume
	PadDetuneRange float64 // Detune range for oscillator pairs

	// Pulse layer settings
	PulseFilterCutoff    float64 // Lowpass filter cutoff
	PulseLFODepth        float64 // Amplitude LFO depth
	PulseMinVolume       float64 // Volume when safe
	PulseMaxVolume       float64 // Maximum volume at high threat
	PulseThreatThreshold float64 // Threat level to activate faster pulse
}
