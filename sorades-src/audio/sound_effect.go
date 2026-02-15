package audio

import (
	"strconv"
	"strings"
)

// WaveType represents the oscillator waveform type for jsfxr
type WaveType int

const (
	WaveSquare   WaveType = 0
	WaveSawtooth WaveType = 1
	WaveSine     WaveType = 2
	WaveNoise    WaveType = 3
)

func (w WaveType) String() string {
	switch w {
	case WaveSquare:
		return "Square"
	case WaveSawtooth:
		return "Sawtooth"
	case WaveSine:
		return "Sine"
	case WaveNoise:
		return "Noise"
	default:
		return "Unknown"
	}
}

// SoundEffect represents a structured jsfxr sound effect with named parameters
type SoundEffect struct {
	ID          int    // Sound effect ID
	Name        string // Human-readable name
	Category    string // Category for grouping (Player, Enemy, Pickup, UI)
	Description string // What the sound is used for

	// Waveform
	WaveType WaveType // 0=Square, 1=Sawtooth, 2=Sine, 3=Noise

	// Envelope (0.0 - 1.0)
	AttackTime   float64 // Time to reach peak
	SustainTime  float64 // Time at peak
	SustainPunch float64 // Punch effect during sustain
	DecayTime    float64 // Time to fade out

	// Frequency
	StartFrequency float64 // Starting frequency (0.0 - 1.0, maps to Hz)
	MinFrequency   float64 // Minimum frequency cutoff
	Slide          float64 // Frequency slide (-1.0 to 1.0)
	DeltaSlide     float64 // Change in slide over time

	// Vibrato
	VibratoDepth float64 // Vibrato intensity
	VibratoSpeed float64 // Vibrato rate

	// Arpeggio
	ArpChange float64 // Arpeggio frequency change
	ArpSpeed  float64 // Arpeggio rate

	// Duty (for square wave)
	SquareDuty float64 // Square wave duty cycle
	DutySweep  float64 // Duty cycle sweep

	// Effects
	RepeatSpeed  float64 // Note repeat speed
	PhaserOffset float64 // Phaser effect offset
	PhaserSweep  float64 // Phaser sweep

	// Filters
	LPFilterCutoff      float64 // Low-pass filter cutoff
	LPFilterCutoffSweep float64 // LP filter sweep
	LPFilterResonance   float64 // LP filter resonance
	HPFilterCutoff      float64 // High-pass filter cutoff
	HPFilterCutoffSweep float64 // HP filter sweep

	// Output
	MasterVolume float64 // Overall volume (0.0 - 1.0)
}

// ToJsfxrString converts the SoundEffect back to jsfxr parameter string
func (s *SoundEffect) ToJsfxrString() string {
	formatFloat := func(f float64) string {
		if f == 0 {
			return ""
		}
		return strconv.FormatFloat(f, 'f', -1, 64)
	}

	parts := []string{
		strconv.Itoa(int(s.WaveType)),
		formatFloat(s.AttackTime),
		formatFloat(s.SustainTime),
		formatFloat(s.SustainPunch),
		formatFloat(s.DecayTime),
		formatFloat(s.StartFrequency),
		formatFloat(s.MinFrequency),
		formatFloat(s.Slide),
		formatFloat(s.DeltaSlide),
		formatFloat(s.VibratoDepth),
		formatFloat(s.VibratoSpeed),
		formatFloat(s.ArpChange),
		formatFloat(s.ArpSpeed),
		formatFloat(s.SquareDuty),
		formatFloat(s.DutySweep),
		formatFloat(s.RepeatSpeed),
		formatFloat(s.PhaserOffset),
		formatFloat(s.PhaserSweep),
		formatFloat(s.LPFilterCutoff),
		formatFloat(s.LPFilterCutoffSweep),
		formatFloat(s.LPFilterResonance),
		formatFloat(s.HPFilterCutoff),
		formatFloat(s.HPFilterCutoffSweep),
		formatFloat(s.MasterVolume),
	}
	return strings.Join(parts, ",")
}

// ParseJsfxrString parses a jsfxr parameter string into a SoundEffect
func ParseJsfxrString(id int, name, category, desc, params string) *SoundEffect {
	sfx := &SoundEffect{
		ID:          id,
		Name:        name,
		Category:    category,
		Description: desc,
	}

	// Split and parse parameters
	parts := strings.Split(params, ",")
	getFloat := func(idx int) float64 {
		if idx >= len(parts) {
			return 0
		}
		val := strings.TrimSpace(parts[idx])
		if val == "" {
			return 0
		}
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0
		}
		return f
	}

	sfx.WaveType = WaveType(int(getFloat(0)))
	sfx.AttackTime = getFloat(1)
	sfx.SustainTime = getFloat(2)
	sfx.SustainPunch = getFloat(3)
	sfx.DecayTime = getFloat(4)
	sfx.StartFrequency = getFloat(5)
	sfx.MinFrequency = getFloat(6)
	sfx.Slide = getFloat(7)
	sfx.DeltaSlide = getFloat(8)
	sfx.VibratoDepth = getFloat(9)
	sfx.VibratoSpeed = getFloat(10)
	sfx.ArpChange = getFloat(11)
	sfx.ArpSpeed = getFloat(12)
	sfx.SquareDuty = getFloat(13)
	sfx.DutySweep = getFloat(14)
	sfx.RepeatSpeed = getFloat(15)
	sfx.PhaserOffset = getFloat(16)
	sfx.PhaserSweep = getFloat(17)
	sfx.LPFilterCutoff = getFloat(18)
	sfx.LPFilterCutoffSweep = getFloat(19)
	sfx.LPFilterResonance = getFloat(20)
	sfx.HPFilterCutoff = getFloat(21)
	sfx.HPFilterCutoffSweep = getFloat(22)
	sfx.MasterVolume = getFloat(23)

	return sfx
}

// GetSoundEffect returns a sound effect by ID
func GetSoundEffect(id int) *SoundEffect {
	if id >= 0 && id < len(SoundEffectLibrary) {
		return SoundEffectLibrary[id]
	}
	return nil
}

// GetSoundEffectsByCategory returns all sound effects in a category
func GetSoundEffectsByCategory(category string) []*SoundEffect {
	var result []*SoundEffect
	for _, sfx := range SoundEffectLibrary {
		if sfx.Category == category {
			result = append(result, sfx)
		}
	}
	return result
}
